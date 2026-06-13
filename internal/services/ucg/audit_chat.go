package ucg

import (
	"context"
	"encoding/json"
	"fmt"

	"hello/internal/dao"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// auditChatMessageFromEvent 私信 MQ 审核：两阶段 Green → apply CAS + Redis 同步。
// Green 单次以 MySQL 行为准；MySQL 未就绪时不调 Green，requeue 等待落库。
func auditChatMessageFromEvent(ctx context.Context, messageID, conversationID uint64, auditVersion int) error {
	msg, err := loadChatMessageForAudit(ctx, conversationID, messageID)
	if err != nil {
		return err
	}
	if msg.Id == 0 {
		glog.Infof(ctx, "[ucg-audit-mq] chat wait mysql msgId=%d conv=%d version=%d", messageID, conversationID, auditVersion)
		return fmt.Errorf("chat mysql not ready msgId=%d", messageID)
	}
	if msg.AuditStatus != ChatAuditStatusPending || msg.AuditVersion != auditVersion {
		glog.Infof(ctx, "[ucg-audit-mq] chat skip stale msgId=%d curAudit=%s curVer=%d eventVer=%d",
			messageID, msg.AuditStatus, msg.AuditVersion, auditVersion)
		return nil
	}

	runChatModerationPhase(ctx, ucgChatQueue, msg, auditVersion)
	msg, err = loadChatMessageForAudit(ctx, conversationID, messageID)
	if err != nil {
		return err
	}
	return runChatApplyPhase(ctx, ucgChatQueue, msg, auditVersion)
}

func approveChatMessageCAS(ctx context.Context, conversationID, messageID uint64, auditVersion int) error {
	affected, err := CasAuditTransition(ctx, CasAuditInput{
		Table:       dao.UcgChatMessage.Table(),
		IDColumn:    dao.UcgChatMessage.Columns().Id,
		ID:          messageID,
		Kind:        AuditCasKindAuditStatus,
		FromStatus:  ChatAuditStatusPending,
		ToStatus:    ChatAuditStatusApproved,
		FromVersion: auditVersion,
		ExtraWhere: g.Map{
			dao.UcgChatMessage.Columns().ConversationId: conversationID,
		},
		Extra: g.Map{
			dao.UcgChatMessage.Columns().RejectReason: "",
		},
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		glog.Infof(ctx, "[ucg-audit-mq] chat approve cas skip msgId=%d version=%d", messageID, auditVersion)
		return nil
	}
	return syncChatAuditToRedis(ctx, conversationID, messageID, ChatAuditStatusApproved, auditVersion, "")
}

func rejectChatMessageCAS(ctx context.Context, conversationID, messageID uint64, auditVersion int, senderWxID int64, clientMsgID, reason string) error {
	if reason == "" {
		reason = rejectReasonDefault
	}
	affected, err := CasAuditTransition(ctx, CasAuditInput{
		Table:       dao.UcgChatMessage.Table(),
		IDColumn:    dao.UcgChatMessage.Columns().Id,
		ID:          messageID,
		Kind:        AuditCasKindAuditStatus,
		FromStatus:  ChatAuditStatusPending,
		ToStatus:    ChatAuditStatusRejected,
		FromVersion: auditVersion,
		ExtraWhere: g.Map{
			dao.UcgChatMessage.Columns().ConversationId: conversationID,
		},
		Extra: g.Map{
			dao.UcgChatMessage.Columns().RejectReason: reason,
		},
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		glog.Infof(ctx, "[ucg-audit-mq] chat reject cas skip msgId=%d version=%d", messageID, auditVersion)
		return nil
	}
	if rErr := syncChatAuditToRedis(ctx, conversationID, messageID, ChatAuditStatusRejected, auditVersion, reason); rErr != nil {
		glog.Warningf(ctx, "[ucg-audit-mq] chat redis sync failed msgId=%d err=%v", messageID, rErr)
	}
	sendChatAuditFailed(senderWxID, clientMsgID, reason)
	recipient, pErr := peerWxID(ctx, conversationID, senderWxID)
	if pErr == nil {
		sendChatMsgHidden(int64(recipient), conversationID, messageID)
	}
	return nil
}

func sendChatMsgHidden(recipientWxID int64, convID, msgID uint64) {
	ChatWSHub().SendJSON(recipientWxID, map[string]interface{}{
		"type":           "msg_hidden",
		"conversationId": convID,
		"messageId":      msgID,
	})
}

// syncChatAuditToRedis CAS 成功后 LSET 更新 Redis 列表中对应消息的审态 JSON。
func syncChatAuditToRedis(ctx context.Context, convID, msgID uint64, auditStatus string, auditVersion int, rejectReason string) error {
	listKey := redisChatMsgListKey(convID)
	lenRaw, err := g.Redis().Do(ctx, "LLEN", listKey)
	if err != nil {
		return err
	}
	n := lenRaw.Int()
	for i := 0; i < n; i++ {
		raw, lErr := g.Redis().Do(ctx, "LINDEX", listKey, i)
		if lErr != nil {
			return lErr
		}
		var msg ChatMessage
		if uErr := json.Unmarshal([]byte(raw.String()), &msg); uErr != nil {
			continue
		}
		if msg.ID != msgID {
			continue
		}
		msg.AuditStatus = auditStatus
		msg.AuditVersion = auditVersion
		msg.RejectReason = rejectReason
		updated, mErr := json.Marshal(msg)
		if mErr != nil {
			return mErr
		}
		_, err = g.Redis().Do(ctx, "LSET", listKey, i, string(updated))
		return err
	}
	return nil
}
