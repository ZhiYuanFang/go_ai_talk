package controller

import (
	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterCashServiceHTTP 注册 cash-service 路由（VIP + 商业功能开通）。
func RegisterCashServiceHTTP(s *ghttp.Server) {
	s.Group("/", func(group *ghttp.RouterGroup) {
		// 仅业务 JSON API 包 envelope；支付宝 notify 在组外 BindHandler 返回纯文本。
		group.Middleware(ghttp.MiddlewareHandlerResponse)
		group.Bind(&CashVipController{})
		group.Bind(&CashFeatureController{})
	})
	registerCashAlipayNotify(s)
}
