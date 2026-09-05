package controller

import (
	notifyctrl "hello/internal/controller/notify"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterNotifyServiceHTTP 注册 notify-service 路由。
func RegisterNotifyServiceHTTP(s *ghttp.Server) {
	notifyctrl.RegisterHTTP(s)
}
