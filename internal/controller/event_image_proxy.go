package controller

import (
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
)

var (
	eventImageProxyMu     sync.Mutex
	eventImageProxyInst   *httputil.ReverseProxy
	eventImageProxyTarget string
)

// installEventImageProxy 将 GET/HEAD /ai_talk_images/* 反代至 device-service 静态读（与管理页同源预览）。
func installEventImageProxy(s *ghttp.Server, target string) {
	target = strings.TrimRight(strings.TrimSpace(target), "/")
	if target == "" {
		target = "http://127.0.0.1:9803"
	}
	eventImageProxyMu.Lock()
	if eventImageProxyInst == nil || eventImageProxyTarget != target {
		eventImageProxyTarget = target
		eventImageProxyInst = buildReverseProxy(target)
	}
	eventImageProxyMu.Unlock()
	s.BindHandler("/ai_talk_images/*", serveEventImageProxy)
}

func serveEventImageProxy(r *ghttp.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	eventImageProxyMu.Lock()
	proxy := eventImageProxyInst
	eventImageProxyMu.Unlock()
	if proxy == nil {
		r.Response.WriteStatusExit(http.StatusBadGateway)
		return
	}
	proxy.ServeHTTP(r.Response.Writer, r.Request)
}

// gatewayDeviceServiceTarget 主网关反代 device 时使用的下游基址（与 DEVICE_API_PROXY_URL 一致）。
func gatewayDeviceServiceTarget() string {
	cfg, _ := deviceProxyFromEnv()
	if t := strings.TrimRight(strings.TrimSpace(cfg.targetURL), "/"); t != "" {
		return t
	}
	return "http://127.0.0.1:9803"
}
