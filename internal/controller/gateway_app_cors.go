package controller

import (
	"net/http"

	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	gatewayAppCORSAllowMethods = "GET, POST, OPTIONS"
	gatewayAppCORSAllowHeaders = "Content-Type, Authorization"
	gatewayAppCORSMaxAge       = "86400"
)

// installGatewayAppCORSMiddleware 仅为 gateway-app-server 安装 CORS：白名单主机见 gatewayapp.ReflectGatewayAppCORSOrigin。
// 在 Next 之后再次写入头，以便反代下游刷写响应后仍带上 CORS。
// 任意 OPTIONS 必须在此短路：若仅在白名单命中时才 204，则 Origin 不在白名单时预检会进入后续路由，
// GoFrame 可能把 OPTIONS 交给仅声明了 POST 的接口，触发空 body 业务校验（如 device_login 报 deviceNo 不能为空）。
func installGatewayAppCORSMiddleware(s *ghttp.Server) {
	s.BindMiddleware("/*", func(r *ghttp.Request) {
		echo, ok := gatewayapp.ReflectGatewayAppCORSOrigin(r.Header.Get("Origin"))
		if ok {
			writeGatewayAppCORSHeaders(r, echo)
		}
		// OPTIONS 在 gatewayAppPathAuthExempt 中已整方法豁免 Bearer；预检一律 204，不再进入业务 handler。
		if r.Method == http.MethodOptions {
			r.Response.WriteStatusExit(http.StatusNoContent, nil)
			return
		}
		r.Middleware.Next()
		if ok {
			writeGatewayAppCORSHeaders(r, echo)
		}
	})
}

func writeGatewayAppCORSHeaders(r *ghttp.Request, allowOrigin string) {
	h := r.Response.Header()
	h.Set("Access-Control-Allow-Origin", allowOrigin)
	h.Set("Access-Control-Allow-Methods", gatewayAppCORSAllowMethods)
	h.Set("Access-Control-Allow-Headers", gatewayAppCORSAllowHeaders)
	h.Set("Access-Control-Max-Age", gatewayAppCORSMaxAge)
}
