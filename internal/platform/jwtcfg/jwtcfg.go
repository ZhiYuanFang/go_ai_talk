package jwtcfg

import (
	"os"
	"strings"
)

// GatewayAppSecret 返回 App access JWT 签名密钥（gateway-app 签发、ucg WS 校验须一致）。
// 仅来自 GATEWAY_APP_JWT_SECRET；prod/test 须在各自 .env 使用不同值以实现令牌环境隔离。
func GatewayAppSecret() string {
	return strings.TrimSpace(os.Getenv("GATEWAY_APP_JWT_SECRET"))
}
