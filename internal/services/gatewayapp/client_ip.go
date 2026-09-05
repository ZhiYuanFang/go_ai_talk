package gatewayapp

import (
	"net"
	"strings"

	"hello/internal/platform/httpmeta"

	"github.com/gogf/gf/v2/net/ghttp"
)

// HeaderInternalClientIP 网关解析的真实客户端 IP（别名 httpmeta）。
const HeaderInternalClientIP = httpmeta.HeaderInternalClientIP

// internalHeaders 仅允许网关注入、须剥离客户端伪造值的内部头。
var internalHeaders = []string{
	HeaderInternalClientIP,
	HeaderInternalWxId,
	HeaderInternalDeviceNo,
}

// StripSpoofedInternalHeaders 移除客户端可能伪造的内部头，避免污染反代下游。
func StripSpoofedInternalHeaders(r *ghttp.Request) {
	StripSpoofedGatewayHeaders(r)
}

// InjectClientIPHeader 从 X-Forwarded-For / RemoteAddr 解析客户端 IP 并写入 HeaderInternalClientIP。
func InjectClientIPHeader(r *ghttp.Request) {
	if r == nil {
		return
	}
	ip := ClientIP(r)
	if ip == "" {
		return
	}
	r.Header.Set(HeaderInternalClientIP, ip)
	if r.Request != nil {
		r.Request.Header.Set(HeaderInternalClientIP, ip)
	}
}

// ClientIP 解析请求来源 IP：优先 X-Forwarded-For 首跳，否则 RemoteAddr（去端口）。
func ClientIP(r *ghttp.Request) string {
	if r == nil {
		return ""
	}
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	host := strings.TrimSpace(r.RemoteAddr)
	if host == "" {
		return ""
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		return strings.TrimSpace(h)
	}
	return host
}
