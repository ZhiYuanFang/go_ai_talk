package controller

import (
	pushctrl "hello/internal/controller/push"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterPushServiceHTTP 注册 push-service 路由。
func RegisterPushServiceHTTP(s *ghttp.Server) {
	s.Use(ghttp.MiddlewareHandlerResponse)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Bind(&pushctrl.AppPushCtrl{})
		group.POST("/push/internal/api/by-biz-type", pushctrl.InternalByBizType)
	})
}
