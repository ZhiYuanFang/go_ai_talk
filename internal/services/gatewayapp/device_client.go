package gatewayapp

import (
	"context"

	deviceclient "hello/internal/clients/device"
)

// DeviceServiceBaseURL 下游 device-service 基址（委托 clients/device）。
func DeviceServiceBaseURL(ctx context.Context) string {
	return deviceclient.DeviceServiceBaseURL(ctx)
}

// FetchDeviceNoByWxID 网关刷新 access 时按 wx 取 device_no（委托 clients/device）。
func FetchDeviceNoByWxID(ctx context.Context, wxID int64) (string, error) {
	return deviceclient.FetchDeviceNoByWxIDViaGClient(ctx, wxID)
}
