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
		group.Bind(NewUcgAppCtrl())                                          // 绑定 App API 控制器（包含 WS 升级接口）；内部 API 在 controller/ucg_internal.go 注册。
		group.Bind(NewUcgAdminCtrl())                                        // 绑定 Admin API 控制器, 给运维使用。
		group.POST("/ucg/app/api/media/upload", ucgMediaUpload)              // 给 Web 前端提供的媒体上传接口（同域代理，规避 OSS CORS）；内部 API 在 controller/ucg_internal.go 注册。
		group.POST("/ucg/internal/api/media/upload", ucgInternalMediaUpload) // 给 device 管理端提供的媒体上传接口（内部接口，鉴权更严格）；内部 API 在 controller/ucg_internal.go 注册。
		group.POST("/ucg/internal/api/chat/send", ucgInternalChatSend)       // 内部 API，在 controller/ucg_internal.go 注册。
		group.POST("/ucg/internal/api/posts/sample", ucgInternalPostsSample) // 内部 API，在 controller/ucg_internal.go 注册。
		group.POST("/ucg/internal/api/chat/sim-unread-sample", ucgInternalChatSimUnreadSample)
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
