package voice

import (
	deviceclient "hello/internal/clients/device"
	historyclient "hello/internal/clients/history"
	contracts "hello/internal/services/contracts"
)

// DeviceAdmin 返回设备管理契约实现；voice-service 仅经 HTTP 访问 device 域，禁止进程内直连他域 DAO。
func DeviceAdmin() contracts.DeviceAdminContract {
	return deviceclient.HTTPDeviceAdmin()
}

// DeviceProfile 返回设备画像契约实现，供语音推理补充画像信息。
func DeviceProfile() deviceclient.DeviceProfileContract {
	return deviceclient.HTTPDeviceProfile()
}

// DeviceHistory 返回历史契约实现，供语音意图写入/查询历史。
func DeviceHistory() contracts.DeviceHistoryContract {
	return historyclient.HTTPDeviceHistory()
}
