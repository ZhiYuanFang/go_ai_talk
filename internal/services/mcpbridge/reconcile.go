package mcpbridge

// reconcile.go：低频全量对齐绑定（xiaozhi-mcp-binding-reconcile）。
//
// OpenSpec 批准：宿主 xiaozhi-mcp-service；失败仅日志；间隔 0 关闭周期任务。

import (
	"context"
	"time"

	deviceclient "hello/internal/clients/device"

	"github.com/gogf/gf/v2/os/glog"
)

// PullAndReconcile 从 device-service 拉全量 active 绑定并对齐 Manager。
func PullAndReconcile(ctx context.Context, m *Manager) error {
	if m == nil {
		return nil
	}
	list, err := deviceclient.ListActiveXiaozhiMcpBindings(ctx)
	if err != nil {
		return err
	}
	desired := make([]BindingSpec, 0, len(list))
	for _, it := range list {
		desired = append(desired, BindingSpec{
			Id:       it.Id,
			McpToken: it.McpToken,
			DeviceNo: it.DeviceNo,
			WxId:     it.WxId,
		})
	}
	m.Reconcile(ctx, desired)
	return nil
}

// StartReconcileLoop 周期 reconcile；interval<=0 时立即返回（不启动 goroutine）。
// 调用方应在独立 goroutine 中调用本函数，或本函数内部已起循环并阻塞至 ctx 取消。
func StartReconcileLoop(ctx context.Context, m *Manager, interval time.Duration) {
	if interval <= 0 {
		glog.Infof(ctx, "[xiaozhi-mcp] reconcile loop disabled (interval<=0)")
		return
	}
	glog.Infof(ctx, "[xiaozhi-mcp] reconcile loop start interval=%v", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			glog.Infof(ctx, "[xiaozhi-mcp] reconcile loop stop")
			return
		case <-ticker.C:
			if err := PullAndReconcile(ctx, m); err != nil {
				glog.Warningf(ctx, "[xiaozhi-mcp] reconcile failed err=%v", err)
			}
		}
	}
}
