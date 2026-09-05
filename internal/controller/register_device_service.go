package controller

import (
	devicectrl "hello/internal/controller/device"
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterDeviceServiceHTTP 注册 device-service 独立进程所需路由。
func RegisterDeviceServiceHTTP(s *ghttp.Server) {
	s.Use(ghttp.MiddlewareHandlerResponse)
	admin := devicectrl.NewAdminCtrl(device.DeviceAdmin())
	s.BindHandler("/device/admin/api/event/add", func(r *ghttp.Request) {
		devicectrl.AdminEventAdd(r, admin)
	})
	s.BindHandler("/device/admin/api/event/update", func(r *ghttp.Request) {
		devicectrl.AdminEventUpdate(r, admin)
	})
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Bind(admin)
		group.Bind(devicectrl.NewDeviceAppUserCtrl())
		group.Bind(devicectrl.NewDeviceAppFeedbackCtrl())
		group.Bind(devicectrl.NewDeviceInternalCtrl())
	})
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(devicectrl.InternalSecretMiddleware)
		group.Bind(devicectrl.NewDeviceUcgInternalCtrl())
		group.Bind(devicectrl.NewDeviceSimInternalCtrl())
	})
}
