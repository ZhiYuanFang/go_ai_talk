package controller

import (
	"context"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"

	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/net/ghttp"
)

var (
	eventImageProxyOnce sync.Once
	eventImageProxy     *httputil.ReverseProxy
)

func installGatewayAppEventImageProxy(s *ghttp.Server) {
	eventImageProxyOnce.Do(func() {
		target := strings.TrimRight(strings.TrimSpace(gatewayapp.DeviceServiceBaseURL(context.Background())), "/")
		if target == "" {
			target = "http://127.0.0.1:9803"
		}
		eventImageProxy = buildReverseProxy(target)
	})
	s.BindHandler("/ai_talk_images/*", gatewayAppEventImageProxy)
}

// gatewayAppEventImageProxy 将 App 网关上的 /ai_talk_images 反代至 device-service 静态读。
func gatewayAppEventImageProxy(r *ghttp.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	if eventImageProxy == nil {
		r.Response.WriteStatusExit(http.StatusBadGateway)
		return
	}
	eventImageProxy.ServeHTTP(r.Response.Writer, r.Request)
}
