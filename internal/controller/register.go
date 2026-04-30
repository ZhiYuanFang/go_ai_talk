package controller

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
)

// RegisterHTTP 注册静态页、设备管理/历史 API、语音 WebSocket 以及 GoFrame 路由组。
func RegisterHTTP(s *ghttp.Server) {
	// 网关仍承载静态资源与入口页，保证前端访问路径稳定。
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

	// 领域服务代理中间件统一安装入口（history/voice/device）。
	installDomainProxyMiddlewares(s)
	// voice WS 入口统一走边缘透传，领域执行下沉到 voice-service。
	installVoiceWSProxyMiddleware(s)
	// 安装全局横切能力（如请求 ID 透传），确保委派前先补齐上下文。
	installGatewayCrosscuttingMiddlewares(s)

	s.Use(ghttp.MiddlewareHandlerResponse)
	// 以主包路径作为资源根目录，兼容本地与容器运行环境。
	s.SetServerRoot(gfile.MainPkgPath())
}
