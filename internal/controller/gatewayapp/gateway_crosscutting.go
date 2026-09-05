package gatewayappctrl

import (
	"fmt"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
)

func installGatewayCrosscuttingMiddlewares(s *ghttp.Server) {
	s.BindMiddleware("/*", func(r *ghttp.Request) {
		// 优先透传上游请求 ID，保证跨服务链路日志可串联。
		requestID := r.GetHeader("X-Request-Id")
		if requestID == "" {
			// 无上游标识时在 gateway 生成，作为全链路兜底追踪 ID。
			requestID = fmt.Sprintf("gw-%d", time.Now().UnixNano())
		}
		// 同时写回响应头和下游请求头，便于前后端与服务间统一追踪。
		r.Response.Header().Set("X-Request-Id", requestID)
		r.Request.Header.Set("X-Request-Id", requestID)
		r.Middleware.Next()
	})
}

