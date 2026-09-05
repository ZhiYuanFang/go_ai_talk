package gatewayappctrl

import (
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
)

const cashServiceBaseURLEnv = "CASH_SERVICE_URL"

var (
	cashProxyOnce sync.Once
	cashProxy     *httputil.ReverseProxy
)

// installCashProxyMiddleware gateway-app 反代 cash-service。
//
// 绑定：
//   - /cash/app/api/*：App VIP 支付/状态（Bearer + X-Internal-Wx-Id 由 Hook 注入）
//   - /cash/admin/api/*：运维 Hub VIP 权益 Admin API（Admin JWT + X-Admin-Password 注入）
func installCashProxyMiddleware(s *ghttp.Server) {
	proxy := cashProxyFromEnv()
	if proxy == nil {
		return
	}
	serve := func(r *ghttp.Request) {
		if r.Method == http.MethodOptions {
			r.Middleware.Next()
			return
		}
		proxy.ServeHTTP(r.Response.Writer, r.Request)
		r.ExitAll()
	}
	s.BindMiddleware("/cash/app/api/*", serve)
	s.BindMiddleware("/cash/admin/api/*", serve)
}

func cashProxyFromEnv() *httputil.ReverseProxy {
	cashProxyOnce.Do(func() {
		target := strings.TrimSpace(os.Getenv(cashServiceBaseURLEnv))
		cashProxy = buildReverseProxy(target)
	})
	return cashProxy
}
