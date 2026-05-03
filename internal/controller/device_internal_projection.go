package controller

import (
	"context"

	v1 "hello/api/v1"
	device "hello/internal/services/device"
)

// DeviceInternalProjectionCtrl 设备域内部投影修复（供 worker 调用，依赖网络隔离）。
type DeviceInternalProjectionCtrl struct{}

// NewDeviceInternalProjectionCtrl 构造 DeviceInternalProjectionCtrl。
func NewDeviceInternalProjectionCtrl() *DeviceInternalProjectionCtrl {
	return &DeviceInternalProjectionCtrl{}
}

// Reconcile 按设备列表刷新画像缓存，并全量重建事件/动作缓存。
func (c *DeviceInternalProjectionCtrl) Reconcile(ctx context.Context, req *v1.DeviceInternalProjectionReconcileReq) (res *v1.DeviceInternalProjectionReconcileRes, err error) {
	if err := device.ReconcileProjectionCachesForWorker(ctx, req.DeviceNos); err != nil {
		return nil, err
	}
	return &v1.DeviceInternalProjectionReconcileRes{}, nil
}

// Apply 在 device-service 进程内执行单条投影（避免 worker 直连 device 库）。
func (c *DeviceInternalProjectionCtrl) Apply(ctx context.Context, req *v1.DeviceInternalProjectionApplyReq) (res *v1.DeviceInternalProjectionApplyRes, err error) {
	if err := device.ApplyProjection(ctx, req.RoutingKey, req.Payload); err != nil {
		return nil, err
	}
	return &v1.DeviceInternalProjectionApplyRes{}, nil
}
