package controller

import (
	"net/http"

	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/net/ghttp"
)

// installGatewayAppCORSMiddleware 仅为 gateway-app-server 安装 CORS：白名单主机见 gatewayapp.ReflectGatewayAppCORSOrigin。
// 在 Next 之后再次写入头，以便反代下游刷写响应后仍带上 CORS。
// 任意 OPTIONS 必须在此短路：若仅在白名单命中时才 204，则 Origin 不在白名单时预检会进入后续路由，
// GoFrame 可能把 OPTIONS 交给仅声明了 POST 的接口，触发空 body 业务校验（如 device_login 报 deviceNo 不能为空）。
func installGatewayAppCORSMiddleware(s *ghttp.Server) {
	s.BindMiddleware("/*", func(r *ghttp.Request) {
		ok := gatewayapp.ApplyGatewayAppCORSHeaders(r.Response.Header(), r.Header.Get("Origin"))
		// OPTIONS 在 gatewayAppPathAuthExempt 中已整方法豁免 Bearer；预检一律 204，不再进入业务 handler。
		if r.Method == http.MethodOptions {
			r.Response.WriteStatusExit(http.StatusNoContent, nil)
			return
		}
		r.Middleware.Next()
		if ok {
			_ = gatewayapp.ApplyGatewayAppCORSHeaders(r.Response.Header(), r.Header.Get("Origin"))
		}
	})
}
