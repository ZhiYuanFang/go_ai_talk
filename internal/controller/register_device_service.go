package controller

import (
	device "hello/internal/services/device"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterDeviceServiceHTTP 注册 device-service 独立进程所需路由。
func RegisterDeviceServiceHTTP(s *ghttp.Server) {
	// 独立服务保持与 gateway 一致的响应封装，降低调用方适配成本。
	s.Use(ghttp.MiddlewareHandlerResponse)
	s.Group("/", func(group *ghttp.RouterGroup) {
		// 设备域服务仅承载设备管理接口，避免职责回流到 gateway。
		group.Bind(NewAdminCtrl(device.DeviceAdmin()))
		group.Bind(NewDeviceProfileCtrl())
	})
}

