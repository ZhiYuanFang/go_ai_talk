package controller

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
)

// RegisterGatewayAppHTTP 注册 App 专用网关：代理链 + Bearer 注入 + App 自有 API + 历史 WS。
func RegisterGatewayAppHTTP(s *ghttp.Server) {
	s.SetFileServerEnabled(true)
	s.SetIndexFolder(true)
	s.BindHandler("/apple-app-site-association", gatewayAppAppleAppSiteAssociation)
	s.BindHandler("/.well-known/apple-app-site-association", gatewayAppAppleAppSiteAssociation)
	s.BindHandler("/wx/ulink/*", gatewayAppUniversalLinksLanding)
	s.BindHandler("/vendor/qrcode.min.js", func(r *ghttp.Request) {
		r.Response.Header().Set("Cache-Control", "public, max-age=604800")
		r.Response.ServeFile("resource/public/vendor/qrcode.min.js")
	})
	s.BindHandler("/", func(r *ghttp.Request) {
		r.Response.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		r.Response.ServeFile("resource/public/pangbao-home.html")
	})
	s.BindHandler("/user-agreement.html", func(r *ghttp.Request) {
		r.Response.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		r.Response.ServeFile("resource/public/user-agreement.html")
	})
	s.BindHandler("/privacy-policy.html", func(r *ghttp.Request) {
		r.Response.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		r.Response.ServeFile("resource/public/privacy-policy.html")
	})

	RegisterAdminStaticPages(s)

	s.BindHandler("/device/app/integration-test.html", func(r *ghttp.Request) {
		r.Response.ServeFile("resource/public/gateway-app-integration-test.html")
	})

	s.BindHandler("/device/app/api/version/admin/upload", gatewayAppVersionAdminUpload)
	s.BindHandler("/device/app/api/version/admin/list", gatewayAppVersionAdminList)
	s.BindHandler("/device/app/api/version/admin/get", gatewayAppVersionAdminGet)
	s.BindHandler("/device/app/api/version/admin/update", gatewayAppVersionAdminUpdate)
	s.BindHandler("/device/app/api/version/admin/delete", gatewayAppVersionAdminDelete)
	s.BindHandler("/device/app/apk/*filename", gatewayAppApkDownload)

	installGatewayAppCORSMiddleware(s)
	installGatewayCrosscuttingMiddlewares(s)
	installGatewayAppBearerMiddleware(s)
	installDeviceAPIAccessTouchMiddleware(s)
	installGatewayAppAPIUsageStatsMiddleware(s)
	installDomainProxyMiddlewares(s)
	installVoiceWSProxyMiddleware(s)
	installUcgWSProxyMiddleware(s)

	s.BindHandler("/device/app/ws/history", gatewayAppHistoryWS)

	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Bind(NewGatewayAppCtrl())
		group.Bind(NewGatewayAppUsageAdminCtrl())
		group.Bind(NewGatewayAdminLoginCtrl())
	})

	s.SetServerRoot(gfile.MainPkgPath())
}
