package controller

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
	// VOICE_WS_ROUTE_MODE: local 表示不走代理，proxy 表示由 gateway 透传到 voice-service。
	voiceWSRouteModeEnv = "VOICE_WS_ROUTE_MODE" // local | proxy
	// VOICE_WS_PROXY_URL: voice WS 代理目标地址（例如 ws/wss 对应的后端入口）。
	voiceWSProxyURLEnv  = "VOICE_WS_PROXY_URL"
)

// voiceWSProxyConfig 定义 WS 透传所需的最小配置集。
type voiceWSProxyConfig struct {
	// mode 控制是否启用代理；非 proxy 一律按本地模式处理。
	mode      string
	// targetURL 为下游 voice-service 的 WS 入口地址。
	targetURL string
}

var (
	// 通过 once 保证配置与代理对象在进程内只初始化一次，避免运行期抖动。
	voiceWSProxyOnce sync.Once
	voiceWSProxyCfg  voiceWSProxyConfig
	voiceWSProxy     *httputil.ReverseProxy
)

// installVoiceWSProxyMiddleware 在网关挂载 voice WS 入口代理中间件。
// 说明：该路由只在 proxy 模式生效，配置非法时返回结构化错误并终止链路。
func installVoiceWSProxyMiddleware(s *ghttp.Server) {
	cfg, proxy := voiceWSProxyFromEnv()
	s.BindMiddleware("/voice/chat/ws", func(r *ghttp.Request) {
		// 非 proxy 模式直接拒绝，避免误将本地模式请求落入不完整链路。
		if cfg.mode != domainRouteModeProxy {
			writeWSProxyConfigError(r, http.StatusServiceUnavailable, "voice ws route mode is not proxy")
			r.ExitAll()
			return
		}
		// 目标为空或代理构建失败时返回明确错误，避免静默降级。
		if strings.TrimSpace(cfg.targetURL) == "" || proxy == nil {
			writeWSProxyConfigError(r, http.StatusBadGateway, "voice ws proxy target is invalid")
			r.ExitAll()
			return
		}
		// 命中代理后直接转发并结束后续中间件，防止重复处理。
		proxy.ServeHTTP(r.Response.Writer, r.Request)
		r.ExitAll()
	})
}

// voiceWSProxyFromEnv 从环境变量读取配置并惰性构建代理对象。
func voiceWSProxyFromEnv() (voiceWSProxyConfig, *httputil.ReverseProxy) {
	voiceWSProxyOnce.Do(func() {
		// 默认回退到 local，只有显式 proxy 才启用透传，避免因误配导致流量外流。
		mode := strings.ToLower(strings.TrimSpace(os.Getenv(voiceWSRouteModeEnv)))
		if mode != domainRouteModeProxy {
			mode = domainRouteModeLocal
		}
		target := strings.TrimSpace(os.Getenv(voiceWSProxyURLEnv))
		voiceWSProxyCfg = voiceWSProxyConfig{
			mode:      mode,
			targetURL: target,
		}
		voiceWSProxy = buildWSReverseProxy(target)
	})
	return voiceWSProxyCfg, voiceWSProxy
}

// buildWSReverseProxy 构建 WS 反向代理，失败时返回 nil 让上层走显式错误分支。
func buildWSReverseProxy(target string) *httputil.ReverseProxy {
	if strings.TrimSpace(target) == "" {
		return nil
	}
	u, err := url.Parse(target)
	if err != nil {
		return nil
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, proxyErr error) {
		// 下游握手/转发失败时返回统一 JSON 错误，便于前端与日志系统识别阶段。
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadGateway)
		_, _ = rw.Write([]byte(`{"type":"error","code":1,"stage":"ws_proxy","message":"voice ws proxy failed"}`))
	}
	return proxy
}

// writeWSProxyConfigError 统一输出 WS 代理配置/模式错误响应结构。
func writeWSProxyConfigError(r *ghttp.Request, status int, message string) {
	r.Response.Status = status
	r.Response.WriteJson(map[string]interface{}{
		"type":    "error",
		"code":    1,
		"stage":   "ws_proxy",
		"message": message,
	})
}
