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
	headerInternalWxCode     = "X-Internal-Wx-Code"
	headerGatewayInternalKey = "X-Gateway-Internal-Secret"
	envGatewayInternalSecret = "DEVICE_GATEWAY_INTERNAL_SECRET"
	cacheKeyWxIDToCode       = "dev:wx:id2code:"
	cacheKeyWxCodeToDevice   = "dev:wx:code2dev:"
	cacheTTLWxLookup         = 120 * time.Second
	maxRandomDeviceAttempts    = 15
)

var wxLookupCache = cachekit.WithObserver(cachekit.NewRedisCache(), cachekit.LoggingObserver{})

// WxLoginResult device 侧微信登录业务返回（不含 JWT）。
type WxLoginResult struct {
	WxId      int64  `json:"wxId"`
	WxCode    string `json:"wxCode"`
	DeviceNo  string `json:"deviceNo"`
	IsNewUser bool   `json:"isNewUser"`
}

// WxLogin 按 wxCode 查找或创建 wx 行，仅返回业务字段。
func WxLogin(ctx context.Context, wxCode, platform string) (*WxLoginResult, error) {
	wxCode = strings.TrimSpace(wxCode)
	if wxCode == "" {
		return nil, errors.New("wxCode 不能为空")
	}
	platform = strings.TrimSpace(platform)
	one, err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().WxCode, wxCode).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		res, insErr := dao.Wx.Ctx(ctx).Data(g.Map{
			dao.Wx.Columns().WxCode:    wxCode,
			dao.Wx.Columns().Platform: platform,
		}).Insert()
		if insErr != nil {
			return nil, insErr
		}
		newID, _ := res.LastInsertId()
		_ = invalidateWxCaches(ctx, newID, wxCode)
		return &WxLoginResult{
			WxId:      newID,
			WxCode:    wxCode,
			DeviceNo:  "",
			IsNewUser: true,
		}, nil
	}
	var row entity.Wx
	if err := one.Struct(&row); err != nil {
		return nil, err
	}
	_ = invalidateWxCaches(ctx, row.Id, wxCode)
	return &WxLoginResult{
		WxId:      row.Id,
		WxCode:    row.WxCode,
		DeviceNo:  strings.TrimSpace(row.DeviceNo),
		IsNewUser: false,
	}, nil
}

// WxBindDevice 将 deviceNo 绑定到当前 wx（wxCode 来自内部头，由 controller 传入）。
func WxBindDevice(ctx context.Context, wxCode, deviceNo string) error {
	wxCode = strings.TrimSpace(wxCode)
	deviceNo = strings.TrimSpace(deviceNo)
	if wxCode == "" || deviceNo == "" {
		return errors.New("wxCode 或 deviceNo 不能为空")
	}
	svc := DeviceAdmin()
	if err := svc.EnsureRegistered(ctx, deviceNo); err != nil {
		return err
	}
	_, err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().WxCode, wxCode).Data(g.Map{
		dao.Wx.Columns().DeviceNo: deviceNo,
	}).Update()
	if err != nil {
		return err
	}
	row, _ := wxRowByCode(ctx, wxCode)
	if row != nil {
		_ = invalidateWxCaches(ctx, row.Id, wxCode)
	}
	return nil
}

// WxAutoSaveProfile 未绑定设备时生成 6 位大写随机 device_no 并注册设备，再写画像；已绑定则只更新画像。
func WxAutoSaveProfile(ctx context.Context, wxCode, birthday string, sex int) (string, error) {
	wxCode = strings.TrimSpace(wxCode)
	if wxCode == "" {
		return "", errors.New("wxCode 不能为空")
	}
	row, err := wxRowByCode(ctx, wxCode)
	if err != nil {
		return "", err
	}
	if row == nil || row.Id == 0 {
		return "", errors.New("wx 记录不存在，请先登录")
	}
	svc := DeviceAdmin()
	deviceNo := strings.TrimSpace(row.DeviceNo)
	if deviceNo != "" {
		if err := svc.SaveUserProfile(ctx, deviceNo, birthday, sex); err != nil {
			return "", err
		}
		_ = invalidateWxCaches(ctx, row.Id, wxCode)
		return deviceNo, nil
	}
	// 未绑定：生成全局唯一 device_no 并注册 user，再绑定 wx。
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
	if err := svc.SaveUserProfile(ctx, dn, birthday, sex); err != nil {
		return "", err
	}
	_ = invalidateWxCaches(ctx, row.Id, wxCode)
	return dn, nil
}

// WxDeviceNoByCode 按 wxCode 返回已绑定 device_no。
func WxDeviceNoByCode(ctx context.Context, wxCode string) (string, error) {
	wxCode = strings.TrimSpace(wxCode)
	if wxCode == "" {
		return "", errors.New("wxCode 不能为空")
	}
	if v, ok, err := wxLookupCache.Get(ctx, cacheKeyWxCodeToDevice+wxCode); err == nil && ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v), nil
	}
	row, err := wxRowByCode(ctx, wxCode)
	if err != nil {
		return "", err
	}
	if row == nil || row.Id == 0 {
		return "", errors.New("wx 记录不存在")
	}
	dn := strings.TrimSpace(row.DeviceNo)
	if dn != "" {
		_ = wxLookupCache.SetEX(ctx, cacheKeyWxCodeToDevice+wxCode, dn, cacheTTLWxLookup)
	}
	return dn, nil
}

// WxCodeByID 供网关解析 JWT sub；调用方须已通过内部密钥校验（见 controller）。
func WxCodeByID(ctx context.Context, id int64) (string, error) {
	if id <= 0 {
		return "", errors.New("id 无效")
	}
	key := cacheKeyWxIDToCode + strconv.FormatInt(id, 10)
	if v, ok, err := wxLookupCache.Get(ctx, key); err == nil && ok && v != "" {
		return v, nil
	}
	var row entity.Wx
	if err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().Id, id).Scan(&row); err != nil {
		return "", err
	}
	if row.Id == 0 || strings.TrimSpace(row.WxCode) == "" {
		return "", errors.New("wx 记录不存在")
	}
	code := strings.TrimSpace(row.WxCode)
	_ = wxLookupCache.SetEX(ctx, key, code, cacheTTLWxLookup)
	return code, nil
}

// ValidateGatewayInternalSecret 校验网关调用内部接口时携带的共享密钥。
func ValidateGatewayInternalSecret(headerVal string) bool {
	sec := strings.TrimSpace(os.Getenv(envGatewayInternalSecret))
	if sec == "" {
		// 未配置时不放行内部接口，避免误暴露。
		return false
	}
	return strings.TrimSpace(headerVal) == sec
}

func wxRowByCode(ctx context.Context, wxCode string) (*entity.Wx, error) {
	var row entity.Wx
	if err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().WxCode, wxCode).Scan(&row); err != nil {
		return nil, err
	}
	if row.Id == 0 {
		return nil, nil
	}
	return &row, nil
}

func invalidateWxCaches(ctx context.Context, id int64, wxCode string) error {
	_ = wxLookupCache.Del(ctx, cacheKeyWxIDToCode+strconv.FormatInt(id, 10))
	_ = wxLookupCache.Del(ctx, cacheKeyWxCodeToDevice+strings.TrimSpace(wxCode))
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
