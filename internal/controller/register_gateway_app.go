package controller

import (
	gatewayappctrl "hello/internal/controller/gatewayapp"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterGatewayAppHTTP 注册 App 专用网关。
func RegisterGatewayAppHTTP(s *ghttp.Server) {
	gatewayappctrl.RegisterGatewayAppHTTP(s)
}

// RegisterHTTP 注册主网关。
func RegisterHTTP(s *ghttp.Server) {
	gatewayappctrl.RegisterHTTP(s)
}
