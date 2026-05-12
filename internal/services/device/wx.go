package device

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	headerInternalWxId = "X-Internal-Wx-Id"
	envGatewayInternalSecret = "DEVICE_GATEWAY_INTERNAL_SECRET"
	cacheKeyWxIDToUnion      = "dev:wx:id2union:"
	cacheKeyWxUnionToDevice  = "dev:wx:union2dev:"
	cacheKeyWxIDToDevice     = "dev:wx:id2dev:"
	cacheTTLWxLookup         = 120 * time.Second
	maxRandomDeviceAttempts  = 15
)

var wxLookupCache = cachekit.WithObserver(cachekit.NewRedisCache(), cachekit.LoggingObserver{})

// WxLoginResult device 侧微信登录业务返回（不含 JWT；不向客户端回传 unionid）。
type WxLoginResult struct {
	WxId      int64  `json:"wxId"`
	DeviceNo  string `json:"deviceNo"`
	IsNewUser bool   `json:"isNewUser"`
}

// ErrWxDeviceLoginRejected 设备号登录被拒绝：设备未注册或 wx 未绑定该设备。
// 故意不区分具体原因，降低通过响应枚举设备状态的攻击面。
var ErrWxDeviceLoginRejected = errors.New("设备登录失败，请确认设备已注册且已绑定账号")

// ErrWxDeviceLoginDeviceNoEmpty 入参 device_no 为空（trim 后）。
var ErrWxDeviceLoginDeviceNoEmpty = errors.New("deviceNo 不能为空")

// WxDeviceLoginByDeviceNo 校验 user 表已存在该 device_no，且 wx 行已绑定同一设备号；成功返回 wxId 等（不含 JWT）。
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
	var row entity.Wx
	if err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().DeviceNo, deviceNo).Scan(&row); err != nil {
		return nil, err
	}
	if row.Id == 0 || strings.TrimSpace(row.DeviceNo) != deviceNo {
		return nil, ErrWxDeviceLoginRejected
	}
	_ = invalidateWxCaches(ctx, row.Id, strings.TrimSpace(row.UnionId))
	return &WxLoginResult{
		WxId:      row.Id,
		DeviceNo:  deviceNo,
		IsNewUser: false,
	}, nil
}

// WxLogin 将客户端临时 js_code 经微信换票得到 unionid，再按 unionid 查找或创建 wx 行。
func WxLogin(ctx context.Context, jsCode, platform string) (*WxLoginResult, error) {
	sess, err := exchangeJsCodeForUnionID(ctx, platform, jsCode)
	if err != nil {
		return nil, err
	}
	_ = sess.SessionKey // 明确丢弃：不落库、不复用，避免成为攻击面
	unionID := strings.TrimSpace(sess.UnionID)
	platform = strings.TrimSpace(platform)

	one, err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().UnionId, unionID).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		res, insErr := dao.Wx.Ctx(ctx).Data(g.Map{
			dao.Wx.Columns().UnionId:   unionID,
			dao.Wx.Columns().Platform: platform,
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
		_ = invalidateWxCaches(ctx, row.Id, strings.TrimSpace(row.UnionId))
	}
	return nil
}

// WxAutoSaveProfile 未绑定设备时生成 6 位大写随机 device_no 并注册设备，再写画像；已绑定则只更新画像。
func WxAutoSaveProfile(ctx context.Context, wxID int64, birthdayUnixSec int64, sex int) (string, error) {
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
	unionID := strings.TrimSpace(row.UnionId)
	svc := DeviceAdmin()
	deviceNo := strings.TrimSpace(row.DeviceNo)
	if deviceNo != "" {
		if err := svc.SaveUserProfile(ctx, deviceNo, birthdayUnixSec, sex); err != nil {
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
	if err := svc.SaveUserProfile(ctx, dn, birthdayUnixSec, sex); err != nil {
		return "", err
	}
	_ = invalidateWxCaches(ctx, row.Id, unionID)
	return dn, nil
}

// WxDeviceNoByWxID 按 wx 主键返回已绑定 device_no。
func WxDeviceNoByWxID(ctx context.Context, wxID int64) (string, error) {
	if wxID <= 0 {
		return "", errors.New("wxId 无效")
	}
	key := cacheKeyWxIDToDevice + strconv.FormatInt(wxID, 10)
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
	key := cacheKeyWxIDToUnion + strconv.FormatInt(id, 10)
	if v, ok, err := wxLookupCache.Get(ctx, key); err == nil && ok && v != "" {
		return v, nil
	}
	var row entity.Wx
	if err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Id, id).Scan(&row); err != nil {
		return "", err
	}
	if row.Id == 0 || strings.TrimSpace(row.UnionId) == "" {
		return "", errors.New("wx 记录不存在")
	}
	u := strings.TrimSpace(row.UnionId)
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
	var row entity.Wx
	if err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Id, wxID).Scan(&row); err != nil {
		return nil, err
	}
	if row.Id == 0 {
		return nil, nil
	}
	return &row, nil
}

func invalidateWxCaches(ctx context.Context, id int64, unionID string) error {
	_ = wxLookupCache.Del(ctx, cacheKeyWxIDToUnion+strconv.FormatInt(id, 10))
	_ = wxLookupCache.Del(ctx, cacheKeyWxUnionToDevice+strings.TrimSpace(unionID))
	_ = wxLookupCache.Del(ctx, cacheKeyWxIDToDevice+strconv.FormatInt(id, 10))
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
