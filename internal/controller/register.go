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
		r.Response.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		r.Response.ServeFile("resource/public/admin.html")
	})

	s.BindHandler("/device/admin/qa-records", func(r *ghttp.Request) {
		r.Response.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		r.Response.ServeFile("resource/public/qa-records.html")
	})

	s.BindHandler("/device/admin/feedback-records", func(r *ghttp.Request) {
		r.Response.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		r.Response.ServeFile("resource/public/feedback-records.html")
	})

	s.BindHandler("/device/admin/api-usage-stats", func(r *ghttp.Request) {
		r.Response.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		r.Response.ServeFile("resource/public/api-usage-stats.html")
	})

	s.BindHandler("/device/admin/ucg-admin.html", func(r *ghttp.Request) {
		r.Response.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		r.Response.ServeFile("resource/public/ucg-admin.html")
	})

	// 事件 logo 已迁移至 OSS/CDN，不再反代 /ai_talk_images。

	s.BindHandler("/device/history/*deviceNo", func(r *ghttp.Request) {
		r.Response.ServeFile("resource/public/history.html")
	})

	// 领域服务代理中间件统一安装入口（history/voice/device）。
	installDomainProxyMiddlewares(s)
	// voice WS 入口统一走边缘透传，领域执行下沉到 voice-service。
	installVoiceWSProxyMiddleware(s)
	// 安装全局横切能力（如请求 ID 透传），确保委派前先补齐上下文。
	installGatewayCrosscuttingMiddlewares(s)
	installDeviceAPIAccessTouchMiddleware(s)

	// 网关动态入口以静态页与 httputil.ReverseProxy 透传为主，不在此挂载 MiddlewareHandlerResponse：
	// 下游若 Content-Length 未知（如 chunked/gzip 解压后），ReverseProxy 会对 ResponseWriter 触发 Flush，
	// GoFrame 缓冲被提前刷空后，若再套一层统一 JSON 封装会写出「第二段响应体」，客户端或前置 Nginx 常表现为 502。
	// device-service / history-service 等独立进程仍在各自 Register*ServiceHTTP 中使用该中间件。
	// 以主包路径作为资源根目录，兼容本地与容器运行环境。
	s.SetServerRoot(gfile.MainPkgPath())
}
