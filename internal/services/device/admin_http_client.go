package device

import (
	deviceclient "hello/internal/clients/device"
	"hello/internal/services/contracts"
)

// HTTPDeviceAdmin 返回 HTTP 版设备管理契约（委托 clients/device，保留旧符号兼容）。
func HTTPDeviceAdmin() contracts.DeviceAdminContract {
	return deviceclient.HTTPDeviceAdmin()
}
