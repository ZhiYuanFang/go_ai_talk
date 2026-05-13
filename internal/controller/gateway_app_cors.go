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
// 在 Next 之后再次写入头，以便反代下游刷写响应后仍带上 CORS；OPTIONS 在白名单下短路 204，避免无路由时预检 404。
func installGatewayAppCORSMiddleware(s *ghttp.Server) {
	s.BindMiddleware("/*", func(r *ghttp.Request) {
		echo, ok := gatewayapp.ReflectGatewayAppCORSOrigin(r.Header.Get("Origin"))
		if ok {
			writeGatewayAppCORSHeaders(r, echo)
		}
		// OPTIONS 在 gatewayAppPathAuthExempt 中已整方法豁免 Bearer，此处仅保证预检为 2xx 且带齐 CORS。
		if ok && r.Method == http.MethodOptions {
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
