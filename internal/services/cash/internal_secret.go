package cash

import (
	"hello/internal/platform/httpmeta"

	"github.com/gogf/gf/v2/net/ghttp"
)

// HeaderInternalSecret 与 platform 内部密钥头一致（兼容旧调用方）。
const HeaderInternalSecret = httpmeta.HeaderDeviceGatewayInternalSecret

// ValidateInternalSecret 校验内部调用密钥（委托 httpmeta）。
func ValidateInternalSecret(headerVal string) bool {
	return httpmeta.ValidateInternalSecret(headerVal)
}

// InternalSecretFromRequest 读取内部密钥头（委托 httpmeta）。
func InternalSecretFromRequest(r *ghttp.Request) string {
	return httpmeta.InternalSecretFromRequest(r)
}
