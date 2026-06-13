package ucg

import (
	"context"

	"hello/internal/platform/eventkit"
	"hello/internal/shared/mq"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// UcgAuditPublisher 发布 UCG 四类审核事件；载荷 MUST 含 auditVersion。
type UcgAuditPublisher struct {
	pub eventkit.Publisher
}

func NewUcgAuditPublisher() (*UcgAuditPublisher, error) {
	pub, err := mq.NewObservedEventPublisher()
	if err != nil {
		return nil, err
	}
	return &UcgAuditPublisher{pub: pub}, nil
}

func defaultUcgAuditPublisher(ctx context.Context) *UcgAuditPublisher {
	p, err := NewUcgAuditPublisher()
	if err != nil {
		glog.Warningf(ctx, "[ucg-audit-mq] publisher init failed: %v", err)
		return nil
	}
	return p
}

func (p *UcgAuditPublisher) publish(ctx context.Context, key eventkit.RouteKey, payload map[string]any) error {
	// 如果 publisher 为空，则不发布
	if p == nil || p.pub == nil {
		glog.Warningf(ctx, "[ucg-audit-mq] publish skipped (no publisher) key=%s payload=%v", key, payload)
		return eventkit.ErrUnavailable
	}
	// 发布 outbox
	return p.pub.Publish(ctx, key.String(), payload)
}

// PublishPostCreated 非事务路径：写 outbox 并即时 relay（兼容旧调用）。
func PublishPostCreated(ctx context.Context, postID uint64, auditVersion int) {
	enqueueAndRelayAuditPublish(ctx, eventkit.RoutingUcgPostCreated, auditPublishPostPayload(postID, auditVersion))
}

// PublishCommentCreated 非事务路径：写 outbox 并即时 relay。
func PublishCommentCreated(ctx context.Context, commentID uint64, auditVersion int) {
	enqueueAndRelayAuditPublish(ctx, eventkit.RoutingUcgCommentCreated, auditPublishCommentPayload(commentID, auditVersion))
}

// PublishProfilePatchSubmitted 非事务路径：写 outbox 并即时 relay。
func PublishProfilePatchSubmitted(ctx context.Context, jobID uint64, auditVersion int) {
	enqueueAndRelayAuditPublish(ctx, eventkit.RoutingUcgProfilePatchSubmitted, auditPublishProfilePayload(jobID, auditVersion))
}

// PublishChatMsgCreated 非事务路径：写 outbox 并即时 relay。
func PublishChatMsgCreated(ctx context.Context, messageID, conversationID uint64, auditVersion int) {
	enqueueAndRelayAuditPublish(ctx, eventkit.RoutingUcgChatMsgCreated, auditPublishChatPayload(messageID, conversationID, auditVersion))
}

// enqueueAndRelayAuditPublish 独立 INSERT outbox + 即时 relay（无业务事务时使用）。
func enqueueAndRelayAuditPublish(ctx context.Context, key eventkit.RouteKey, payload map[string]any) {
	var outboxID uint64
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var txErr error
		outboxID, txErr = enqueueAuditPublishOutboxTx(ctx, tx, key.String(), payload)
		return txErr
	})
	if err != nil {
		glog.Warningf(ctx, "[ucg-audit-outbox] enqueue failed key=%s payload=%v err=%v", key, payload, err)
		return
	}
	scheduleAuditPublishAfterCommit(ctx, outboxID)
}
