package ucg

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"hello/internal/platform/cachekit"
)

var ucgCache = cachekit.Default()

// ChatMessage 聊天消息（Redis 持久化 JSON）。
type ChatMessage struct {
	ID                uint64 `json:"id"`
	ClientMsgID       string `json:"clientMsgId,omitempty"`
	SenderWxID        int64  `json:"senderWxId"`
	Content           string `json:"content"`
	ImageKey          string `json:"imageKey,omitempty"`
	VideoKey          string `json:"videoKey,omitempty"`
	MediaCdnUrl       string `json:"mediaCdnUrl,omitempty"`
	MediaThumbnailUrl string `json:"mediaThumbnailUrl,omitempty"`
	CreatedAt         int64  `json:"createdAt"`
	Status            string `json:"status"`
	AuditStatus       string `json:"auditStatus,omitempty"`
	AuditVersion      int    `json:"auditVersion,omitempty"`
	RejectReason      string `json:"rejectReason,omitempty"`
}

func appendChatMessage(ctx context.Context, convID uint64, msg ChatMessage) (ChatMessage, error) {
	seq, err := ucgCache.Incr(ctx, cachekit.UCGChatMsgSeqKey(convID))
	if err != nil {
		return msg, err
	}
	msg.ID = uint64(seq)
	if msg.CreatedAt == 0 {
		msg.CreatedAt = time.Now().Unix()
	}
	raw, err := json.Marshal(msg)
	if err != nil {
		return msg, err
	}
	if err = ucgCache.ListPush(ctx, cachekit.UCGChatMsgListKey(convID), string(raw)); err != nil {
		return msg, err
	}
	_ = ucgCache.Persist(ctx, cachekit.UCGChatMsgListKey(convID))
	return msg, nil
}

func listChatMessagesForViewer(ctx context.Context, convID uint64, viewerWxID int64, page, pageSize int) (total int, list []ChatMessage, err error) {
	_, list, err = listChatMessages(ctx, convID, page, pageSize)
	if err != nil {
		return 0, nil, err
	}
	filtered := filterChatMessagesForViewer(list, viewerWxID)
	return len(filtered), filtered, nil
}

func filterChatMessagesForViewer(msgs []ChatMessage, viewerWxID int64) []ChatMessage {
	out := make([]ChatMessage, 0, len(msgs))
	for _, msg := range msgs {
		if chatMessageVisibleToViewer(msg, viewerWxID) {
			out = append(out, msg)
		}
	}
	return out
}

func chatMessageVisibleToViewer(msg ChatMessage, viewerWxID int64) bool {
	audit := strings.TrimSpace(msg.AuditStatus)
	if audit == "" || audit == ChatAuditStatusApproved || audit == ChatAuditStatusPending {
		return true
	}
	if audit == ChatAuditStatusRejected {
		return msg.SenderWxID == viewerWxID
	}
	return true
}

func listChatMessages(ctx context.Context, convID uint64, page, pageSize int) (total int, list []ChatMessage, err error) {
	p := NormalizePage(page, pageSize)
	listKey := cachekit.UCGChatMsgListKey(convID)
	redisLen, err := ucgCache.ListLen(ctx, listKey)
	if err != nil {
		return 0, nil, err
	}
	if redisLen == 0 {
		mysqlCount, cErr := countChatMessagesMySQL(ctx, convID)
		if cErr != nil {
			return 0, nil, cErr
		}
		if mysqlCount > 0 {
			_ = tryWarmChatFromMySQL(ctx, convID)
			if mysqlCount > chatWarmMaxMessages() {
				return listChatMessagesMySQL(ctx, convID, page, pageSize)
			}
			redisLen, err = ucgCache.ListLen(ctx, listKey)
			if err != nil {
				return 0, nil, err
			}
			if redisLen == 0 {
				return listChatMessagesMySQL(ctx, convID, page, pageSize)
			}
		} else {
			return 0, []ChatMessage{}, nil
		}
	}
	total = int(redisLen)
	end := total - (p.Page-1)*p.PageSize - 1
	start := end - p.PageSize + 1
	if end < 0 {
		return total, []ChatMessage{}, nil
	}
	if start < 0 {
		start = 0
	}
	rows, err := ucgCache.ListRange(ctx, listKey, int64(start), int64(end))
	if err != nil {
		return 0, nil, err
	}
	list = make([]ChatMessage, 0, len(rows))
	for _, item := range rows {
		var msg ChatMessage
		if uErr := json.Unmarshal([]byte(item), &msg); uErr != nil {
			continue
		}
		enrichChatMessageMedia(&msg)
		list = append(list, msg)
	}
	return total, list, nil
}

func incrUnread(ctx context.Context, convID uint64, wxID int64) error {
	_, err := ucgCache.Incr(ctx, cachekit.UCGChatUnreadKey(convID, wxID))
	return err
}

func resetUnread(ctx context.Context, convID uint64, wxID int64) error {
	return ucgCache.Del(ctx, cachekit.UCGChatUnreadKey(convID, wxID))
}

func getUnread(ctx context.Context, convID uint64, wxID int64) (int, error) {
	raw, ok, err := ucgCache.Get(ctx, cachekit.UCGChatUnreadKey(convID, wxID))
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, nil
	}
	n, _ := strconv.Atoi(strings.TrimSpace(raw))
	return n, nil
}

func touchUserConversation(ctx context.Context, wxID int64, convID uint64, score int64) error {
	return ucgCache.SortedSetAdd(ctx, cachekit.UCGChatUserConvKey(wxID), float64(score), strconv.FormatUint(convID, 10))
}

func enrichChatMessageMedia(msg *ChatMessage) {
	if msg == nil {
		return
	}
	imageKey := strings.TrimSpace(msg.ImageKey)
	videoKey := strings.TrimSpace(msg.VideoKey)
	if imageKey != "" {
		if strings.TrimSpace(msg.MediaCdnUrl) == "" {
			msg.MediaCdnUrl = BuildCdnURL(imageKey)
		}
		msg.MediaThumbnailUrl = BuildImageThumbnailURL(imageKey)
		return
	}
	if videoKey != "" {
		if strings.TrimSpace(msg.MediaCdnUrl) == "" {
			msg.MediaCdnUrl = BuildCdnURL(videoKey)
		}
		msg.MediaThumbnailUrl = BuildVideoThumbnailURL(videoKey)
	}
}
