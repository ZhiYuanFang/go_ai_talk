package controller

import (
	"hash/fnv"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	domainRouteModeLocal  = "local"
	domainRouteModeProxy  = "proxy"
	domainRouteModeCanary = "canary"
)

type domainRouteProxyConfig struct {
	mode          string
	targetURL     string
	canaryPercent int
}

func installDomainProxyMiddlewares(s *ghttp.Server) {
	// 统一安装领域代理中间件；具体代理实现仍按 history/voice/device/ucg 分文件维护。
	installHistoryProxyMiddleware(s)
	installVoiceProxyMiddleware(s)
	installDeviceProxyMiddleware(s)
	installUcgProxyMiddleware(s)
}

func routeKeyForDomainRequest(r *ghttp.Request) string {
	// 金丝雀/分流键仅用 query 与可信头，避免调用 r.Get("deviceNo") 触发对 POST JSON 的 parseForm/parseBody，
	// 降低与反向代理转发链路的耦合；deviceNo 若仅出现在 body 中则回退到 addr+path 键。
	if deviceNo := strings.TrimSpace(r.GetQuery("deviceNo").String()); deviceNo != "" {
		return deviceNo
	}
	if header := strings.TrimSpace(r.GetHeader("X-Device-No")); header != "" {
		return header
	}
	return strings.TrimSpace(r.RemoteAddr + "|" + r.URL.Path)
}

func readDomainProxyConfig(modeEnv, targetEnv, canaryPercentEnv string) domainRouteProxyConfig {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv(modeEnv)))
	switch mode {
	case domainRouteModeProxy, domainRouteModeCanary:
	default:
		mode = domainRouteModeLocal
	}
	target := strings.TrimSpace(os.Getenv(targetEnv))
	canary := 0
	if canaryPercentEnv != "" {
		canary = parseCanaryPercent(os.Getenv(canaryPercentEnv))
	}
	return domainRouteProxyConfig{
		mode:          mode,
		targetURL:     target,
		canaryPercent: canary,
	}
}

func parseCanaryPercent(v string) int {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return 0
	}
	if n < 0 {
		return 0
	}
	if n > 100 {
		return 100
	}
	return n
}

func shouldProxyDomainRequest(cfg domainRouteProxyConfig, key string) bool {
	if strings.TrimSpace(cfg.targetURL) == "" {
		return false
	}
	switch cfg.mode {
	case domainRouteModeProxy:
		return true
	case domainRouteModeCanary:
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

func buildReverseProxy(target string) *httputil.ReverseProxy {
	if strings.TrimSpace(target) == "" {
		return nil
	}
	u, err := url.Parse(target)
	if err != nil {
		return nil
	}
	p := httputil.NewSingleHostReverseProxy(u)
	// 领域反代在 ServeHTTP 内直接写回客户端，外层 CORS 中间件在 Next 之后补头不可靠；在此处统一合并，
	// history/voice/device 凡经 buildReverseProxy 构建的实例均自动带上与 gateway_app_cors 一致的 CORS 语义。
	p.ModifyResponse = func(resp *http.Response) error {
		if resp == nil {
			return nil
		}
		// RoundTrip 返回的 Response.Request 为发往下游的 outreq（含浏览器原始 Origin 等头），用于白名单判定。
		if resp.Request == nil {
			return nil
		}
		_ = gatewayapp.ApplyGatewayAppCORSHeaders(resp.Header, resp.Request.Header.Get("Origin"))
		return nil
	}
	return p
}
