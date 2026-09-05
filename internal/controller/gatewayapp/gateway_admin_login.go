package gatewayappctrl

import (
	"context"
	"sync"
	"time"

	v1 "hello/api/v1"
	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

const gatewayAdminLoginMinInterval = time.Second

var (
	gatewayAdminLoginBurst  sync.Mutex
	gatewayAdminLastLoginIP = map[string]time.Time{}
)

// GatewayAdminLoginCtrl 运维 Hub 登录（签发 Admin JWT）。
type GatewayAdminLoginCtrl struct{}

func NewGatewayAdminLoginCtrl() *GatewayAdminLoginCtrl {
	return &GatewayAdminLoginCtrl{}
}

// Login POST /device/admin/api/login
func (c *GatewayAdminLoginCtrl) Login(ctx context.Context, req *v1.GatewayAdminLoginReq) (res *v1.GatewayAdminLoginRes, err error) {
	if r := ghttp.RequestFromCtx(ctx); r != nil {
		remote := gatewayapp.ClientIP(r)
		gatewayAdminLoginBurst.Lock()
		if t, ok := gatewayAdminLastLoginIP[remote]; ok && time.Since(t) < gatewayAdminLoginMinInterval {
			gatewayAdminLoginBurst.Unlock()
			return nil, gerror.NewCode(gcode.CodeInvalidOperation, "请求过快")
		}
		gatewayAdminLastLoginIP[remote] = time.Now()
		gatewayAdminLoginBurst.Unlock()
	}

	if !gatewayapp.AdminLoginEnabled() {
		glog.Warningf(ctx, "[gateway-admin-login] 未配置 GATEWAY_APP_ADMIN_PASSWORD，拒绝登录")
		return nil, gerror.NewCode(gcode.CodeNotSupported, "管理未启用（未配置口令）")
	}
	if !gatewayapp.VerifyAdminLogin(req.Username, req.Password) {
		if r := ghttp.RequestFromCtx(ctx); r != nil {
			glog.Warningf(ctx, "[gateway-admin-login] 账号或密码错误 remote=%s", gatewayapp.ClientIP(r))
		}
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "账号或密码错误")
	}
	token, expiresIn, err := gatewayapp.SignAdminAccess(ctx, req.Username)
	if err != nil {
		glog.Warningf(ctx, "[gateway-admin-login] 签发 JWT 失败 err=%v", err)
		return nil, gerror.NewCode(gcode.CodeInternalError, "签发令牌失败")
	}
	return &v1.GatewayAdminLoginRes{AccessToken: token, ExpiresIn: expiresIn}, nil
}

// requireGatewayAdminJWT 本机 admin handler 校验 Admin JWT 已由 Hook 标记；失败返回未授权。
func requireGatewayAdminJWT(ctx context.Context) error {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil || !gatewayapp.RequestAdminVerified(r) {
		return gerror.NewCode(gcode.CodeNotAuthorized, "请先登录管理 Hub")
	}
	if !gatewayapp.AdminLoginEnabled() {
		return gerror.NewCode(gcode.CodeNotSupported, "管理未启用（未配置口令）")
	}
	return nil
}
