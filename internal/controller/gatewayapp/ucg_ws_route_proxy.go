package gatewayappctrl

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	ucgWSRouteModeEnv = "UCG_WS_ROUTE_MODE" // local | proxy
	ucgWSProxyURLEnv  = "UCG_WS_PROXY_URL"
)

var ucgWSProxyPath = "/ucg/app/ws/chat"

type ucgWSProxyConfig struct {
	mode      string
	targetURL string
}

var (
	ucgWSProxyOnce sync.Once
	ucgWSProxyCfg  ucgWSProxyConfig
	ucgWSProxy     *httputil.ReverseProxy
)

// installUcgWSProxyMiddleware 在 gateway-app 挂载 UCG 聊天 WS 升级代理（对外 /ucg/app/ws/chat → 内部 /ws/chat）。
func installUcgWSProxyMiddleware(s *ghttp.Server) {
	cfg, proxy := ucgWSProxyFromEnv()
	handler := func(r *ghttp.Request) {
		if cfg.mode != domainRouteModeProxy {
			writeWSProxyConfigError(r, http.StatusServiceUnavailable, "ucg ws route mode is not proxy")
			r.ExitAll()
			return
		}
		if strings.TrimSpace(cfg.targetURL) == "" || proxy == nil {
			writeUcgWSProxyConfigError(r, http.StatusBadGateway, "ucg ws proxy target is invalid")
			r.ExitAll()
			return
		}
		proxy.ServeHTTP(r.Response.Writer, r.Request)
		r.ExitAll()
	}
	s.BindMiddleware(ucgWSProxyPath, handler)
}

func ucgWSProxyFromEnv() (ucgWSProxyConfig, *httputil.ReverseProxy) {
	ucgWSProxyOnce.Do(func() {
		mode := strings.ToLower(strings.TrimSpace(os.Getenv(ucgWSRouteModeEnv)))
		if mode != domainRouteModeProxy {
			mode = domainRouteModeLocal
		}
		target := strings.TrimSpace(os.Getenv(ucgWSProxyURLEnv))
		ucgWSProxyCfg = ucgWSProxyConfig{
			mode:      mode,
			targetURL: target,
		}
		ucgWSProxy = buildUcgWSReverseProxy(target)
	})
	return ucgWSProxyCfg, ucgWSProxy
}

func buildUcgWSReverseProxy(target string) *httputil.ReverseProxy {
	if strings.TrimSpace(target) == "" {
		return nil
	}
	u, err := url.Parse(target)
	if err != nil {
		return nil
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	ucgChatWSUpgradeProxyDirector(proxy)
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, proxyErr error) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadGateway)
		_, _ = rw.Write([]byte(`{"type":"error","code":1,"stage":"ws_proxy","message":"ucg ws proxy failed"}`))
	}
	return proxy
}

func writeUcgWSProxyConfigError(r *ghttp.Request, status int, message string) {
	r.Response.Status = status
	r.Response.WriteJson(map[string]interface{}{
		"type":    "error",
		"code":    1,
		"stage":   "ws_proxy",
		"message": message,
	})
}

// ucgChatWSUpgradeProxyDirector 将 gateway 对外路径 /ucg/app/ws/chat 改写为 ucg-service 内部 /ws/chat。
func ucgChatWSUpgradeProxyDirector(proxy *httputil.ReverseProxy) {
	if proxy == nil {
		return
	}
	orig := proxy.Director
	proxy.Director = func(req *http.Request) {
		if orig != nil {
			orig(req)
		}
		req.URL.Path = "/ws/chat"
		req.URL.RawPath = "/ws/chat"
	}
}
