package device

import (
	"context"
	"errors"
	"strings"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

var (
	// ErrAppleSubTakenByOther Apple sub 已绑定其他 wx 行。
	ErrAppleSubTakenByOther = errors.New("Apple 账号已绑定其他用户")
	// ErrAppleAlreadyBound 当前 wx 行已绑定 Apple。
	ErrAppleAlreadyBound = errors.New("当前账号已绑定 Apple")
	// ErrAccountMergeConflict 不可合并两条已独立存在的完整账号。
	ErrAccountMergeConflict = errors.New("无法合并两个已独立创建的账号")
)

// WxAppleLogin 校验 Apple identityToken，按 apple_sub 查找或创建 wx 行（不存储 email/name）。
func WxAppleLogin(ctx context.Context, identityToken, platform string) (*WxLoginResult, error) {
	sub, err := verifyAppleIdentityToken(ctx, identityToken)
	if err != nil {
		return nil, err
	}
	platform = strings.TrimSpace(platform)

	one, err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().AppleSub, sub).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		res, insErr := dao.Wx.Ctx(ctx).Data(g.Map{
			dao.Wx.Columns().AppleSub: sub,
			dao.Wx.Columns().Platform: platform,
		}).Insert()
		if insErr != nil {
			return nil, insErr
		}
		newID, _ := res.LastInsertId()
		_ = invalidateWxCaches(ctx, newID, "")
		return &WxLoginResult{WxId: newID, DeviceNo: "", IsNewUser: true}, nil
	}
	var row entity.Wx
	if err := one.Struct(&row); err != nil {
		return nil, err
	}
	_ = invalidateWxCaches(ctx, row.Id, strings.TrimSpace(row.Unionid))
	return &WxLoginResult{
		WxId:      row.Id,
		DeviceNo:  strings.TrimSpace(row.DeviceNo),
		IsNewUser: false,
	}, nil
}

// WxBindApple 将 Apple 身份绑定到当前 wx 行（Bearer 会话 wxId）。
func WxBindApple(ctx context.Context, wxID int64, identityToken, platform string) error {
	if wxID <= 0 {
		return ErrWxIDInvalid
	}
	row, err := wxRowByWxID(ctx, wxID)
	if err != nil {
		return err
	}
	if row == nil {
		return ErrWxIDInvalid
	}
	sub, err := verifyAppleIdentityToken(ctx, identityToken)
	if err != nil {
		return err
	}
	currentSub := strings.TrimSpace(row.AppleSub)
	if currentSub != "" {
		if currentSub == sub {
			return nil
		}
		return ErrAppleAlreadyBound
	}
	if conflictErr := checkAppleSubConflict(ctx, wxID, sub); conflictErr != nil {
		return conflictErr
	}
	_, err = dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Id, wxID).Data(g.Map{
		dao.Wx.Columns().AppleSub: sub,
		dao.Wx.Columns().Platform: strings.TrimSpace(platform),
	}).Update()
	if err != nil {
		return err
	}
	_ = invalidateWxCaches(ctx, wxID, strings.TrimSpace(row.Unionid))
	glog.Infof(ctx, "[apple-bind] success wxId=%d", wxID)
	return nil
}

// WxBindWxByCode 将微信 unionid 绑定到当前 wx 行（不限于用户名账号）。
func WxBindWxByCode(ctx context.Context, wxID int64, jsCode, platform string) error {
	if wxID <= 0 {
		return ErrWxIDInvalid
	}
	row, err := wxRowByWxID(ctx, wxID)
	if err != nil {
		return err
	}
	if row == nil {
		return ErrWxIDInvalid
	}
	if strings.TrimSpace(row.Unionid) != "" {
		return ErrWxAlreadyBoundUnionID
	}
	oauth, err := exchangeAuthCodeForUnionID(ctx, platform, jsCode)
	if err != nil {
		return err
	}
	unionID := strings.TrimSpace(oauth.UnionID)
	if unionID == "" {
		return ErrWxUnionIDRequired
	}
	if conflictErr := checkUnionIDConflict(ctx, wxID, unionID); conflictErr != nil {
		return conflictErr
	}
	_, err = dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Id, wxID).Data(g.Map{
		dao.Wx.Columns().Unionid:  unionID,
		dao.Wx.Columns().Platform: strings.TrimSpace(platform),
	}).Update()
	if err != nil {
		return err
	}
	_ = invalidateWxCaches(ctx, wxID, unionID)
	glog.Infof(ctx, "[wx-bind] success wxId=%d", wxID)
	return nil
}

func checkAppleSubConflict(ctx context.Context, currentWxID int64, sub string) error {
	one, err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().AppleSub, sub).One()
	if err != nil {
		return err
	}
	if one.IsEmpty() {
		return nil
	}
	otherID := one[dao.Wx.Columns().Id].Int64()
	if otherID == currentWxID {
		return nil
	}
	var other entity.Wx
	if err := one.Struct(&other); err != nil {
		return err
	}
	if isCompleteWxAccount(&other) {
		return ErrAccountMergeConflict
	}
	return ErrAppleSubTakenByOther
}

func checkUnionIDConflict(ctx context.Context, currentWxID int64, unionID string) error {
	one, err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Unionid, unionID).One()
	if err != nil {
		return err
	}
	if one.IsEmpty() {
		return nil
	}
	otherID := one[dao.Wx.Columns().Id].Int64()
	if otherID == currentWxID {
		return nil
	}
	var other entity.Wx
	if err := one.Struct(&other); err != nil {
		return err
	}
	if isCompleteWxAccount(&other) {
		return ErrAccountMergeConflict
	}
	return ErrWxUnionIDTakenByOther
}

// isCompleteWxAccount 两行各含独立完整身份（Apple + 微信均已绑定）时不可合并。
func isCompleteWxAccount(row *entity.Wx) bool {
	if row == nil {
		return false
	}
	return strings.TrimSpace(row.AppleSub) != "" && strings.TrimSpace(row.Unionid) != ""
}

func deriveAuthProviders(row *entity.Wx) []string {
	if row == nil {
		return nil
	}
	var providers []string
	if strings.TrimSpace(row.AppleSub) != "" {
		providers = append(providers, "apple")
	}
	if strings.TrimSpace(row.Unionid) != "" {
		providers = append(providers, "wechat")
	}
	if strings.TrimSpace(row.Account) != "" && strings.TrimSpace(row.Password) != "" {
		providers = append(providers, "username")
	}
	return providers
}
