package voice

import (
	contracts "hello/internal/services/contracts"
	device "hello/internal/services/device"
	history "hello/internal/services/history"
)

// DeviceAdmin 返回设备管理契约实现，供语音领域编排复用。
func DeviceAdmin() contracts.DeviceAdminContract {
	return device.DeviceAdmin()
}

// DeviceProfile 返回设备画像契约实现，供语音推理补充画像信息。
func DeviceProfile() device.DeviceProfileContract {
	return device.DeviceProfile()
}

// DeviceHistory 返回历史契约实现，供语音意图写入/查询历史。
func DeviceHistory() history.Contract {
	return history.DeviceHistory()
}
