package notifyctrl

import (
	"context"
	"sync"
	"time"

	v1 "hello/api/v1"
	"hello/internal/services/appstatus"
	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

const appStatusAdminLoginMinInterval = time.Second

var (
	appStatusAdminLoginBurst  sync.Mutex
	appStatusAdminLastLoginIP = map[string]time.Time{}
)

// AppStatusAdminCtrl 独立 status 服务运维 API（Admin JWT）。
type AppStatusAdminCtrl struct{}

func NewAppStatusAdminCtrl() *AppStatusAdminCtrl { return &AppStatusAdminCtrl{} }

func requireAppStatusAdminJWT(ctx context.Context) error {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil || !gatewayapp.RequestAdminVerified(r) {
		return gerror.NewCode(gcode.CodeNotAuthorized, "请先登录")
	}
	if !gatewayapp.AdminLoginEnabled() {
		return gerror.NewCode(gcode.CodeNotSupported, "管理未启用（未配置口令）")
	}
	return nil
}

// Login POST /admin/api/login
func (c *AppStatusAdminCtrl) Login(ctx context.Context, req *v1.AppStatusAdminLoginReq) (res *v1.AppStatusAdminLoginRes, err error) {
	_ = c
	if r := ghttp.RequestFromCtx(ctx); r != nil {
		remote := gatewayapp.ClientIP(r)
		appStatusAdminLoginBurst.Lock()
		if t, ok := appStatusAdminLastLoginIP[remote]; ok && time.Since(t) < appStatusAdminLoginMinInterval {
			appStatusAdminLoginBurst.Unlock()
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, "请求过快")
		}
		appStatusAdminLastLoginIP[remote] = time.Now()
		appStatusAdminLoginBurst.Unlock()
	}
	if !gatewayapp.AdminLoginEnabled() {
		glog.Warningf(ctx, "[app-status-admin-login] 未配置 GATEWAY_APP_ADMIN_PASSWORD，拒绝登录")
		return nil, gerror.NewCode(gcode.CodeNotSupported, "管理未启用（未配置口令）")
	}
	if !gatewayapp.VerifyAdminLogin(req.Username, req.Password) {
		if r := ghttp.RequestFromCtx(ctx); r != nil {
			glog.Warningf(ctx, "[app-status-admin-login] 账号或密码错误 remote=%s", gatewayapp.ClientIP(r))
		}
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "账号或密码错误")
	}
	token, expiresIn, err := gatewayapp.SignAdminAccess(ctx, req.Username)
	if err != nil {
		glog.Warningf(ctx, "[app-status-admin-login] 签发 JWT 失败 err=%v", err)
		return nil, gerror.NewCode(gcode.CodeInternalError, "签发令牌失败")
	}
	return &v1.AppStatusAdminLoginRes{AccessToken: token, ExpiresIn: expiresIn}, nil
}

// BannerGet GET /admin/api/banner
func (c *AppStatusAdminCtrl) BannerGet(ctx context.Context, req *v1.AppStatusAdminBannerGetReq) (res *v1.AppStatusAdminBannerGetRes, err error) {
	_ = c
	_ = req
	if err = requireAppStatusAdminJWT(ctx); err != nil {
		return nil, err
	}
	return bannerToAdminRes(appstatus.Snapshot()), nil
}

// BannerPut PUT /admin/api/banner
func (c *AppStatusAdminCtrl) BannerPut(ctx context.Context, req *v1.AppStatusAdminBannerPutReq) (res *v1.AppStatusAdminBannerPutRes, err error) {
	_ = c
	if err = requireAppStatusAdminJWT(ctx); err != nil {
		return nil, err
	}
	next := appstatus.Update(appstatus.BannerState{
		Active:        req.Active,
		Title:         req.Title,
		Message:       req.Message,
		ExpectedEndAt: req.ExpectedEndAt,
		Dismissible:   req.Dismissible,
	})
	admin := bannerToAdminRes(next)
	return &v1.AppStatusAdminBannerPutRes{
		Active:        admin.Active,
		Title:         admin.Title,
		Message:       admin.Message,
		ExpectedEndAt: admin.ExpectedEndAt,
		Dismissible:   admin.Dismissible,
		UpdatedAt:     admin.UpdatedAt,
		ContentKey:    admin.ContentKey,
	}, nil
}
