package controller

import (
	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// installGatewayAppBearerMiddleware 使用 HookBeforeServe 在「任意 BindMiddleware 反代短路」之前
// 对除白名单外的请求校验 Bearer 并注入 X-Internal-Wx-Code，从而覆盖经本进程的全部 API（含 history/voice/device 反代），
// 无需在各 *_route_proxy 中重复鉴权。参见 ghttp.Server.ServeHTTP：BeforeServe 先于 Middleware.Next()。
func installGatewayAppBearerMiddleware(s *ghttp.Server) {
	s.BindHookHandler("/*", ghttp.HookBeforeServe, func(r *ghttp.Request) {
		if gatewayAppPathAuthExempt(r) {
			return
		}
		if err := gatewayapp.InjectWxCodeFromBearer(r); err != nil {
			r.Response.Status = 401
			r.Response.WriteJson(g.Map{"code": 401, "message": err.Error()})
			r.ExitAll()
			return
		}
	})
}
