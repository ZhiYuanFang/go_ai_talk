package device

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	headerInternalWxId       = "X-Internal-Wx-Id"
	envGatewayInternalSecret = "DEVICE_GATEWAY_INTERNAL_SECRET"
	cacheTTLWxLookup         = 120 * time.Second
	maxRandomDeviceAttempts  = 15
)

var wxLookupCache = cachekit.Default()

// WxLoginResult device 侧微信登录业务返回（不含 JWT；不向客户端回传 unionid）。
type WxLoginResult struct {
	WxId      int64  `json:"wxId"`
	DeviceNo  string `json:"deviceNo"`
	IsNewUser bool   `json:"isNewUser"`
}

// ErrWxDeviceLoginRejected 设备号登录被拒绝：仅当设备号未在 user 表注册时使用（与「无 wx 绑定」无关：无绑定时 wxId 返回 0）。
var ErrWxDeviceLoginRejected = errors.New("设备登录失败，请确认设备已注册")

// ErrWxDeviceLoginDeviceNoEmpty 入参 device_no 为空（trim 后）。
var ErrWxDeviceLoginDeviceNoEmpty = errors.New("deviceNo 不能为空")

// ErrWxDeactivateWxIDInvalid 注销入参 wxId 非法。
var ErrWxDeactivateWxIDInvalid = errors.New("wxId 无效")

// ErrWxDeactivateNotFound 注销目标不存在（已注销或记录不存在）。
var ErrWxDeactivateNotFound = errors.New("账号不存在或已注销")

// WxUserProfile 当前账号 profile 读模型（不含 unionid/apple_sub/password 等敏感字段）。
type WxUserProfile struct {
	IsWxBound     bool
	IsAppleBound  bool
	AuthProviders []string
	Account       string
	DeviceNo      string
}

// WxDeviceLoginByDeviceNo 仅校验 user 表已注册该 device_no；若 wx 已绑定同号则返回对应 wxId，否则 wxId=0（网关仍凭 device_no 签发 access）。
// 不使用 Scan 查空集：部分驱动会返回 sql.ErrNoRows，经网关拼接进 message。
func WxDeviceLoginByDeviceNo(ctx context.Context, deviceNo string) (*WxLoginResult, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil, ErrWxDeviceLoginDeviceNoEmpty
	}
	if err := DeviceAdmin().EnsureRegistered(ctx, deviceNo); err != nil {
		if errors.Is(err, ErrDeviceNotRegistered) {
			return nil, ErrWxDeviceLoginRejected
		}
		return nil, err
	}
	one, err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().DeviceNo, deviceNo).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return &WxLoginResult{WxId: 0, DeviceNo: deviceNo, IsNewUser: false}, nil
	}
	var row entity.Wx
	if err := one.Struct(&row); err != nil {
		return nil, err
	}
	if row.Id == 0 || strings.TrimSpace(row.DeviceNo) != deviceNo {
		return &WxLoginResult{WxId: 0, DeviceNo: deviceNo, IsNewUser: false}, nil
	}
	_ = invalidateWxCaches(ctx, row.Id, strings.TrimSpace(row.Unionid))
	return &WxLoginResult{
		WxId:      row.Id,
		DeviceNo:  deviceNo,
		IsNewUser: false,
	}, nil
}

// WxLogin 将客户端 jsCode（微信开放平台授权临时 code）经 OAuth 换票得到 unionid，再按 unionid 查找或创建 wx 行。
func WxLogin(ctx context.Context, jsCode, platform string) (*WxLoginResult, error) {
	oauth, err := exchangeAuthCodeForUnionID(ctx, platform, jsCode)
	if err != nil {
		return nil, err
	}
	unionID := strings.TrimSpace(oauth.UnionID)
	platform = strings.TrimSpace(platform)

	one, err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Unionid, unionID).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		res, insErr := dao.Wx.Ctx(ctx).Data(g.Map{
			dao.Wx.Columns().Unionid:   unionID,
			dao.Wx.Columns().Platform: platform,
			dao.Wx.Columns().CreatedAt: time.Now().Unix(),
		}).Insert()
		if insErr != nil {
			return nil, insErr
		}
		newID, _ := res.LastInsertId()
		_ = invalidateWxCaches(ctx, newID, unionID)
		return &WxLoginResult{
			WxId:      newID,
			DeviceNo:  "",
			IsNewUser: true,
		}, nil
	}
	var row entity.Wx
	if err := one.Struct(&row); err != nil {
		return nil, err
	}
	_ = invalidateWxCaches(ctx, row.Id, unionID)
	return &WxLoginResult{
		WxId:      row.Id,
		DeviceNo:  strings.TrimSpace(row.DeviceNo),
		IsNewUser: false,
	}, nil
}

