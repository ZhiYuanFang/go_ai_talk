package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/glog"
)

// GatewayAppCtrl App 网关自有接口（登录、刷新、版本）。
type GatewayAppCtrl struct{}

func NewGatewayAppCtrl() *GatewayAppCtrl { return &GatewayAppCtrl{} }

// Login POST /device/app/api/login
func (c *GatewayAppCtrl) Login(ctx context.Context, req *v1.GatewayAppLoginReq) (res *v1.GatewayAppLoginRes, err error) {
	base := deviceServiceBase(ctx)
	if base == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidConfiguration, "DEVICE_SERVICE_URL 未配置")
	}
	url := strings.TrimRight(base, "/") + "/device/app/api/user/login"
	resp, err := gclient.New().ContentJson().Post(ctx, url, g.Map{
		"jsCode":   strings.TrimSpace(req.JsCode),
		"platform": strings.TrimSpace(req.Platform),
	})
	if err != nil {
		return nil, err
	}
	j := gjson.New(resp.ReadAllString())
	if j.Get("code").Int() != 0 {
		return nil, gerror.NewCodef(gcode.CodeBusinessValidationFailed, "微信登录失败: %s", j.Get("message").String())
	}
	data := j.GetJson("data")
	wxID := data.Get("wxId").Int64()
	deviceNo := strings.TrimSpace(data.Get("deviceNo").String())
	isNew := data.Get("isNewUser").Bool()
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInternalError, "device 返回缺少 wxId")
	}
	access, err := gatewayapp.SignAccess(ctx, wxID, deviceNo)
	if err != nil {
		return nil, err
	}
	refresh, err := gatewayapp.IssueRefreshToken(ctx, wxID, "")
	if err != nil {
		return nil, err
	}
	return &v1.GatewayAppLoginRes{
		WxId:         wxID,
		DeviceNo:     deviceNo,
		IsNewUser:    isNew,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

// DeviceLogin POST /device/app/api/device_login：device 设备号业务校验通过后签发 access/refresh。
func (c *GatewayAppCtrl) DeviceLogin(ctx context.Context, req *v1.GatewayAppDeviceLoginReq) (res *v1.GatewayAppDeviceLoginRes, err error) {
	base := deviceServiceBase(ctx)
	if base == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidConfiguration, "DEVICE_SERVICE_URL 未配置")
	}
	url := strings.TrimRight(base, "/") + "/device/app/api/user/device_login"
	resp, err := gclient.New().ContentJson().Post(ctx, url, g.Map{
		"deviceNo": strings.TrimSpace(req.DeviceNo),
	})
	if err != nil {
		return nil, err
	}
	j := gjson.New(resp.ReadAllString())
	if j.Get("code").Int() != 0 {
		return nil, gerror.NewCodef(gcode.CodeBusinessValidationFailed, "设备登录失败: %s", j.Get("message").String())
	}
	data := j.GetJson("data")
	wxID := data.Get("wxId").Int64()
	deviceNo := strings.TrimSpace(data.Get("deviceNo").String())
	isNew := data.Get("isNewUser").Bool()
	if wxID < 0 {
		return nil, gerror.NewCode(gcode.CodeInternalError, "device 返回 wxId 无效")
	}
	// 下游 data 未带 deviceNo 时，用本次请求体兜底，保证聚合响应与 JWT 内 device_no 一致（见 openspec gateway-app-device-login-return-device-no）。
	if deviceNo == "" {
		deviceNo = strings.TrimSpace(req.DeviceNo)
	}
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInternalError, "device 返回缺少 deviceNo")
	}
	access, err := gatewayapp.SignAccess(ctx, wxID, deviceNo)
	if err != nil {
		return nil, err
	}
	refresh, err := gatewayapp.IssueRefreshToken(ctx, wxID, deviceNo)
	if err != nil {
		return nil, err
	}
	return &v1.GatewayAppDeviceLoginRes{
		WxId:         wxID,
		DeviceNo:     deviceNo,
		IsNewUser:    isNew,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

// UsernameLogin POST /device/app/api/username_login：用户名聚合登录，device 完成业务校验后由 gateway 签发 access/refresh。
func (c *GatewayAppCtrl) UsernameLogin(ctx context.Context, req *v1.GatewayAppUsernameLoginReq) (res *v1.GatewayAppUsernameLoginRes, err error) {
	base := deviceServiceBase(ctx)
	if base == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidConfiguration, "DEVICE_SERVICE_URL 未配置")
	}
	url := strings.TrimRight(base, "/") + "/device/app/api/user/username/login"
	resp, err := gclient.New().ContentJson().Post(ctx, url, g.Map{
		"userName": strings.TrimSpace(req.UserName),
		"password": strings.TrimSpace(req.Password),
	})
	if err != nil {
		return nil, err
	}
	j := gjson.New(resp.ReadAllString())
	if j.Get("code").Int() != 0 {
		return nil, gerror.NewCodef(gcode.CodeBusinessValidationFailed, "用户名登录失败: %s", j.Get("message").String())
	}
	data := j.GetJson("data")
	wxID := data.Get("wxId").Int64()
	deviceNo := strings.TrimSpace(data.Get("deviceNo").String())
	isNew := data.Get("isNewUser").Bool()
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInternalError, "device 返回 wxId 无效")
	}
	access, err := gatewayapp.SignAccess(ctx, wxID, deviceNo)
	if err != nil {
		return nil, err
	}
	refresh, err := gatewayapp.IssueRefreshToken(ctx, wxID, deviceNo)
	if err != nil {
		return nil, err
	}
	return &v1.GatewayAppUsernameLoginRes{
		WxId:         wxID,
		DeviceNo:     deviceNo,
		IsNewUser:    isNew,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

// TokenRefresh POST /device/app/api/token/refresh（单次旋转 refresh）。
func (c *GatewayAppCtrl) TokenRefresh(ctx context.Context, req *v1.GatewayAppTokenRefreshReq) (res *v1.GatewayAppTokenRefreshRes, err error) {
	wxID, rtDeviceNo, err := gatewayapp.ConsumeRefreshToken(ctx, req.RefreshToken, true)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, err.Error())
	}
	deviceNo := strings.TrimSpace(rtDeviceNo)
	if deviceNo == "" && wxID > 0 {
		if dn, e2 := gatewayapp.FetchDeviceNoByWxID(ctx, wxID); e2 == nil {
			deviceNo = strings.TrimSpace(dn)
		}
	}
	access, err := gatewayapp.SignAccess(ctx, wxID, deviceNo)
	if err != nil {
		return nil, err
	}
	refresh, err := gatewayapp.IssueRefreshToken(ctx, wxID, deviceNo)
	if err != nil {
		return nil, err
	}
	return &v1.GatewayAppTokenRefreshRes{AccessToken: access, RefreshToken: refresh}, nil
}

