package gatewayapp

import (
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
)

const headerAdminPassword = "X-Admin-Password"

// StripSpoofedGatewayHeaders 移除客户端可能伪造的内部头与运维口令头。
func StripSpoofedGatewayHeaders(r *ghttp.Request) {
	if r == nil {
		return
	}
	for _, h := range []string{
		HeaderInternalClientIP,
		HeaderInternalWxId,
		HeaderInternalDeviceNo,
		HeaderInternalAdminVerified,
		headerAdminPassword,
	} {
		r.Header.Del(h)
		if r.Request != nil {
			r.Request.Header.Del(h)
		}
	}
}

// SetInternalHeader 同时写入 GoFrame 与 net/http 请求头，供反代下游读取。
func SetInternalHeader(r *ghttp.Request, key, value string) {
	if r == nil {
		return
	}
	r.Header.Set(key, value)
	if r.Request != nil {
		r.Request.Header.Set(key, value)
	}
}

// MarkAdminJWTVerified 标记当前请求已通过 Admin JWT 校验。
func MarkAdminJWTVerified(r *ghttp.Request) {
	SetInternalHeader(r, HeaderInternalAdminVerified, "1")
}

// InjectAdminDownstreamPassword 按路径为反代下游注入 X-Admin-Password（仅服务端持有 env 值）。
func InjectAdminDownstreamPassword(r *ghttp.Request) {
	if r == nil {
		return
	}
	path := r.URL.Path
	var pwd string
	switch {
	case strings.HasPrefix(path, "/ucg/admin/api/"):
		pwd = UcgAdminPassword()
	default:
		pwd = DeviceAdminPassword()
	}
	if pwd == "" {
		return
	}
	SetInternalHeader(r, headerAdminPassword, pwd)
}

// BearerTokenFromRequest 解析 Authorization Bearer 原始 token。
func BearerTokenFromRequest(r *ghttp.Request) string {
	if r == nil {
		return ""
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	const pfx = "Bearer "
	if len(auth) < len(pfx) || !strings.EqualFold(auth[:len(pfx)], pfx) {
		return ""
	}
	return strings.TrimSpace(auth[len(pfx):])
}

// RequestAdminVerified 当前 HTTP 请求是否已通过 Admin JWT（供 controller 使用）。
func RequestAdminVerified(r *ghttp.Request) bool {
	if r == nil {
		return false
	}
	if IsAdminJWTVerified(r.Header.Get) {
		return true
	}
	if r.Request != nil && IsAdminJWTVerified(r.Request.Header.Get) {
		return true
	}
	return false
}