// WxBindDevice 将 deviceNo 绑定到 wx 行（wx 主键由 Header X-Internal-Wx-Id 标识，由 controller 传入 id）。
func WxBindDevice(ctx context.Context, wxID int64, deviceNo string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if wxID <= 0 || deviceNo == "" {
		return errors.New("wxId 或 deviceNo 不能为空")
	}
	svc := DeviceAdmin()
	// 确保设备已注册
	if err := svc.EnsureRegistered(ctx, deviceNo); err != nil {
		return err
	}
	_, err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Id, wxID).Data(g.Map{
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

// WxAutoSaveProfile 未绑定设备时生成 6 位大写随机 device_no 并注册设备，再写画像；已绑定则只更新画像。
func WxAutoSaveProfile(ctx context.Context, wxID int64, babyName string, birthdayUnixSec int64, sex int) (string, error) {
	if wxID <= 0 {
		return "", errors.New("wxId 无效")
	}
	row, err := wxRowByWxID(ctx, wxID)
	if err != nil {
		return "", err
	}
	if row == nil || row.Id == 0 {
		return "", errors.New("wx 记录不存在，请先登录")
	}
	unionID := strings.TrimSpace(row.Unionid)
	svc := DeviceAdmin()
	deviceNo := strings.TrimSpace(row.DeviceNo)
	if deviceNo != "" {
		if err := svc.SaveUserProfile(ctx, deviceNo, strings.TrimSpace(babyName), birthdayUnixSec, sex); err != nil {
			return "", err
		}
		_ = invalidateWxCaches(ctx, row.Id, unionID)
		return deviceNo, nil
	}
	dn, genErr := generateUniqueDeviceNo(ctx)
	if genErr != nil {
		return "", genErr
	}
	if _, err := svc.Register(ctx, dn); err != nil {
		return "", err
	}
	if _, err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Id, row.Id).Data(g.Map{
		dao.Wx.Columns().DeviceNo: dn,
	}).Update(); err != nil {
		return "", err
	}
	if err := svc.SaveUserProfile(ctx, dn, strings.TrimSpace(babyName), birthdayUnixSec, sex); err != nil {
		return "", err
	}
	_ = invalidateWxCaches(ctx, row.Id, unionID)
	return dn, nil
}

// WxDeactivateByID 按 wx 主键执行账号注销（删除 wx 单条记录）。
// 失败语义：
// 1) wxId 非法返回 ErrWxDeactivateWxIDInvalid；
// 2) 目标不存在返回 ErrWxDeactivateNotFound；
// 3) 删除失败返回底层数据库错误。
func WxDeactivateByID(ctx context.Context, wxID int64) error {
	if wxID <= 0 {
		return ErrWxDeactivateWxIDInvalid
	}
	row, err := wxRowByWxID(ctx, wxID)
	if err != nil {
		return err
	}
	if row == nil || row.Id == 0 {
		return ErrWxDeactivateNotFound
	}
	if _, err = dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Id, wxID).Delete(); err != nil {
		return err
	}
	_ = invalidateWxCaches(ctx, row.Id, strings.TrimSpace(row.Unionid))
	return nil
}

// WxUserProfileByWxID 按 wx 主键读取账号 profile（单次查库，派生 isWxBound/account/deviceNo）。
func WxUserProfileByWxID(ctx context.Context, wxID int64) (*WxUserProfile, error) {
	if wxID <= 0 {
		return nil, ErrWxDeactivateWxIDInvalid
	}
	row, err := wxRowByWxID(ctx, wxID)
	if err != nil {
		return nil, err
	}
	if row == nil || row.Id == 0 {
		return nil, ErrWxDeactivateNotFound
	}
	return &WxUserProfile{
		IsWxBound:     strings.TrimSpace(row.Unionid) != "",
		IsAppleBound:  strings.TrimSpace(row.AppleSub) != "",
		AuthProviders: deriveAuthProviders(row),
		Account:       strings.TrimSpace(row.Account),
		DeviceNo:      strings.TrimSpace(row.DeviceNo),
	}, nil
}

