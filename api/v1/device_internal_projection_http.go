package v1

import "github.com/gogf/gf/v2/frame/g"

// DeviceInternalProjectionReconcileReq worker 触发的设备域缓存投影修复。
type DeviceInternalProjectionReconcileReq struct {
	g.Meta    `path:"/device/internal/api/projection/reconcile" method:"post" tags:"device" summary:"内部-投影缓存修复"`
	DeviceNos []string `json:"deviceNos" dc:"需刷新画像缓存的设备号列表"`
}

// DeviceInternalProjectionReconcileRes 占位响应。
type DeviceInternalProjectionReconcileRes struct{}

// DeviceInternalProjectionApplyReq 应用单条 device 域缓存投影（供 worker outbox 中继调用）。
type DeviceInternalProjectionApplyReq struct {
	g.Meta     `path:"/device/internal/api/projection/apply" method:"post" tags:"device" summary:"内部-应用投影事件"`
	RoutingKey string `json:"routingKey" dc:"路由键"`
	Payload    string `json:"payload" dc:"JSON 负载原文"`
}

// DeviceInternalProjectionApplyRes 占位响应。
type DeviceInternalProjectionApplyRes struct{}
