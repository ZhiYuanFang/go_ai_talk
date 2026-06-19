package ucg

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// StartAuditPublishRelayWorker 轮询 ucg_audit_publish_outbox，重试 HTTP Publish；不扫 pending 业务表。
func StartAuditPublishRelayWorker(ctx context.Context) {
	cfg := LoadAuditPublishConfig(ctx)
	interval := time.Duration(cfg.RelayIntervalMs) * time.Millisecond
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := relayOneAuditPublishOutbox(ctx); err != nil {
					glog.Warningf(ctx, "[ucg-audit-outbox] relay tick failed: %v", err)
				}
			}
		}
	}()
	glog.Infof(ctx, "[ucg-audit-outbox] relay worker started interval=%s maxAttempts=%d", interval, cfg.MaxAttempts)
}

func relayOneAuditPublishOutbox(ctx context.Context) error {
	cfg := LoadAuditPublishConfig(ctx)
	var row auditPublishOutboxRow
	err := g.DB().Model(auditPublishOutboxTable).Ctx(ctx).
		WhereIn("status", []string{auditPublishOutboxPending, auditPublishOutboxFailed}).
		WhereLT("attempts", cfg.MaxAttempts).
		OrderAsc("id").
		Limit(1).
		Scan(&row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if row.Id == 0 {
		return nil
	}
	return relayAuditPublishOutboxRow(ctx, row)
}
