package controller

import (
	"context"

	v1 "hello/api/v1"
	device "hello/internal/services/device"
	"hello/internal/shared/eventlogo"
)

// DeviceInternalCtrl 设备域内部只读接口（供 history/voice 跨服务调用，依赖网络隔离）。
type DeviceInternalCtrl struct{}

// NewDeviceInternalCtrl 构造 DeviceInternalCtrl。
func NewDeviceInternalCtrl() *DeviceInternalCtrl {
	return &DeviceInternalCtrl{}
}

// EventOptions 返回事件字典，等价于管理端事件列表但免口令。
func (c *DeviceInternalCtrl) EventOptions(ctx context.Context, req *v1.DeviceInternalEventOptionsReq) (res *v1.DeviceInternalEventOptionsRes, err error) {
	_ = req
	items, err := device.DeviceAdmin().ListEvents(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DeviceInternalEventOptionsRes{List: eventlogo.MapEventsLogoCdn(ctx, items)}, nil
}
