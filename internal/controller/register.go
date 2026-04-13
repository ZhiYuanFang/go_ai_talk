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

	registerVoiceChatWS(s)

	s.Use(ghttp.MiddlewareHandlerResponse)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Bind(
			NewHistoryCtrl(service.DeviceHistory(), service.Voice()),
			NewAdminCtrl(service.DeviceAdmin()),
			NewVoiceTextCtrl(service.Voice(), service.DeviceAdmin()),
			Voice,
		)
	})
	s.SetServerRoot(gfile.MainPkgPath())
}
