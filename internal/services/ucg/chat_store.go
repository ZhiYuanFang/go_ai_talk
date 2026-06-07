package ucg

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

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
	Status            string `json:"status"` // pending | delivered
}

func appendChatMessage(ctx context.Context, convID uint64, msg ChatMessage) (ChatMessage, error) {
	seqRaw, err := g.Redis().Do(ctx, "INCR", redisChatMsgSeqKey(convID))
	if err != nil {
		return msg, err
	}
	msg.ID = uint64(seqRaw.Int64())
	if msg.CreatedAt == 0 {
		msg.CreatedAt = time.Now().Unix()
	}
	raw, err := json.Marshal(msg)
	if err != nil {
		return msg, err
	}
	if _, err = g.Redis().Do(ctx, "RPUSH", redisChatMsgListKey(convID), string(raw)); err != nil {
		return msg, err
	}
	// 永久保留：显式 PERSIST（若 key 曾误设 TTL 则清除）
	_, _ = g.Redis().Do(ctx, "PERSIST", redisChatMsgListKey(convID))
	return msg, nil
}

func listChatMessages(ctx context.Context, convID uint64, page, pageSize int) (total int, list []ChatMessage, err error) {
	p := NormalizePage(page, pageSize)
	lenRaw, err := g.Redis().Do(ctx, "LLEN", redisChatMsgListKey(convID))
	if err != nil {
		return 0, nil, err
	}
	total = lenRaw.Int()
	if total == 0 {
		return 0, []ChatMessage{}, nil
	}
	// 分页从最新消息往回取：page=1 为最新一页
	end := total - (p.Page-1)*p.PageSize - 1
	start := end - p.PageSize + 1
	if end < 0 {
		return total, []ChatMessage{}, nil
	}
	if start < 0 {
		start = 0
	}
	rows, err := g.Redis().Do(ctx, "LRANGE", redisChatMsgListKey(convID), start, end)
	if err != nil {
		return 0, nil, err
	}
	list = make([]ChatMessage, 0, len(rows.Array()))
	for _, item := range rows.Array() {
		var msg ChatMessage
		if uErr := json.Unmarshal([]byte(g.NewVar(item).String()), &msg); uErr != nil {
			continue
		}
		enrichChatMessageMedia(&msg)
		list = append(list, msg)
	}
	// LRANGE 升序，前端通常要时间升序展示
	return total, list, nil
}

func incrUnread(ctx context.Context, convID uint64, wxID int64) error {
	_, err := g.Redis().Do(ctx, "INCR", redisChatUnreadKey(convID, wxID))
	return err
}

func resetUnread(ctx context.Context, convID uint64, wxID int64) error {
	_, err := g.Redis().Do(ctx, "DEL", redisChatUnreadKey(convID, wxID))
	return err
}

func getUnread(ctx context.Context, convID uint64, wxID int64) (int, error) {
	raw, err := g.Redis().Do(ctx, "GET", redisChatUnreadKey(convID, wxID))
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(raw.String())
	if s == "" {
		return 0, nil
	}
	n, _ := strconv.Atoi(s)
	return n, nil
}

func touchUserConversation(ctx context.Context, wxID int64, convID uint64, score int64) error {
	_, err := g.Redis().Do(ctx, "ZADD", redisChatUserConvKey(wxID), score, strconv.FormatUint(convID, 10))
	return err
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
	if videoKey != "" && strings.TrimSpace(msg.MediaCdnUrl) == "" {
		msg.MediaCdnUrl = BuildCdnURL(videoKey)
	}
}
