package controller

import (
	"hash/fnv"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	historyRouteModeEnv         = "HISTORY_API_ROUTE_MODE" // local | proxy | canary
	historyProxyURLEnv          = "HISTORY_API_PROXY_URL"
	historyProxyCanaryPercentEnv = "HISTORY_API_PROXY_CANARY_PERCENT"

	historyRouteModeLocal  = "local"
	historyRouteModeProxy  = "proxy"
	historyRouteModeCanary = "canary"
)

type historyRouteProxyConfig struct {
	mode          string
	targetURL     string
	canaryPercent int
}

var (
	historyProxyOnce sync.Once
	historyProxyCfg  historyRouteProxyConfig
	historyProxy     *httputil.ReverseProxy
)

func installHistoryProxyMiddleware(s *ghttp.Server) {
	cfg, proxy := historyProxyFromEnv()
	if proxy == nil {
		return
	}
	s.BindMiddleware("/device/history/api/*", func(r *ghttp.Request) {
		if !shouldProxyHistoryRequest(cfg, routeKeyForHistoryRequest(r)) {
			r.Middleware.Next()
			return
		}
		proxy.ServeHTTP(r.Response.Writer, r.Request)
		r.ExitAll()
	})
}

func routeKeyForHistoryRequest(r *ghttp.Request) string {
	deviceNo := strings.TrimSpace(r.Get("deviceNo").String())
	if deviceNo != "" {
		return deviceNo
	}
	if header := strings.TrimSpace(r.GetHeader("X-Device-No")); header != "" {
		return header
	}
	return strings.TrimSpace(r.RemoteAddr + "|" + r.URL.Path)
}

func historyProxyFromEnv() (historyRouteProxyConfig, *httputil.ReverseProxy) {
	historyProxyOnce.Do(func() {
		mode := strings.ToLower(strings.TrimSpace(os.Getenv(historyRouteModeEnv)))
		switch mode {
		case historyRouteModeProxy, historyRouteModeCanary:
		default:
			mode = historyRouteModeLocal
		}
		target := strings.TrimSpace(os.Getenv(historyProxyURLEnv))
		canary, err := strconv.Atoi(strings.TrimSpace(os.Getenv(historyProxyCanaryPercentEnv)))
		if err != nil {
			canary = 0
		}
		if canary < 0 {
			canary = 0
		}
		if canary > 100 {
			canary = 100
		}
		historyProxyCfg = historyRouteProxyConfig{
			mode:          mode,
			targetURL:     target,
			canaryPercent: canary,
		}
		if target == "" {
			return
		}
		u, err := url.Parse(target)
		if err != nil {
			return
		}
		historyProxy = httputil.NewSingleHostReverseProxy(u)
	})
	return historyProxyCfg, historyProxy
}

func shouldProxyHistoryRequest(cfg historyRouteProxyConfig, key string) bool {
	if strings.TrimSpace(cfg.targetURL) == "" {
		return false
	}
	switch cfg.mode {
	case historyRouteModeProxy:
		return true
	case historyRouteModeCanary:
		if cfg.canaryPercent <= 0 {
			return false
		}
		if cfg.canaryPercent >= 100 {
			return true
		}
		h := fnv.New32a()
		_, _ = h.Write([]byte(strings.TrimSpace(key)))
		return int(h.Sum32()%100) < cfg.canaryPercent
	default:
		return false
	}
}
