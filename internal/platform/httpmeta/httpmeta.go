// Package httpmeta 跨进程共用的 HTTP 头常量与轻量解析（无业务域依赖）。
//
// 业务：网关注入头、内部密钥头、wxId 解析、常量时间比较；供各服务与 controller 共用，
// 避免 voice/device/gatewayapp 互相 import 仅为取常量。
package httpmeta

import (
	"crypto/subtle"
	"os"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	// HeaderInternalWxId 下游识别 wx 行（主键），由网关注入。
	HeaderInternalWxId = "X-Internal-Wx-Id"
	// HeaderInternalDeviceNo 下游识别当前会话绑定设备号，由网关注入。
	HeaderInternalDeviceNo = "X-Internal-Device-No"
	// HeaderInternalClientIP 网关解析的真实客户端 IP（客户端不可伪造）。
	HeaderInternalClientIP = "X-Internal-Client-IP"

	// HeaderDeviceGatewayInternalSecret 服务间内部调用共享密钥头。
	HeaderDeviceGatewayInternalSecret = "X-Device-Gateway-Internal-Secret"
	// HeaderGatewayInternalSecretLegacy 历史内部密钥头，与新头二选一接受。
	HeaderGatewayInternalSecretLegacy = "X-Gateway-Internal-Secret"
)

// ParseHeaderWxID 解析 X-Internal-Wx-Id；非法或空返回 0。
func ParseHeaderWxID(header string) int64 {
	s := strings.TrimSpace(header)
	if s == "" {
		return 0
	}
	wxID, err := strconv.ParseInt(s, 10, 64)
	if err != nil || wxID <= 0 {
		return 0
	}
	return wxID
}

// RequireHeaderWxID 解析 X-Internal-Wx-Id；缺失或非法返回 error 文案（供 App controller）。
func RequireHeaderWxID(header string) (int64, string) {
	s := strings.TrimSpace(header)
	if s == "" {
		return 0, "缺少 X-Internal-Wx-Id"
	}
	wxID, err := strconv.ParseInt(s, 10, 64)
	if err != nil || wxID <= 0 {
		return 0, "X-Internal-Wx-Id 无效"
	}
	return wxID, ""
}

// ConstantTimeEqual 常量时间字符串比较，用于口令校验。
func ConstantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		dummy := strings.Repeat("\x00", 64)
		_ = subtle.ConstantTimeCompare([]byte(dummy), []byte(dummy))
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// ValidateInternalSecret 校验内部共享密钥是否与 DEVICE_GATEWAY_INTERNAL_SECRET 一致。
func ValidateInternalSecret(headerVal string) bool {
	expected := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if expected == "" {
		return false
	}
	return strings.TrimSpace(headerVal) == expected
}

// InternalSecretFromRequest 从请求读取内部密钥（兼容新旧头名）。
func InternalSecretFromRequest(r *ghttp.Request) string {
	if r == nil {
		return ""
	}
	if v := strings.TrimSpace(r.GetHeader(HeaderDeviceGatewayInternalSecret)); v != "" {
		return v
	}
	return strings.TrimSpace(r.GetHeader(HeaderGatewayInternalSecretLegacy))
}

// InternalSecretFromHeaderMap 从 header map 读取内部密钥（兼容新旧头名）。
func InternalSecretFromHeaderMap(headers map[string]string) string {
	if headers == nil {
		return ""
	}
	for _, key := range []string{HeaderDeviceGatewayInternalSecret, HeaderGatewayInternalSecretLegacy} {
		if v := strings.TrimSpace(headers[key]); v != "" {
			return v
		}
	}
	return ""
}
