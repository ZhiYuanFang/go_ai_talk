package ucg

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"hello/internal/platform/eventkit"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const auditPublishOutboxTable = "ucg_audit_publish_outbox"

const (
	auditPublishOutboxPending = "pending"
	auditPublishOutboxDone    = "done"
	auditPublishOutboxFailed  = "failed"
)

// AuditPublishConfig relay worker 参数。
type AuditPublishConfig struct {
	RelayIntervalMs int
	MaxAttempts     int
}

func LoadAuditPublishConfig(ctx context.Context) AuditPublishConfig {
	cfg := AuditPublishConfig{
		RelayIntervalMs: g.Cfg().MustGet(ctx, "ucg.auditPublish.relayIntervalMs").Int(),
		MaxAttempts:     g.Cfg().MustGet(ctx, "ucg.auditPublish.maxAttempts").Int(),
	}
	if cfg.RelayIntervalMs <= 0 {
		cfg.RelayIntervalMs = 1000
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 20
	}
	return cfg
}

type auditPublishOutboxRow struct {
	Id          uint64 `json:"id"`
	RoutingKey  string `json:"routing_key"`
	Payload     string `json:"payload"`
	Status      string `json:"status"`
	Attempts    uint   `json:"attempts"`
	LastError   string `json:"last_error"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// enqueueAuditPublishOutboxTx 在业务事务内写入 outbox；payload 在入队时冻结。
func enqueueAuditPublishOutboxTx(ctx context.Context, tx gdb.TX, routingKey string, payload map[string]any) (uint64, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	now := time.Now().Unix()
	res, err := tx.Model(auditPublishOutboxTable).Ctx(ctx).Data(g.Map{
		"routing_key": strings.TrimSpace(routingKey),
		"payload":     string(raw),
		"status":      auditPublishOutboxPending,
		"attempts":    0,
		"last_error":  "",
		"created_at":  now,
		"updated_at":  now,
	}).Insert()
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return uint64(id), nil
}

// scheduleAuditPublishAfterCommit 事务提交后 best-effort 即时 Publish。
func scheduleAuditPublishAfterCommit(ctx context.Context, outboxID uint64) {
	if outboxID == 0 {
		return
	}
	if err := tryRelayAuditOutboxByID(ctx, outboxID); err != nil {
		glog.Warningf(ctx, "[ucg-audit-outbox] immediate relay failed id=%d err=%v", outboxID, err)
	}
}

func tryRelayAuditOutboxByID(ctx context.Context, outboxID uint64) error {
	row, ok, err := loadAuditPublishOutboxByID(ctx, outboxID)
	if err != nil || !ok {
		return err
	}
	if row.Status != auditPublishOutboxPending && row.Status != auditPublishOutboxFailed {
		return nil
	}
	cfg := LoadAuditPublishConfig(ctx)
	if row.Attempts >= uint(cfg.MaxAttempts) {
		return nil
	}
	return relayAuditPublishOutboxRow(ctx, row)
}

func loadAuditPublishOutboxByID(ctx context.Context, id uint64) (auditPublishOutboxRow, bool, error) {
	var row auditPublishOutboxRow
	rec, err := g.DB().Model(auditPublishOutboxTable).Ctx(ctx).Where("id", id).One()
	if err != nil {
		return row, false, err
	}
	if rec.IsEmpty() {
		return row, false, nil
	}
	row.Id = rec["id"].Uint64()
	row.RoutingKey = rec["routing_key"].String()
	row.Payload = rec["payload"].String()
	row.Status = rec["status"].String()
	row.Attempts = rec["attempts"].Uint()
	row.LastError = rec["last_error"].String()
	row.CreatedAt = rec["created_at"].Int64()
	row.UpdatedAt = rec["updated_at"].Int64()
	return row, true, nil
}

func relayAuditPublishOutboxRow(ctx context.Context, row auditPublishOutboxRow) error {
	var payload map[string]any
	if err := json.Unmarshal([]byte(row.Payload), &payload); err != nil {
		_ = markAuditPublishOutboxFailed(ctx, row.Id, row.Attempts, "invalid payload json")
		return err
	}
	p := defaultUcgAuditPublisher(ctx)
	if err := p.publish(ctx, eventkit.RouteKey(row.RoutingKey), payload); err != nil {
		_ = markAuditPublishOutboxFailed(ctx, row.Id, row.Attempts, err.Error())
		return err
	}
	return markAuditPublishOutboxDone(ctx, row.Id, row.Attempts)
}

func markAuditPublishOutboxDone(ctx context.Context, id uint64, attempts uint) error {
	now := time.Now().Unix()
	_, err := g.DB().Model(auditPublishOutboxTable).Ctx(ctx).Where("id", id).Data(g.Map{
		"status":     auditPublishOutboxDone,
		"attempts":   attempts + 1,
		"last_error": "",
		"updated_at": now,
	}).Update()
	return err
}

func markAuditPublishOutboxFailed(ctx context.Context, id uint64, attempts uint, errMsg string) error {
	errMsg = truncateChatError(errMsg, 512)
	now := time.Now().Unix()
	_, err := g.DB().Model(auditPublishOutboxTable).Ctx(ctx).Where("id", id).Data(g.Map{
		"status":     auditPublishOutboxFailed,
		"attempts":   attempts + 1,
		"last_error": errMsg,
		"updated_at": now,
	}).Update()
	return err
}

func auditPublishPostPayload(postID uint64, auditVersion int) map[string]any {
	return map[string]any{"postId": postID, "auditVersion": auditVersion}
}

func auditPublishCommentPayload(commentID uint64, auditVersion int) map[string]any {
	return map[string]any{"commentId": commentID, "auditVersion": auditVersion}
}

func auditPublishProfilePayload(jobID uint64, auditVersion int) map[string]any {
	return map[string]any{"jobId": jobID, "auditVersion": auditVersion}
}

func auditPublishChatPayload(messageID, conversationID uint64, auditVersion int) map[string]any {
	return map[string]any{
		"messageId":      messageID,
		"conversationId": conversationID,
		"auditVersion":   auditVersion,
	}
}
