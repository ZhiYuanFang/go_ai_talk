package controller

import (
	cashctrl "hello/internal/controller/cash"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterCashServiceHTTP 注册 cash-service 路由（商业功能开通 + 小智 MCP；支付回调共用历史 VIP path）。
func RegisterCashServiceHTTP(s *ghttp.Server) {
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Bind(&cashctrl.CashFeatureController{})
		group.Bind(&cashctrl.CashXiaozhiMcpController{})
	})
	// 支付宝 / Apple 回调仍挂历史 /cash/app/api/vip/* 路径，供 feature + 小智 MCP 履约。
	cashctrl.RegisterAlipayNotify(s)
	cashctrl.RegisterAppleNotifications(s)
}
