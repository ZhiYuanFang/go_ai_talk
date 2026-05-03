package v1

import "github.com/gogf/gf/v2/frame/g"

// HistoryInternalProjectionReconcileReq worker 触发的读模型投影修复（仅 history 进程内访问本库）。
type HistoryInternalProjectionReconcileReq struct {
	g.Meta `path:"/history/internal/api/projection/reconcile" method:"post" tags:"history" summary:"内部-投影缓存修复"`
	Limit  int `json:"limit" dc:"抽样设备数量上限"`
}

// HistoryInternalProjectionReconcileRes 返回本次修复涉及的设备号，供 worker 继续调用 device。
type HistoryInternalProjectionReconcileRes struct {
	DeviceNos []string `json:"deviceNos"`
}
