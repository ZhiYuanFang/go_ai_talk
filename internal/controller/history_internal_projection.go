package controller

import (
	"context"

	v1 "hello/api/v1"
	historysvc "hello/internal/services/history"
)

// HistoryInternalProjectionCtrl history 域内部投影修复（供 worker 调用，依赖网络隔离）。
type HistoryInternalProjectionCtrl struct{}

// NewHistoryInternalProjectionCtrl 构造 HistoryInternalProjectionCtrl。
func NewHistoryInternalProjectionCtrl() *HistoryInternalProjectionCtrl {
	return &HistoryInternalProjectionCtrl{}
}

// Reconcile 抽样设备并重建历史读模型、生日与事件选项等缓存。
func (c *HistoryInternalProjectionCtrl) Reconcile(ctx context.Context, req *v1.HistoryInternalProjectionReconcileReq) (res *v1.HistoryInternalProjectionReconcileRes, err error) {
	nos, err := historysvc.ReconcileProjectionCachesForWorker(ctx, req.Limit)
	if err != nil {
		return nil, err
	}
	return &v1.HistoryInternalProjectionReconcileRes{DeviceNos: nos}, nil
}
