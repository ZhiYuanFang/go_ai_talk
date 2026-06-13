package ucg

import (
	"context"
	"encoding/json"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// auditChatMessageFromEvent 私信 MQ 审核：单阶段 Green + MySQL CAS + Redis LSET 同步。
// Green err → requeue → 重复调 Green；无 moderation_verdict 两阶段隔离。
func auditChatMessageFromEvent(ctx context.Context, messageID, conversationID uint64, auditVersion int) error {
	row, err := dao.UcgChatMessage.Ctx(ctx).
		Where(dao.UcgChatMessage.Columns().ConversationId, conversationID).
		Where(dao.UcgChatMessage.Columns().Id, messageID).
		One()
	if err != nil {
		return err
	}
	if row.IsEmpty() {
		// MySQL 尚未落库时尝试 Redis 路径（outbox 与 consumer 竞态）
		return auditChatMessageFromRedis(ctx, conversationID, messageID, auditVersion)
	}
	var msg entity.UcgChatMessage
	if err = row.Struct(&msg); err != nil {
		return err
	}
	if msg.AuditStatus != ChatAuditStatusPending || msg.AuditVersion != auditVersion {
		glog.Infof(ctx, "[ucg-audit-mq] chat skip stale msgId=%d curAudit=%s curVer=%d eventVer=%d",
			messageID, msg.AuditStatus, msg.AuditVersion, auditVersion)
		return nil // 已处理 → Ack
	}
	return runChatGreenAndCAS(ctx, conversationID, messageID, auditVersion, msg.SenderWxId, msg.Content, msg.ImageKey, msg.VideoKey, msg.MediaCdnUrl, msg.ClientMsgId)
}

func auditChatMessageFromRedis(ctx context.Context, conversationID, messageID uint64, auditVersion int) error {
	chatMsg, ok, err := findChatMessageInRedis(ctx, conversationID, messageID)
	if err != nil || !ok {
		glog.Infof(ctx, "[ucg-audit-mq] chat skip not found msgId=%d conv=%d version=%d", messageID, conversationID, auditVersion)
		return err // Redis 也找不到：err 可能 nil（ok=false）→ Ack；err 非 nil → requeue
	}
	if chatMsg.AuditStatus != ChatAuditStatusPending || chatMsg.AuditVersion != auditVersion {
		glog.Infof(ctx, "[ucg-audit-mq] chat redis skip stale msgId=%d version=%d", messageID, auditVersion)
		return nil
	}
	return runChatGreenAndCAS(ctx, conversationID, messageID, auditVersion, uint64(chatMsg.SenderWxID), chatMsg.Content, chatMsg.ImageKey, chatMsg.VideoKey, chatMsg.MediaCdnUrl, chatMsg.ClientMsgID)
}

// runChatGreenAndCAS 串行：文本 → 图片 → 视频 Green；任一步 err 则整 handler err → requeue。
func runChatGreenAndCAS(ctx context.Context, conversationID, messageID uint64, auditVersion int,
	senderWxID uint64, content, imageKey, videoKey, mediaCdnURL, clientMsgID string) error {
	moderator := EffectiveGreen()
	cfg := LoadOSSConfig(ctx)

	if content != "" {
		if verdict, mErr := moderator.ModerateText(ctx, "comment_detection", content); mErr != nil {
			return mErr
		} else if !verdict.Pass {
			return rejectChatMessageCAS(ctx, conversationID, messageID, auditVersion, int64(senderWxID), clientMsgID, verdict.Reason)
		}
	}
	if imageKey != "" {
		url := mediaCdnURL
		if url == "" {
			url = cfg.CdnBaseURL + "/" + imageKey
		}
		if verdict, mErr := moderator.ModerateImageURL(ctx, url); mErr != nil {
			return mErr
		} else if !verdict.Pass {
			return rejectChatMessageCAS(ctx, conversationID, messageID, auditVersion, int64(senderWxID), clientMsgID, verdict.Reason)
		}
	}
	if videoKey != "" {
		url := mediaCdnURL
		if url == "" {
			url = cfg.CdnBaseURL + "/" + videoKey
		}
		if verdict, mErr := moderator.ModerateVideoURL(ctx, url); mErr != nil {
			return mErr
		} else if !verdict.Pass {
			return rejectChatMessageCAS(ctx, conversationID, messageID, auditVersion, int64(senderWxID), clientMsgID, verdict.Reason)
		}
	}
	return approveChatMessageCAS(ctx, conversationID, messageID, auditVersion)
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
		return err // CAS 失败 → requeue → 可能重复 Green
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

func findChatMessageInRedis(ctx context.Context, convID, msgID uint64) (ChatMessage, bool, error) {
	listKey := redisChatMsgListKey(convID)
	rows, err := g.Redis().Do(ctx, "LRANGE", listKey, 0, -1)
	if err != nil {
		return ChatMessage{}, false, err
	}
	for _, item := range rows.Array() {
		var msg ChatMessage
		if uErr := json.Unmarshal([]byte(g.NewVar(item).String()), &msg); uErr != nil {
			continue
		}
		if msg.ID == msgID {
			return msg, true, nil
		}
	}
	return ChatMessage{}, false, nil
}
