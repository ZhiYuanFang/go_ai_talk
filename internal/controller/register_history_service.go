package controller

import (
	history "hello/internal/services/history"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterHistoryServiceHTTP 注册 history-service 独立进程所需路由。
func RegisterHistoryServiceHTTP(s *ghttp.Server) {
	s.Use(ghttp.MiddlewareHandlerResponse)
	s.Group("/", func(group *ghttp.RouterGroup) {
		inner := NewHistoryCtrl(history.DeviceHistory())
		group.Bind(inner)
		group.Bind(NewHistoryAdminCtrl(inner))
	})
}