// VersionCheck GET /device/app/api/version/check
func (c *GatewayAppCtrl) VersionCheck(ctx context.Context, req *v1.GatewayAppVersionCheckReq) (res *v1.GatewayAppVersionCheckRes, err error) {
	cur := strings.TrimSpace(req.CurrentVersion)
	row, ok := loadLatestAppVersionRow(ctx)
	if !ok {
		return versionCheckWhenNoPublishedRow(cur), nil
	}
	return buildVersionRes(cur, row), nil
}

// versionCheckWhenNoPublishedRow 版本表无行或无可用的 latestVersion：业务成功、无需更新；latestVersion 回填为客户端当前版本（与历史 Scan 失败兜底一致）。
func versionCheckWhenNoPublishedRow(current string) *v1.GatewayAppVersionCheckRes {
	return &v1.GatewayAppVersionCheckRes{
		NeedUpdate:    false,
		LatestVersion: strings.TrimSpace(current),
		ReleaseDate:   0,
		ReleaseNotes:  "",
		DownloadUrl:   "",
		ForceUpdate:   false,
	}
}

func buildVersionRes(current string, row entity.AppVersion) *v1.GatewayAppVersionCheckRes {
	latest := strings.TrimSpace(row.LatestVersion)
	need := gatewayapp.ShouldNeedAppUpdate(current, latest)
	return &v1.GatewayAppVersionCheckRes{
		NeedUpdate:    need,
		LatestVersion: latest,
		ReleaseDate:   row.ReleaseDate,
		ReleaseNotes:  strings.TrimSpace(row.ReleaseNotes),
		DownloadUrl:   gatewayapp.NormalizeAssetPath(strings.TrimSpace(row.DownloadUrl)),
		ForceUpdate:   row.ForceUpdate != 0,
	}
}

// loadLatestAppVersionRow 读取最新版本行；失败或无可用版本时返回 ok=false，由调用方决定降级语义。
func loadLatestAppVersionRow(ctx context.Context) (row entity.AppVersion, ok bool) {
	cacheKey := gatewayapp.RedisKeyAppVersionLatestCache
	if raw, err := g.Redis().Do(ctx, "GET", cacheKey); err == nil && raw != nil {
		s := strings.TrimSpace(raw.String())
		if s != "" {
			if err := json.Unmarshal([]byte(s), &row); err == nil && strings.TrimSpace(row.LatestVersion) != "" {
				return row, true
			}
		}
	}
	one, err := dao.AppVersion.Ctx(ctx).OrderDesc(dao.AppVersion.Columns().Id).Limit(1).One()
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			glog.Warningf(ctx, "[gateway-app] 读取 version 表失败 err=%v", err)
		}
		return entity.AppVersion{}, false
	}
	if one.IsEmpty() {
		return entity.AppVersion{}, false
	}
	if err := one.Struct(&row); err != nil {
		glog.Warningf(ctx, "[gateway-app] 解析 version 行失败 err=%v", err)
		return entity.AppVersion{}, false
	}
	if strings.TrimSpace(row.LatestVersion) == "" {
		return entity.AppVersion{}, false
	}
	if blob, err := json.Marshal(row); err == nil {
		_, _ = g.Redis().Do(ctx, "SET", cacheKey, string(blob), "EX", 60)
	}
	return row, true
}

func deviceServiceBase(ctx context.Context) string {
	return strings.TrimRight(strings.TrimSpace(gatewayapp.DeviceServiceBaseURL(ctx)), "/")
}
