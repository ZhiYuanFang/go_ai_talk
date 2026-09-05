package gatewayappctrl

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
)

// RegisterHTTP 注册主网关：语音 WS、领域反代；admin 静态页 302 至 App 网关。
func RegisterHTTP(s *ghttp.Server) {
	s.SetFileServerEnabled(true)
	s.SetIndexFolder(true)
	s.BindHandler("/", func(r *ghttp.Request) {
		r.Response.Write("智能语音")
	})

	installGatewayAdminRedirects(s)

	installDomainProxyMiddlewares(s)
	installVoiceWSProxyMiddleware(s)
	installGatewayCrosscuttingMiddlewares(s)
	installDeviceAPIAccessTouchMiddleware(s)

	s.SetServerRoot(gfile.MainPkgPath())
}
