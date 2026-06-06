package controller

import (
	"net/http"
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
		// GoFrame 将更具体路径的中间件排在 CORS(/*) 之前；若把 OPTIONS 反代到 ucg-service，
		// 下游仅注册 POST /media/upload 会返回 405，浏览器预检失败。预检由 gateway CORS 中间件 204 短路。
		if r.Method == http.MethodOptions {
			r.Middleware.Next()
			return
		}
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
