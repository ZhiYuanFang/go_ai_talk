package controller

import (
	"context"
	"strings"

	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/net/ghttp"
)

func installGatewayAppEventImageProxy(s *ghttp.Server) {
	target := strings.TrimRight(strings.TrimSpace(gatewayapp.DeviceServiceBaseURL(context.Background())), "/")
	if target == "" {
		target = "http://127.0.0.1:9803"
	}
	installEventImageProxy(s, target)
}
