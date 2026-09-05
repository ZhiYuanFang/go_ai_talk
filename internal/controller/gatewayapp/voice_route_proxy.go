package gatewayappctrl

import (
	"net/http/httputil"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	voiceRouteModeEnv          = "VOICE_API_ROUTE_MODE" // 本地执行 | 全量代理 | 金丝雀代理
	voiceProxyURLEnv           = "VOICE_API_PROXY_URL"
	voiceProxyCanaryPercentEnv = "VOICE_API_PROXY_CANARY_PERCENT"
)

var (
	voiceProxyOnce sync.Once
	voiceProxyCfg  domainRouteProxyConfig
	voiceProxy     *httputil.ReverseProxy
)

func installVoiceProxyMiddleware(s *ghttp.Server) {
	cfg, proxy := voiceProxyFromEnv()
	if proxy == nil {
		return
	}
	serve := func(r *ghttp.Request) {
		if !shouldProxyDomainRequest(cfg, routeKeyForDomainRequest(r)) {
			r.Middleware.Next()
			return
		}
		proxy.ServeHTTP(r.Response.Writer, r.Request)
		r.ExitAll()
	}
	// /device/api/care-alert/*：护理留意日缓存（GET daily / DELETE item / POST feedback）；宿主 voice。
	// tip SSE 与 clinic|tip HTTP 飞轮已下线（remove-tip-and-clinic-feedback）。
	for _, pattern := range []string{
		"/voice/text/*",
		"/voice/app/api/*",
		"/voice/admin/api/*",
		"/device/api/care-alert/*",
	} {
		s.BindMiddleware(pattern, serve)
	}
}

func voiceProxyFromEnv() (domainRouteProxyConfig, *httputil.ReverseProxy) {
	voiceProxyOnce.Do(func() {
		voiceProxyCfg = readDomainProxyConfig(voiceRouteModeEnv, voiceProxyURLEnv, voiceProxyCanaryPercentEnv)
		voiceProxy = buildReverseProxy(voiceProxyCfg.targetURL)
	})
	return voiceProxyCfg, voiceProxy
}
