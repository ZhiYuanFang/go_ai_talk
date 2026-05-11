package controller

import (
	"net/http/httputil"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	deviceRouteModeEnv          = "DEVICE_API_ROUTE_MODE" // 本地执行 | 全量代理 | 金丝雀代理
	deviceProxyURLEnv           = "DEVICE_API_PROXY_URL"
	deviceProxyCanaryPercentEnv = "DEVICE_API_PROXY_CANARY_PERCENT"
)

var (
	deviceProxyOnce sync.Once
	deviceProxyCfg  domainRouteProxyConfig
	deviceProxy     *httputil.ReverseProxy
)

func installDeviceProxyMiddleware(s *ghttp.Server) {
	cfg, proxy := deviceProxyFromEnv()
	if proxy == nil {
		return
	}
	// gateway-app-server 等边缘进程不挂载 DeviceAppUserCtrl，依赖本中间件把请求透传到 device-service。
	// 不得使用 /device/app/api/* 总前缀：会与网关本机路由（token、version 等）冲突；App 用户域统一为 /device/app/api/user/*（与聚合登录 /device/app/api/login 区分）。
	serve := func(r *ghttp.Request) {
		if !shouldProxyDomainRequest(cfg, routeKeyForDomainRequest(r)) {
			r.Middleware.Next()
			return
		}
		proxy.ServeHTTP(r.Response.Writer, r.Request)
		r.ExitAll()
	}
	for _, pattern := range []string{
		"/device/admin/api/*",
		"/device/app/api/user/*",
	} {
		s.BindMiddleware(pattern, serve)
	}
}

func deviceProxyFromEnv() (domainRouteProxyConfig, *httputil.ReverseProxy) {
	deviceProxyOnce.Do(func() {
		deviceProxyCfg = readDomainProxyConfig(deviceRouteModeEnv, deviceProxyURLEnv, deviceProxyCanaryPercentEnv)
		deviceProxy = buildReverseProxy(deviceProxyCfg.targetURL)
	})
	return deviceProxyCfg, deviceProxy
}
