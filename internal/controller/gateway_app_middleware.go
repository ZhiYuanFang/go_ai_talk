package controller

import (
	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/net/ghttp"
)

// installGatewayAppBearerMiddleware 使用 HookBeforeServe 在「任意 BindMiddleware 反代短路」之前
// 分流 Admin JWT 与用户 JWT，并向反代注入下游 X-Admin-Password。
func installGatewayAppBearerMiddleware(s *ghttp.Server) {
	s.BindHookHandler("/*", ghttp.HookBeforeServe, func(r *ghttp.Request) {
		gatewayapp.StripSpoofedInternalHeaders(r)
		gatewayapp.InjectClientIPHeader(r)

		path := r.URL.Path

		if gatewayAppPathAuthExempt(r) {
			if raw := gatewayapp.BearerTokenFromRequest(r); raw != "" {
				if !gatewayapp.IsAdminJWTString(r.Context(), raw) {
					_ = gatewayapp.InjectAccessHeadersFromBearer(r)
				}
			}
			return
		}

		if gatewayapp.IsGatewayAdminAPIPath(path) {
			raw := gatewayapp.BearerTokenFromRequest(r)
			if raw == "" {
				writeGatewayAppAuthJSON(r, 401, "缺少或无效的 Authorization")
				return
			}
			if _, err := gatewayapp.ParseAdminClaims(r.Context(), raw); err != nil {
				writeGatewayAppAuthJSON(r, 401, "admin access_token 无效或已过期")
				return
			}
			gatewayapp.MarkAdminJWTVerified(r)
			gatewayapp.InjectAdminDownstreamPassword(r)
			return
		}

		raw := gatewayapp.BearerTokenFromRequest(r)
		if raw != "" && gatewayapp.IsAdminJWTString(r.Context(), raw) {
			writeGatewayAppAuthJSON(r, 403, "admin token 不能访问 App 用户接口")
			return
		}

		if err := gatewayapp.InjectAccessHeadersFromBearer(r); err != nil {
			writeGatewayAppAuthJSON(r, 401, err.Error())
			return
		}
	})
}
