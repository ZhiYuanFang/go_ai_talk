package gatewayappctrl

import (
	"net/http/httputil"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	historyRouteModeEnv          = "HISTORY_API_ROUTE_MODE" // 本地执行 | 全量代理 | 金丝雀代理
	historyProxyURLEnv           = "HISTORY_API_PROXY_URL"
	historyProxyCanaryPercentEnv = "HISTORY_API_PROXY_CANARY_PERCENT"
)

var (
	historyProxyOnce sync.Once
	historyProxyCfg  domainRouteProxyConfig
	historyProxy     *httputil.ReverseProxy
)

func installHistoryProxyMiddleware(s *ghttp.Server) {
	cfg, proxy := historyProxyFromEnv()
	if proxy == nil {
		return
	}
	serveHistoryProxy := func(r *ghttp.Request) {
		// 通过稳定路由键做一致性分流，确保同一设备命中同一发布策略。
		if !shouldProxyHistoryRequest(cfg, routeKeyForHistoryRequest(r)) {
			r.Middleware.Next()
			return
		}
		// 命中代理后直接短路，避免本地 handler 与下游重复处理。
		proxy.ServeHTTP(r.Response.Writer, r.Request)
		r.ExitAll()
	}
	s.BindMiddleware("/device/history/api/*", serveHistoryProxy)
	s.BindMiddleware("/device/admin/api/history/*", serveHistoryProxy)
}

func historyProxyFromEnv() (domainRouteProxyConfig, *httputil.ReverseProxy) {
	historyProxyOnce.Do(func() {
		// 历史路由配置仅初始化一次，保证同一进程内行为稳定。
		historyProxyCfg = readDomainProxyConfig(historyRouteModeEnv, historyProxyURLEnv, historyProxyCanaryPercentEnv)
		historyProxy = buildReverseProxy(historyProxyCfg.targetURL)
	})
	return historyProxyCfg, historyProxy
}

func routeKeyForHistoryRequest(r *ghttp.Request) string {
	// history 分流键与 voice/device 保持同一策略，减少跨域治理差异。
	return routeKeyForDomainRequest(r)
}

func shouldProxyHistoryRequest(cfg domainRouteProxyConfig, key string) bool {
	// history 转发判定复用统一规则，避免三套逻辑发生语义漂移。
	return shouldProxyDomainRequest(cfg, key)
}
