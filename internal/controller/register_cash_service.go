package controller

import (
	cashctrl "hello/internal/controller/cash"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterCashServiceHTTP 注册 cash-service 路由（VIP + 商业功能开通）。
func RegisterCashServiceHTTP(s *ghttp.Server) {
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Bind(&cashctrl.CashVipController{})
		group.Bind(&cashctrl.CashFeatureController{})
	})
	cashctrl.RegisterAlipayNotify(s)
}
