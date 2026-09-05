package gatewayappctrl

import (
	"hello/internal/services/gatewayapp/apiregistry"
	"hello/internal/services/gatewayapp/usagestats"

	"github.com/gogf/gf/v2/net/ghttp"
)

// installGatewayAppAPIUsageStatsMiddleware 在响应完成后记录 2xx App API 使用统计（网关本机 Handler 路径）。
// 经领域反代的请求在 buildReverseProxy.ModifyResponse 中记录。
func installGatewayAppAPIUsageStatsMiddleware(s *ghttp.Server) {
	apiregistry.Init()
	s.BindMiddleware("/*", func(r *ghttp.Request) {
		r.Middleware.Next()
		usagestats.RecordGHTTPRequest(r)
	})
}
