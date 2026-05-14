package controller

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
)

// RegisterGatewayAppHTTP 注册 App 专用网关：代理链 + Bearer 注入 + App 自有 API + 历史 WS。
func RegisterGatewayAppHTTP(s *ghttp.Server) {
	s.SetFileServerEnabled(true)
	s.SetIndexFolder(true)
	s.BindHandler("/", func(r *ghttp.Request) {
		r.Response.Write("智能语音 App 网关")
	})
	s.BindHandler("/device/admin", func(r *ghttp.Request) {
		r.Response.ServeFile("resource/public/admin.html")
	})
	s.BindHandler("/device/history/*deviceNo", func(r *ghttp.Request) {
		r.Response.ServeFile("resource/public/history.html")
	})
	// App 联调页：模拟登录、绑定、画像、文本对话、WS、piece 趋势（仅用于验证 gateway-app 能力）。
	s.BindHandler("/device/app/integration-test.html", func(r *ghttp.Request) {
		r.Response.ServeFile("resource/public/gateway-app-integration-test.html")
	})

	// 版本管理：口令登录 + APK 上传 + 匿名 APK 下载（路径在 Bearer 白名单）。
	s.BindHandler("/device/app/version-admin.html", func(r *ghttp.Request) {
		r.Response.ServeFile("resource/public/gateway-app-version-admin.html")
	})
	s.BindHandler("/device/app/api/version/admin/login", gatewayAppVersionAdminLogin)
	s.BindHandler("/device/app/api/version/admin/upload", gatewayAppVersionAdminUpload)
	s.BindHandler("/device/app/apk/*filename", gatewayAppApkDownload)

	// CORS 仅 App 网关：须在跨切面之前注册，使外层在中间件返回后仍可补写反代响应头。
	installGatewayAppCORSMiddleware(s)
	installGatewayCrosscuttingMiddlewares(s)
	installGatewayAppBearerMiddleware(s)
	installDomainProxyMiddlewares(s)
	installVoiceWSProxyMiddleware(s)

	s.BindHandler("/device/app/ws/history", gatewayAppHistoryWS)

	// 必须使用前缀为 "/" 的 Group：`g.Meta` 的 path 已是绝对路径（如 /device/app/api/login），
	// 若再套 Group("/device/app/api") 会拼接成 /device/app/api/device/app/api/login 导致 404。
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Bind(NewGatewayAppCtrl())
	})

	s.SetServerRoot(gfile.MainPkgPath())
}
