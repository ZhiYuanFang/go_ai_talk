package device

import "strings"

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

// ValidateGatewayInternalSecretHeader 校验内部共享密钥是否与环境变量一致。
func ValidateGatewayInternalSecretHeader(headerVal string) bool {
	return ValidateGatewayInternalSecret(headerVal)
}
