package gatewayappctrl

import (
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
)

const simServiceBaseURLEnv = "SIM_SERVICE_BASE_URL"

var (
	simProxyOnce sync.Once
	simProxy     *httputil.ReverseProxy
)

func installSimProxyMiddleware(s *ghttp.Server) {
	proxy := simProxyFromEnv()
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
	s.BindMiddleware("/sim/admin/api/*", serve)
}

func simProxyFromEnv() *httputil.ReverseProxy {
	simProxyOnce.Do(func() {
		target := strings.TrimSpace(os.Getenv(simServiceBaseURLEnv))
		simProxy = buildReverseProxy(target)
	})
	return simProxy
}
