package device

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"golang.org/x/crypto/bcrypt"
)

var (
	usernamePattern = regexp.MustCompile(`^[a-z0-9_]{4,32}$`)

	// ErrWxIDInvalid 账号主键非法。
	ErrWxIDInvalid = errors.New("wxId 无效")
	// ErrWxUsernameInvalid 用户名不符合规则（仅允许 a-z0-9_，长度 4-32）。
	ErrWxUsernameInvalid = errors.New("用户名格式无效，仅支持 4-32 位小写字母、数字、下划线")
	// ErrWxPasswordInvalid 密码不符合规则。
	ErrWxPasswordInvalid = errors.New("密码格式无效，请输入 6-64 位字符")
	// ErrWxUsernameTaken 用户名已被占用。
	ErrWxUsernameTaken = errors.New("用户名已存在")
	// ErrWxUsernameAuthFailed 用户名认证失败（防枚举统一文案）。
	ErrWxUsernameAuthFailed = errors.New("用户名或密码错误")
	// ErrWxUsernameNotSet 账号未设置用户名密码。
	ErrWxUsernameNotSet = errors.New("当前账号未设置用户名密码")
	// ErrWxUsernameAlreadySet 当前账号已设置用户名。
	ErrWxUsernameAlreadySet = errors.New("账号已存在用户名密码")
	// ErrWxAlreadyBoundUnionID 当前账号已绑定微信。
	ErrWxAlreadyBoundUnionID = errors.New("当前账号已绑定微信")
	// ErrWxUnionIDTakenByOther 目标微信已绑定其他账号。
	ErrWxUnionIDTakenByOther = errors.New("微信已绑定其他账号")
	// ErrWxUnionIDRequired 微信账号下创建用户名必须已有微信绑定。
	ErrWxUnionIDRequired = errors.New("当前账号未绑定微信")
)

func normalizeUserName(raw string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if !usernamePattern.MatchString(normalized) {
		return "", ErrWxUsernameInvalid
	}
	return normalized, nil
}

func validatePassword(raw string) (string, error) {
	password := strings.TrimSpace(raw)
	if l := len(password); l < 6 || l > 64 {
		return "", ErrWxPasswordInvalid
	}
	return password, nil
}

func hashPassword(raw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func verifyPassword(hashText, raw string) bool {
	if strings.TrimSpace(hashText) == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hashText), []byte(raw)) == nil
}

func wxRowByUserName(ctx context.Context, userName string) (*entity.Wx, error) {
	var row entity.Wx
	if err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().UserName, userName).Scan(&row); err != nil {
		return nil, err
	}
	if row.Id == 0 {
		return nil, nil
	}
	return &row, nil
}

// WxUsernameRegister 基于 wx 表注册用户名密码账号（unionid 可为空）。
func WxUsernameRegister(ctx context.Context, userName, password string) (int64, error) {
	normalized, err := normalizeUserName(userName)
	if err != nil {
		return 0, err
	}
	password, err = validatePassword(password)
	if err != nil {
		return 0, err
	}
	if row, err := wxRowByUserName(ctx, normalized); err != nil {
		return 0, err
	} else if row != nil {
		return 0, ErrWxUsernameTaken
	}
	hash, err := hashPassword(password)
	if err != nil {
		return 0, err
	}
	res, err := dao.Wx.Ctx(ctx).Data(g.Map{
		dao.Wx.Columns().UserName: normalized,
		dao.Wx.Columns().Password: hash,
	}).Insert()
	if err != nil {
		if row, e2 := wxRowByUserName(ctx, normalized); e2 == nil && row != nil {
			return 0, ErrWxUsernameTaken
		}
		return 0, err
	}
	newID, _ := res.LastInsertId()
	_ = invalidateWxCaches(ctx, newID, "")
	glog.Infof(ctx, "[wx-username] register success wxId=%d userName=%s", newID, normalized)
	return newID, nil
}

// WxUsernameLogin 用户名密码登录（仅业务，不签发 JWT）。
func WxUsernameLogin(ctx context.Context, userName, password string) (*WxLoginResult, error) {
	normalized, err := normalizeUserName(userName)
	if err != nil {
		return nil, err
	}
	password, err = validatePassword(password)
	if err != nil {
		return nil, err
	}
	row, err := wxRowByUserName(ctx, normalized)
	if err != nil {
		return nil, err
	}
	if row == nil || !verifyPassword(row.Password, password) {
		return nil, ErrWxUsernameAuthFailed
	}
	_ = invalidateWxCaches(ctx, row.Id, strings.TrimSpace(row.Unionid))
	return &WxLoginResult{WxId: row.Id, DeviceNo: strings.TrimSpace(row.DeviceNo), IsNewUser: false}, nil
}