// WxDeviceNoByWxID 按 wx 主键返回已绑定 device_no。
func WxDeviceNoByWxID(ctx context.Context, wxID int64) (string, error) {
	if wxID <= 0 {
		return "", errors.New("wxId 无效")
	}
	key := cachekit.WxIDToDeviceKey(wxID)
	if v, ok, err := wxLookupCache.Get(ctx, key); err == nil && ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v), nil
	}
	row, err := wxRowByWxID(ctx, wxID)
	if err != nil {
		return "", err
	}
	if row == nil || row.Id == 0 {
		return "", errors.New("wx 记录不存在")
	}
	dn := strings.TrimSpace(row.DeviceNo)
	if dn != "" {
		_ = wxLookupCache.SetEX(ctx, key, dn, cacheTTLWxLookup)
	}
	return dn, nil
}

// WxUnionIDByID 供运维/内部接口；网关 Bearer 热路径不再依赖。
func WxUnionIDByID(ctx context.Context, id int64) (string, error) {
	if id <= 0 {
		return "", errors.New("id 无效")
	}
	key := cachekit.WxIDToUnionKey(id)
	if v, ok, err := wxLookupCache.Get(ctx, key); err == nil && ok && v != "" {
		return v, nil
	}
	one, err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Id, id).Limit(1).One()
	if err != nil {
		return "", err
	}
	if one.IsEmpty() {
		return "", errors.New("wx 记录不存在")
	}
	var row entity.Wx
	if err = one.Struct(&row); err != nil {
		return "", err
	}
	if row.Id == 0 || strings.TrimSpace(row.Unionid) == "" {
		return "", errors.New("wx 记录不存在")
	}
	u := strings.TrimSpace(row.Unionid)
	_ = wxLookupCache.SetEX(ctx, key, u, cacheTTLWxLookup)
	return u, nil
}

// ValidateGatewayInternalSecret 校验网关调用内部接口时携带的共享密钥。
func ValidateGatewayInternalSecret(headerVal string) bool {
	sec := strings.TrimSpace(os.Getenv(envGatewayInternalSecret))
	if sec == "" {
		return false
	}
	return strings.TrimSpace(headerVal) == sec
}

func wxRowByWxID(ctx context.Context, wxID int64) (*entity.Wx, error) {
	one, err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Id, wxID).Limit(1).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return nil, nil
	}
	var row entity.Wx
	if err = one.Struct(&row); err != nil {
		return nil, err
	}
	if row.Id == 0 {
		return nil, nil
	}
	return &row, nil
}

func invalidateWxCaches(ctx context.Context, id int64, unionID string) error {
	_ = wxLookupCache.Del(ctx, cachekit.WxIDToUnionKey(id))
	_ = wxLookupCache.Del(ctx, cachekit.WxUnionToDeviceKey(strings.TrimSpace(unionID)))
	_ = wxLookupCache.Del(ctx, cachekit.WxIDToDeviceKey(id))
	return nil
}

func generateUniqueDeviceNo(ctx context.Context) (string, error) {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for attempt := 0; attempt < maxRandomDeviceAttempts; attempt++ {
		var b strings.Builder
		for i := 0; i < 6; i++ {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
			if err != nil {
				return "", err
			}
			b.WriteByte(letters[n.Int64()])
		}
		candidate := b.String()
		cnt, err := dao.User.Ctx(ctx).Where(dao.User.Columns().DeviceNo, candidate).Count()
		if err != nil {
			return "", err
		}
		if cnt == 0 {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("生成唯一 device_no 失败，已达最大重试次数 %d", maxRandomDeviceAttempts)
}

// HeaderInternalWxId 与 gateway-app 注入头一致，供 device controller 读取。
func HeaderInternalWxId() string { return headerInternalWxId }
