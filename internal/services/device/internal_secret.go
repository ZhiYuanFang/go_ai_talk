package device

import (
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	// HeaderDeviceGatewayInternalSecret ucg/网关调用 device 内部 UCG 接口时携带（与 DEVICE_GATEWAY_INTERNAL_SECRET 对应）。
	HeaderDeviceGatewayInternalSecret = "X-Device-Gateway-Internal-Secret"
	// headerGatewayInternalSecretLegacy 历史网关内部头，与 HeaderDeviceGatewayInternalSecret 二选一接受。
	headerGatewayInternalSecretLegacy = "X-Gateway-Internal-Secret"
)

// GatewayInternalSecretHeaderValue 从请求头读取共享密钥（兼容新旧头名）。
func GatewayInternalSecretHeaderValue(headers map[string]string) string {
	if headers == nil {
		return ""
	}
	for _, key := range []string{HeaderDeviceGatewayInternalSecret, headerGatewayInternalSecretLegacy} {
		if v := strings.TrimSpace(headers[key]); v != "" {
			return v
		}
	}
	return ""
}

// GatewayInternalSecretHeaderFromRequest 从 ghttp.Request 读取共享密钥（兼容新旧头名）。
// 优先读取新头名 X-Device-Gateway-Internal-Secret，为空时回退旧头名 X-Gateway-Internal-Secret。
// 用于统一 device-service 内部接口的鉴权取值，避免各 controller 自行拼头名导致遗漏兼容。
func GatewayInternalSecretHeaderFromRequest(r *ghttp.Request) string {
	if r == nil {
		return ""
	}
	if v := strings.TrimSpace(r.GetHeader(HeaderDeviceGatewayInternalSecret)); v != "" {
		return v
	}
	return strings.TrimSpace(r.GetHeader(headerGatewayInternalSecretLegacy))
}

// ValidateGatewayInternalSecretHeader 校验内部共享密钥是否与环境变量一致。
func ValidateGatewayInternalSecretHeader(headerVal string) bool {
	return ValidateGatewayInternalSecret(headerVal)
}
