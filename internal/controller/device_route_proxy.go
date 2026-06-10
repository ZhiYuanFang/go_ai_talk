package controller

import (
	"net/http/httputil"
	"strings"
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
	// 不得使用 /device/app/api/* 总前缀：会与网关本机路由（token、version 等）冲突。
	// App 设备域 API 按子路径显式登记反代（user、feedback 等），与聚合登录 /device/app/api/login 区分。
	// App 网关下 Bearer 与 X-Internal-Wx-Id / X-Internal-Device-No 由 installGatewayAppBearerMiddleware（HookBeforeServe）统一处理，先于本反代执行，此处不再重复鉴权。
	serve := func(r *ghttp.Request) {
		// 使用统计读 API 由 gateway-app 本机处理（Redis 在边缘），不得反代至 device-service。
		if strings.HasPrefix(r.URL.Path, "/device/admin/api/usage/") {
			r.Middleware.Next()
			return
		}
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
		"/device/app/api/feedback/*",
		"/device/app/api/ai-quota",
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
