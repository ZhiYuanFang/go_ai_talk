package ucg

import (
	"context"

	"github.com/gogf/gf/v2/os/glog"
)

// StartAuditWorker 已废弃：审核改由 MQ consumer 触发，保留空实现兼容旧调用。
func StartAuditWorker(ctx context.Context) {
	glog.Infof(ctx, "[ucg-audit] StartAuditWorker deprecated; use StartUcgAuditMQConsumer")
}
