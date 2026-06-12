package ucg

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/os/glog"
)

const (
	ucgAuditConsumerEnabledEnv = "UCG_AUDIT_MQ_CONSUMER_ENABLED"

	ucgPostQueue    = "ucg.post.created.q"
	ucgCommentQueue = "ucg.comment.created.q"
	ucgProfileQueue = "ucg.profile.patch.submitted.q"
	ucgChatQueue    = "ucg.chat.msg.created.q"
)

var ucgAuditQueues = []string{ucgPostQueue, ucgCommentQueue, ucgProfileQueue, ucgChatQueue}

func dispatchUcgAuditPayload(ctx context.Context, queueName, payload string) error {
	var raw map[string]any
	if err := json.Unmarshal([]byte(payload), &raw); err != nil {
		glog.Warningf(ctx, "[ucg-audit-mq] invalid payload queue=%s err=%v", queueName, err)
		return nil
	}
	auditVersion := int(jsonInt(raw, "auditVersion"))
	if auditVersion <= 0 {
		glog.Warningf(ctx, "[ucg-audit-mq] missing auditVersion queue=%s payload=%s", queueName, payload)
		return nil
	}
	switch queueName {
	case ucgPostQueue:
		postID := uint64(jsonInt(raw, "postId"))
		if postID == 0 {
			return nil
		}
		return auditPostFromEvent(ctx, postID, auditVersion)
	case ucgCommentQueue:
		commentID := uint64(jsonInt(raw, "commentId"))
		if commentID == 0 {
			return nil
		}
		return auditCommentFromEvent(ctx, commentID, auditVersion)
	case ucgProfileQueue:
		jobID := uint64(jsonInt(raw, "jobId"))
		if jobID == 0 {
			return nil
		}
		return auditProfileJobFromEvent(ctx, jobID, auditVersion)
	case ucgChatQueue:
		msgID := uint64(jsonInt(raw, "messageId"))
		convID := uint64(jsonInt(raw, "conversationId"))
		if msgID == 0 || convID == 0 {
			return nil
		}
		return auditChatMessageFromEvent(ctx, msgID, convID, auditVersion)
	default:
		return nil
	}
}

func jsonInt(m map[string]any, key string) int64 {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case int64:
		return n
	case json.Number:
		i, _ := n.Int64()
		return i
	default:
		return 0
	}
}

func ucgAuditConsumerEnabled() bool {
	v := strings.TrimSpace(os.Getenv(ucgAuditConsumerEnabledEnv))
	if v == "" {
		return true
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return true
	}
	return b
}
