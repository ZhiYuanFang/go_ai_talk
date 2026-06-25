package controller

import "github.com/gogf/gf/v2/net/ghttp"

// RegisterSimUserServiceHTTP 注册 sim-user-service 路由。
func RegisterSimUserServiceHTTP(s *ghttp.Server) {
	s.Use(ghttp.MiddlewareHandlerResponse)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Bind(NewSimAdminCtrl(), NewSimAdminLLMLanesCtrl())
	})
}
