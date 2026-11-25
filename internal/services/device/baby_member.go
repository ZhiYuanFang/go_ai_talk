package device

import (
	"context"
	"errors"
	"strings"
	"time"

	"hello/internal/dao"

	"github.com/gogf/gf/v2/frame/g"
)

var (
	ErrBabyMemberAlreadyBound   = errors.New("账号已绑定宝宝，请使用新客户端添加更多宝宝")
	ErrBabyMemberNotFound      = errors.New("未找到该宝宝成员关系")
	ErrBabyMemberNotOwned      = errors.New("宝宝不在账号成员列表中")
	ErrBabyUnbindNotAllowed    = errors.New("客户端不允许解绑宝宝")
	ErrWxRoleChangeForbidden   = errors.New("不允许变更账号 role")
	ErrWxRoleStaffOnly         = errors.New("仅 staff 子账户可执行该操作")
	ErrWxRoleParentOnly        = errors.New("仅 parent 可执行该操作")
	ErrMerchantBabyQuota       = errors.New("免费额度已满，请先开通商户订阅")
	ErrMerchantBabyNotOwned    = errors.New("宝宝不属于该商户")
	ErrAssignStaffRoleRequired = errors.New("只能分发给 role=staff 的子账户")
)

// BabyMemberItem 成员宝宝摘要。
type BabyMemberItem struct {
	DeviceNo string `json:"deviceNo"`
	BabyName string `json:"babyName"`
	Source   string `json:"source"`
	IsFocus  bool   `json:"isFocus"`
}

func wxRole(ctx context.Context, wxID int64) (string, error) {
	one, err := g.DB().Model("wx").Ctx(ctx).Fields("role").Where("id", wxID).Limit(1).One()
	if err != nil {
		return "", err
	}
	if one.IsEmpty() {
		return "", ErrWxIDInvalid
	}
	role := strings.TrimSpace(one["role"].String())
	if role == "" {
		return WxRoleParent, nil
	}
	return role, nil
}

// ListBabyMembers 列出 wx 全部成员宝宝（含画像名与是否焦点）。
func ListBabyMembers(ctx context.Context, wxID int64) ([]BabyMemberItem, error) {
	if wxID <= 0 {
		return nil, ErrWxIDInvalid
	}
	focus, _ := WxDeviceNoByWxID(ctx, wxID)
	rows, err := g.DB().Model("wx_baby_member").Ctx(ctx).
		Where("wx_id", wxID).OrderAsc("id").All()
	if err != nil {
		return nil, err
	}
	out := make([]BabyMemberItem, 0, len(rows))
	for _, r := range rows {
		dn := strings.TrimSpace(r["device_no"].String())
		name := ""
		if p, e := DeviceProfile().GetProfile(ctx, dn); e == nil {
			name = strings.TrimSpace(p.BabyName)
		}
		out = append(out, BabyMemberItem{
			DeviceNo: dn,
			BabyName: name,
			Source:   r["source"].String(),
			IsFocus:  dn != "" && dn == focus,
		})
	}
	return out, nil
}

func countBabyMembers(ctx context.Context, wxID int64) (int, error) {
	n, err := g.DB().Model("wx_baby_member").Ctx(ctx).Where("wx_id", wxID).Count()
	return n, err
}

func hasBabyMember(ctx context.Context, wxID int64, deviceNo string) (bool, error) {
	n, err := g.DB().Model("wx_baby_member").Ctx(ctx).
		Where("wx_id", wxID).Where("device_no", deviceNo).Count()
	return n > 0, err
}

// AddBabyMember 写入成员边（幂等）；不自动改焦点 unless setFocus。
func AddBabyMember(ctx context.Context, wxID int64, deviceNo, source string, setFocus bool) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if wxID <= 0 || deviceNo == "" {
		return errors.New("wxId 或 deviceNo 不能为空")
	}
	if strings.TrimSpace(source) == "" {
		source = BabyMemberSourceSelf
	}
	now := time.Now().Unix()
	_, err := g.DB().Model("wx_baby_member").Ctx(ctx).Data(g.Map{
		"wx_id":      wxID,
		"device_no":  deviceNo,
		"source":     source,
		"created_at": now,
	}).InsertIgnore()
	if err != nil {
		return err
	}
	if setFocus {
		return SetFocusBaby(ctx, wxID, deviceNo)
	}
	return nil
}

// SetFocusBaby 校验成员后更新 wx.device_no 焦点镜像。
func SetFocusBaby(ctx context.Context, wxID int64, deviceNo string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	ok, err := hasBabyMember(ctx, wxID, deviceNo)
	if err != nil {
		return err
	}
	if !ok {
		return ErrBabyMemberNotOwned
	}
	_, err = dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Id, wxID).Data(g.Map{
		dao.Wx.Columns().DeviceNo: deviceNo,
	}).Update()
	if err != nil {
		return err
	}
	row, _ := wxRowByWxID(ctx, wxID)
	if row != nil {
		_ = invalidateWxCaches(ctx, row.Id, strings.TrimSpace(row.Unionid))
	}
	return nil
}

// AppUnbindBabyRejected 客户端解绑一律拒绝。
func AppUnbindBabyRejected(ctx context.Context, wxID int64, deviceNo string) error {
	_ = ctx
	_ = wxID
	_ = deviceNo
	return ErrBabyUnbindNotAllowed
}

// CreateVirtualBabyForParent parent 创建虚拟宝宝并加入成员；默认切到新宝宝为焦点。
func CreateVirtualBabyForParent(ctx context.Context, wxID int64, babyName string, birthdayUnixSec int64, sex int) (string, error) {
	role, err := wxRole(ctx, wxID)
	if err != nil {
		return "", err
	}
	if role != WxRoleParent {
		return "", ErrWxRoleParentOnly
	}
	dn, err := generateUniqueDeviceNo(ctx)
	if err != nil {
		return "", err
	}
	if _, err := DeviceAdmin().Register(ctx, dn); err != nil {
		return "", err
	}
	if err := DeviceAdmin().SaveUserProfile(ctx, dn, strings.TrimSpace(babyName), birthdayUnixSec, sex); err != nil {
		return "", err
	}
	if err := AddBabyMember(ctx, wxID, dn, BabyMemberSourceSelf, true); err != nil {
		return "", err
	}
	return dn, nil
}
