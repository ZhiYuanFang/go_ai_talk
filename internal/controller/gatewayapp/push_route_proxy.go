package gatewayappctrl

import (
	"net/http/httputil"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	pushRouteModeEnv          = "PUSH_API_ROUTE_MODE"
	pushProxyURLEnv           = "PUSH_API_PROXY_URL"
	pushProxyCanaryPercentEnv = "PUSH_API_PROXY_CANARY_PERCENT"
)

var (
	pushProxyOnce sync.Once
	pushProxyCfg  domainRouteProxyConfig
	pushProxy     *httputil.ReverseProxy
)

func installPushProxyMiddleware(s *ghttp.Server) {
	cfg, proxy := pushProxyFromEnv()
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
	// 全局推送注册/注销；宿主 push-service。
	s.BindMiddleware("/app/api/push/*", serve)
	// 运维推送设备 Admin API；宿主 push-service。
	s.BindMiddleware("/push/admin/api/*", serve)
}

func pushProxyFromEnv() (domainRouteProxyConfig, *httputil.ReverseProxy) {
	pushProxyOnce.Do(func() {
		pushProxyCfg = readDomainProxyConfig(pushRouteModeEnv, pushProxyURLEnv, pushProxyCanaryPercentEnv)
		pushProxy = buildReverseProxy(pushProxyCfg.targetURL)
	})
	return pushProxyCfg, pushProxy
}
