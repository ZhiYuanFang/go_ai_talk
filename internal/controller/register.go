package controller

import (
	"hello/internal/service"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
)

// RegisterHTTP 注册静态页、设备管理/历史 API、语音 WebSocket 以及 GoFrame 路由组。
func RegisterHTTP(s *ghttp.Server) {
	s.SetFileServerEnabled(true)
	s.SetIndexFolder(true)
	s.BindHandler("/", func(r *ghttp.Request) {
		r.Response.Write("智能语音")
	})

	s.BindHandler("/device/admin", func(r *ghttp.Request) {
		r.Response.ServeFile("resource/public/admin.html")
	})

	s.BindHandler("/device/history/*deviceNo", func(r *ghttp.Request) {
		r.Response.ServeFile("resource/public/history.html")
	})

	// 历史 API 代理中间件。控制流量流向，支持本地新版、代理旧版、金丝雀发布。
	installHistoryProxyMiddleware(s)
	registerVoiceChatWS(s)

	s.Use(ghttp.MiddlewareHandlerResponse)
	deps := service.NewHTTPDeps()
	s.Group("/", func(group *ghttp.RouterGroup) {
		registerHistoryRoutes(group, deps)
		registerAdminRoutes(group, deps)
		registerVoiceTextRoutes(group, deps)
	})
	s.SetServerRoot(gfile.MainPkgPath())
}

func registerHistoryRoutes(group *ghttp.RouterGroup, deps service.HTTPDeps) {
	group.Bind(NewHistoryCtrl(deps.History, deps.Voice))
}

func registerAdminRoutes(group *ghttp.RouterGroup, deps service.HTTPDeps) {
	group.Bind(NewAdminCtrl(deps.Admin))
}

func registerVoiceTextRoutes(group *ghttp.RouterGroup, deps service.HTTPDeps) {
	group.Bind(NewVoiceTextCtrl(deps.Voice, deps.Admin), Voice)
}
