package device

import (
	"hello/internal/platform/httpmeta"

	"github.com/gogf/gf/v2/net/ghttp"
)

// HeaderDeviceGatewayInternalSecret 服务间内部调用共享密钥头（别名 httpmeta）。
const HeaderDeviceGatewayInternalSecret = httpmeta.HeaderDeviceGatewayInternalSecret

// GatewayInternalSecretHeaderValue 从 header map 读取共享密钥（委托 httpmeta）。
func GatewayInternalSecretHeaderValue(headers map[string]string) string {
	return httpmeta.InternalSecretFromHeaderMap(headers)
}

// GatewayInternalSecretHeaderFromRequest 从 ghttp.Request 读取共享密钥（委托 httpmeta）。
func GatewayInternalSecretHeaderFromRequest(r *ghttp.Request) string {
	return httpmeta.InternalSecretFromRequest(r)
}

// ValidateGatewayInternalSecretHeader 校验内部共享密钥（委托 httpmeta）。
func ValidateGatewayInternalSecretHeader(headerVal string) bool {
	return httpmeta.ValidateInternalSecret(headerVal)
}
