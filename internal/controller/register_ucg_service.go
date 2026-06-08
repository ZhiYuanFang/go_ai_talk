package controller

import (
	"net/http"
	"net/http/httputil"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterUcgServiceHTTP 注册 ucg-service 独立进程路由（App API 与内部 WS 入口）。
func RegisterUcgServiceHTTP(s *ghttp.Server) {
	s.Use(ghttp.MiddlewareHandlerResponse)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Bind(NewUcgAppCtrl())
		group.Bind(NewUcgAdminCtrl())
		group.POST("/ucg/app/api/media/upload", ucgMediaUpload)
	})
	registerUcgChatWS(s)
}

// ucgChatWSUpgradeProxyDirector 将 gateway 对外路径 /ucg/app/ws/chat 改写为 ucg-service 内部 /ws/chat。
func ucgChatWSUpgradeProxyDirector(proxy *httputil.ReverseProxy) {
	if proxy == nil {
		return
	}
	orig := proxy.Director
	proxy.Director = func(req *http.Request) {
		if orig != nil {
			orig(req)
		}
		req.URL.Path = "/ws/chat"
		req.URL.RawPath = "/ws/chat"
	}
}
