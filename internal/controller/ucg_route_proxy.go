package controller

import (
	"net/http/httputil"
	"os"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
)

const ucgServiceBaseURLEnv = "UCG_SERVICE_BASE_URL"

var (
	ucgProxyOnce sync.Once
	ucgProxy     *httputil.ReverseProxy
)

func installUcgProxyMiddleware(s *ghttp.Server) {
	proxy := ucgProxyFromEnv()
	if proxy == nil {
		return
	}
	s.BindMiddleware("/ucg/app/api/*", func(r *ghttp.Request) {
		// Bearer 与 X-Internal-Wx-Id 由 gateway-app HookBeforeServe 统一注入，此处仅透传。
		proxy.ServeHTTP(r.Response.Writer, r.Request)
		r.ExitAll()
	})
}

func ucgProxyFromEnv() *httputil.ReverseProxy {
	ucgProxyOnce.Do(func() {
		target := strings.TrimSpace(os.Getenv(ucgServiceBaseURLEnv))
		ucgProxy = buildReverseProxy(target)
	})
	return ucgProxy
}
