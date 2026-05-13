package gatewayapp

import (
	"net/url"
	"strings"
)

// 联调阶段允许通过 CORS 回显 Origin 的主机（任意端口、http/https），与 openspec change gateway-app-cors-ip-allowlist 对齐。
var gatewayAppCORSAllowedHosts = map[string]struct{}{
	"192.168.0.131":   {},
	"120.55.50.105": {},
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
