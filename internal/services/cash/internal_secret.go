package cash

import (
	"os"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	// HeaderInternalSecret 与 device 网关内部密钥头一致，减少配置爆炸。
	HeaderInternalSecret       = "X-Device-Gateway-Internal-Secret"
	headerInternalSecretLegacy = "X-Gateway-Internal-Secret"
)

// ValidateInternalSecret 校验内部调用密钥（复用 DEVICE_GATEWAY_INTERNAL_SECRET）。
func ValidateInternalSecret(headerVal string) bool {
	expected := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if expected == "" {
		return false
	}
	return strings.TrimSpace(headerVal) == expected
}

// InternalSecretFromRequest 读取内部密钥头。
func InternalSecretFromRequest(r *ghttp.Request) string {
	if r == nil {
		return ""
	}
	if v := strings.TrimSpace(r.GetHeader(HeaderInternalSecret)); v != "" {
		return v
	}
	return strings.TrimSpace(r.GetHeader(headerInternalSecretLegacy))
}
