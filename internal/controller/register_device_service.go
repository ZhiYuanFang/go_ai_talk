package controller

import (
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterDeviceServiceHTTP 注册 device-service 独立进程所需路由。
func RegisterDeviceServiceHTTP(s *ghttp.Server) {
	// 独立服务保持与 gateway 一致的响应封装，降低调用方适配成本。
	s.Use(ghttp.MiddlewareHandlerResponse)
	admin := NewAdminCtrl(device.DeviceAdmin())
	// 事件 add/update 使用 multipart（非 JSON 绑定）；与 api 文档路径一致。
	s.BindHandler("/device/admin/api/event/add", func(r *ghttp.Request) {
		deviceAdminEventAdd(r, admin)
	})
	s.BindHandler("/device/admin/api/event/update", func(r *ghttp.Request) {
		deviceAdminEventUpdate(r, admin)
	})
	s.Group("/", func(group *ghttp.RouterGroup) {
		// 设备域服务仅承载设备管理接口，避免职责回流到 gateway。
		group.Bind(admin)
		group.Bind(NewDeviceAppUserCtrl())
		group.Bind(NewDeviceAppFeedbackCtrl())
		group.Bind(NewDeviceAppAIQuotaCtrl())
		group.Bind(NewDeviceInternalCtrl())
		group.Bind(NewDeviceInternalProjectionCtrl())
	})
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(deviceUcgInternalSecretMiddleware)
		group.Bind(NewDeviceUcgInternalCtrl())
		group.Bind(NewDeviceAIQuotaInternalCtrl())
	})
}

