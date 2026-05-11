package controller

import (
	"context"
	"encoding/json"
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
	url := strings.TrimRight(base, "/") + "/device/wx/api/login"
	resp, err := gclient.New().ContentJson().Post(ctx, url, g.Map{
		"wxCode":   strings.TrimSpace(req.WxCode),
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
	wxCode := strings.TrimSpace(data.Get("wxCode").String())
	deviceNo := strings.TrimSpace(data.Get("deviceNo").String())
	isNew := data.Get("isNewUser").Bool()
	if wxID <= 0 || wxCode == "" {
		return nil, gerror.NewCode(gcode.CodeInternalError, "device 返回缺少 wxId/wxCode")
	}
	access, err := gatewayapp.SignAccess(ctx, wxID)
	if err != nil {
		return nil, err
	}
	refresh, err := gatewayapp.IssueRefreshToken(ctx, wxID)
	if err != nil {
		return nil, err
	}
	return &v1.GatewayAppLoginRes{
		WxId:         wxID,
		WxCode:       wxCode,
		DeviceNo:     deviceNo,
		IsNewUser:    isNew,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

// TokenRefresh POST /device/app/api/token/refresh（单次旋转 refresh）。
func (c *GatewayAppCtrl) TokenRefresh(ctx context.Context, req *v1.GatewayAppTokenRefreshReq) (res *v1.GatewayAppTokenRefreshRes, err error) {
	wxID, err := gatewayapp.ConsumeRefreshToken(ctx, req.RefreshToken, true)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, err.Error())
	}
	access, err := gatewayapp.SignAccess(ctx, wxID)
	if err != nil {
		return nil, err
	}
	refresh, err := gatewayapp.IssueRefreshToken(ctx, wxID)
	if err != nil {
		return nil, err
	}
	return &v1.GatewayAppTokenRefreshRes{AccessToken: access, RefreshToken: refresh}, nil
}

// VersionCheck GET /device/app/api/version/check
func (c *GatewayAppCtrl) VersionCheck(ctx context.Context, req *v1.GatewayAppVersionCheckReq) (res *v1.GatewayAppVersionCheckRes, err error) {
	cur := strings.TrimSpace(req.CurrentVersion)
	cacheKey := "gw:app:version:latest"
	if raw, err := g.Redis().Do(ctx, "GET", cacheKey); err == nil && raw != nil {
		s := strings.TrimSpace(raw.String())
		if s != "" {
			var row entity.AppVersion
			if err := json.Unmarshal([]byte(s), &row); err == nil && strings.TrimSpace(row.LatestVersion) != "" {
				return buildVersionRes(cur, row), nil
			}
		}
	}
	var row entity.AppVersion
	if err := dao.AppVersion.Ctx(ctx).OrderDesc(dao.AppVersion.Columns().Id).Limit(1).Scan(&row); err != nil {
		glog.Warningf(ctx, "[gateway-app] 读取 version 表失败 err=%v", err)
		return &v1.GatewayAppVersionCheckRes{
			NeedUpdate:    false,
			LatestVersion: cur,
			ReleaseDate:   0,
		}, nil
	}
	if blob, err := json.Marshal(row); err == nil {
		_, _ = g.Redis().Do(ctx, "SET", cacheKey, string(blob), "EX", 60)
	}
	return buildVersionRes(cur, row), nil
}

func buildVersionRes(current string, row entity.AppVersion) *v1.GatewayAppVersionCheckRes {
	latest := strings.TrimSpace(row.LatestVersion)
	need := latest != "" && strings.TrimSpace(current) != latest
	return &v1.GatewayAppVersionCheckRes{
		NeedUpdate:    need,
		LatestVersion: latest,
		ReleaseDate:   row.ReleaseDate,
		ReleaseNotes:  strings.TrimSpace(row.ReleaseNotes),
		DownloadUrl:   strings.TrimSpace(row.DownloadUrl),
		ForceUpdate:   row.ForceUpdate != 0,
	}
}

func deviceServiceBase(ctx context.Context) string {
	return strings.TrimRight(strings.TrimSpace(gatewayapp.DeviceServiceBaseURL(ctx)), "/")
}
