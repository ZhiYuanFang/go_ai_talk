package gatewayapp

import (
	"net/http"
	"net/url"
	"strings"
)

// App 网关 CORS 固定响应头（与浏览器预检及带 Authorization 的 JSON 请求对齐）。
const (
	GatewayAppCORSAllowMethods = "GET, POST, OPTIONS"
	GatewayAppCORSAllowHeaders = "Content-Type, Authorization"
	GatewayAppCORSMaxAge       = "86400"
)

// 联调阶段允许通过 CORS 回显 Origin 的主机（任意端口、http/https），与 openspec change gateway-app-cors-ip-allowlist 对齐。
// localhost / 127.0.0.1：Flutter Web、Vite 等本地开发页跨域请求线上/局域网网关时使用。
var gatewayAppCORSAllowedHosts = map[string]struct{}{
	"www.pangbao.cuplay.top": {},
	"pangbao.cuplay.top":     {},
	"localhost":              {},
	"127.0.0.1":              {},
}

// ReflectGatewayAppCORSOrigin 解析并校验浏览器发来的 Origin 头。
// 若 scheme 为 http/https 且 hostname 属于白名单，则返回原始 Origin 字符串供写入 Access-Control-Allow-Origin；否则 ok 为 false。
// 回显使用原始 header 值，避免 url 重序列化改变大小写或缺省端口表示。
func ReflectGatewayAppCORSOrigin(originHeader string) (echo string, ok bool) {
	raw := strings.TrimSpace(originHeader)
	if raw == "" {
		return "", false
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", false
	}
	host := u.Hostname()
	if host == "" {
		return "", false
	}
	if _, allowed := gatewayAppCORSAllowedHosts[host]; !allowed {
		return "", false
	}
	return raw, true
}

// SetGatewayAppCORSHeaders 将 CORS 头写入 dst（调用方已校验 allowOrigin 合法）。
func SetGatewayAppCORSHeaders(dst http.Header, allowOrigin string) {
	dst.Set("Access-Control-Allow-Origin", allowOrigin)
	dst.Set("Access-Control-Allow-Methods", GatewayAppCORSAllowMethods)
	dst.Set("Access-Control-Allow-Headers", GatewayAppCORSAllowHeaders)
	dst.Set("Access-Control-Max-Age", GatewayAppCORSMaxAge)
}

// ApplyGatewayAppCORSHeaders 根据浏览器 Origin 头选择性写入 CORS：仅白名单命中时写入并返回 true。
// 用于 HTTP 中间件与 httputil.ReverseProxy.ModifyResponse 等共享同一语义。
func ApplyGatewayAppCORSHeaders(dst http.Header, originHeader string) bool {
	echo, ok := ReflectGatewayAppCORSOrigin(originHeader)
	if !ok {
		return false
	}
	SetGatewayAppCORSHeaders(dst, echo)
	return true
}
