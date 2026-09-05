package notifyctrl

import (
	"strings"

	gatewayappctrl "hello/internal/controller/gatewayapp"
	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterHTTP 注册 notify-service 路由（无 MySQL/Redis 依赖）。
func RegisterHTTP(s *ghttp.Server) {
	s.Use(ghttp.MiddlewareHandlerResponse)
	installAppStatusAdminJWTMiddleware(s)
	registerAppStatusStaticPages(s)

	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Bind(NewAppStatusBannerCtrl(), NewAppStatusAdminCtrl())
	})
}

func installAppStatusAdminJWTMiddleware(s *ghttp.Server) {
	s.BindHookHandler("/*", ghttp.HookBeforeServe, func(r *ghttp.Request) {
		path := r.URL.Path
		if !strings.HasPrefix(path, "/admin/api/") {
			return
		}
		if path == "/admin/api/login" {
			return
		}
		raw := gatewayapp.BearerTokenFromRequest(r)
		if raw == "" {
			gatewayappctrl.WriteAuthJSON(r, 401, "缺少或无效的 Authorization")
			return
		}
		if _, err := gatewayapp.ParseAdminClaims(r.Context(), raw); err != nil {
			gatewayappctrl.WriteAuthJSON(r, 401, "admin access_token 无效或已过期")
			return
		}
		gatewayapp.MarkAdminJWTVerified(r)
	})
}

func registerAppStatusStaticPages(s *ghttp.Server) {
	s.BindHandler("/admin", func(r *ghttp.Request) {
		r.Response.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		r.Response.ServeFile("resource/public/app-status-admin.html")
	})
	s.BindHandler("/resource/public/*", func(r *ghttp.Request) {
		r.Response.ServeFile("resource/public/" + strings.TrimPrefix(r.URL.Path, "/resource/public/"))
	})
}