// WxUsernameBindWxByCode 用户名账号绑定微信（通过微信授权 code 换 unionid）。
func WxUsernameBindWxByCode(ctx context.Context, wxID int64, jsCode, platform string) error {
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
	if strings.TrimSpace(row.UserName) == "" || strings.TrimSpace(row.Password) == "" {
		return ErrWxUsernameNotSet
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
	if one, err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Unionid, unionID).One(); err != nil {
		return err
	} else if !one.IsEmpty() && one[dao.Wx.Columns().Id].Int64() != wxID {
		return ErrWxUnionIDTakenByOther
	}
	_, err = dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Id, wxID).Data(g.Map{
		dao.Wx.Columns().Unionid:  unionID,
		dao.Wx.Columns().Platform: strings.TrimSpace(platform),
	}).Update()
	if err != nil {
		return err
	}
	_ = invalidateWxCaches(ctx, wxID, unionID)
	glog.Infof(ctx, "[wx-username] bind wx success wxId=%d", wxID)
	return nil
}

// WxUsernameBindDevice 用户名账号绑定设备号。
func WxUsernameBindDevice(ctx context.Context, wxID int64, deviceNo string) error {
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
	if strings.TrimSpace(row.UserName) == "" || strings.TrimSpace(row.Password) == "" {
		return ErrWxUsernameNotSet
	}
	if err = WxBindDevice(ctx, wxID, deviceNo); err != nil {
		return err
	}
	glog.Infof(ctx, "[wx-username] bind device success wxId=%d deviceNo=%s", wxID, strings.TrimSpace(deviceNo))
	return nil
}

// WxUsernameChangePassword 修改用户名账号密码。
func WxUsernameChangePassword(ctx context.Context, wxID int64, oldPassword, newPassword string) error {
	if wxID <= 0 {
		return ErrWxIDInvalid
	}
	oldPassword, err := validatePassword(oldPassword)
	if err != nil {
		return err
	}
	newPassword, err = validatePassword(newPassword)
	if err != nil {
		return err
	}
	row, err := wxRowByWxID(ctx, wxID)
	if err != nil {
		return err
	}
	if row == nil {
		return ErrWxIDInvalid
	}
	if strings.TrimSpace(row.UserName) == "" || strings.TrimSpace(row.Password) == "" {
		return ErrWxUsernameNotSet
	}
	if !verifyPassword(row.Password, oldPassword) {
		return ErrWxUsernameAuthFailed
	}
	hash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	_, err = dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Id, wxID).Data(g.Map{dao.Wx.Columns().Password: hash}).Update()
	if err != nil {
		return err
	}
	glog.Infof(ctx, "[wx-username] change password success wxId=%d", wxID)
	return nil
}

// WxCreateUsernamePassword 微信账号下创建用户名密码。
func WxCreateUsernamePassword(ctx context.Context, wxID int64, userName, password string) error {
	if wxID <= 0 {
		return ErrWxIDInvalid
	}
	normalized, err := normalizeUserName(userName)
	if err != nil {
		return err
	}
	password, err = validatePassword(password)
	if err != nil {
		return err
	}
	row, err := wxRowByWxID(ctx, wxID)
	if err != nil {
		return err
	}
	if row == nil {
		return ErrWxIDInvalid
	}
	if strings.TrimSpace(row.Unionid) == "" {
		return ErrWxUnionIDRequired
	}
	if strings.TrimSpace(row.UserName) != "" {
		return ErrWxUsernameAlreadySet
	}
	if existed, err := wxRowByUserName(ctx, normalized); err != nil {
		return err
	} else if existed != nil {
		return ErrWxUsernameTaken
	}
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	_, err = dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Id, wxID).Data(g.Map{
		dao.Wx.Columns().UserName: normalized,
		dao.Wx.Columns().Password: hash,
	}).Update()
	if err != nil {
		if existed, e2 := wxRowByUserName(ctx, normalized); e2 == nil && existed != nil && existed.Id != wxID {
			return ErrWxUsernameTaken
		}
		return err
	}
	_ = invalidateWxCaches(ctx, wxID, strings.TrimSpace(row.Unionid))
	glog.Infof(ctx, "[wx-username] create username success wxId=%d userName=%s", wxID, normalized)
	return nil
}
