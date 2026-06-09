package ucg

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	chatOutboxStatusPending = "pending"
	chatOutboxStatusDone    = "done"
	chatOutboxStatusFailed  = "failed"

	chatPersistPollMsEnv    = "UCG_CHAT_PERSIST_POLL_MS"
	chatPersistMaxAttempts  = 10
	chatWarmMaxMessagesEnv  = "UCG_CHAT_WARM_MAX_MESSAGES"
	chatWarmMaxMessagesDef  = 200
	chatRebuildLockSeconds  = 30
)

func chatPersistPollInterval() time.Duration {
	ms := 1500
	if v := strings.TrimSpace(os.Getenv(chatPersistPollMsEnv)); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 100 {
			ms = n
		}
	}
	return time.Duration(ms) * time.Millisecond
}

func chatWarmMaxMessages() int {
	if v := strings.TrimSpace(os.Getenv(chatWarmMaxMessagesEnv)); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return chatWarmMaxMessagesDef
}

func redisChatRebuildLockKey(convID uint64) string {
	return fmt.Sprintf("ucg:chat:conv:%d:rebuild", convID)
}

// enqueueChatMessageOutbox 同步写入 MySQL outbox，供 persist worker 异步落 ucg_chat_message。
func enqueueChatMessageOutbox(ctx context.Context, convID uint64, msg ChatMessage) error {
	raw, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	_, err = dao.UcgChatMessageOutbox.Ctx(ctx).Data(g.Map{
		dao.UcgChatMessageOutbox.Columns().ConversationId: convID,
		dao.UcgChatMessageOutbox.Columns().Payload:        string(raw),
		dao.UcgChatMessageOutbox.Columns().Status:         chatOutboxStatusPending,
		dao.UcgChatMessageOutbox.Columns().Attempts:     0,
		dao.UcgChatMessageOutbox.Columns().LastError:    "",
		dao.UcgChatMessageOutbox.Columns().CreatedAt:    now,
	}).Insert()
	return err
}

func countChatMessagesMySQL(ctx context.Context, convID uint64) (int, error) {
	n, err := dao.UcgChatMessage.Ctx(ctx).
		Where(dao.UcgChatMessage.Columns().ConversationId, convID).
		Count()
	return n, err
}

func listChatMessagesMySQL(ctx context.Context, convID uint64, page, pageSize int) (total int, list []ChatMessage, err error) {
	p := NormalizePage(page, pageSize)
	total, err = countChatMessagesMySQL(ctx, convID)
	if err != nil || total == 0 {
		return total, []ChatMessage{}, err
	}
	offsetAsc := total - p.Page*p.PageSize
	limit := p.PageSize
	if offsetAsc < 0 {
		limit = p.PageSize + offsetAsc
		if limit <= 0 {
			return total, []ChatMessage{}, nil
		}
		offsetAsc = 0
	}
	var rows []entity.UcgChatMessage
	err = dao.UcgChatMessage.Ctx(ctx).
		Where(dao.UcgChatMessage.Columns().ConversationId, convID).
		OrderAsc(dao.UcgChatMessage.Columns().Id).
		Limit(limit).
		Offset(offsetAsc).
		Scan(&rows)
	if err != nil {
		return 0, nil, err
	}
	list = make([]ChatMessage, 0, len(rows))
	for _, row := range rows {
		msg := chatMessageFromEntity(row)
		enrichChatMessageMedia(&msg)
		list = append(list, msg)
	}
	return total, list, nil
}

func lastChatMessageMySQL(ctx context.Context, convID uint64) (ChatMessage, bool, error) {
	var row entity.UcgChatMessage
	err := dao.UcgChatMessage.Ctx(ctx).
		Where(dao.UcgChatMessage.Columns().ConversationId, convID).
		OrderDesc(dao.UcgChatMessage.Columns().Id).
		Limit(1).
		Scan(&row)
	if err != nil {
		return ChatMessage{}, false, err
	}
	if row.Id == 0 {
		return ChatMessage{}, false, nil
	}
	msg := chatMessageFromEntity(row)
	enrichChatMessageMedia(&msg)
	return msg, true, nil
}

func chatMessageFromEntity(row entity.UcgChatMessage) ChatMessage {
	return ChatMessage{
		ID:          row.Id,
		ClientMsgID: row.ClientMsgId,
		SenderWxID:  int64(row.SenderWxId),
		Content:     row.Content,
		ImageKey:    row.ImageKey,
		VideoKey:    row.VideoKey,
		MediaCdnUrl: row.MediaCdnUrl,
		CreatedAt:   row.CreatedAt,
		Status:      row.Status,
	}
}

// persistChatMessageRow 幂等写入 ucg_chat_message。
func persistChatMessageRow(ctx context.Context, convID uint64, msg ChatMessage) error {
	cols := dao.UcgChatMessage.Columns()
	cnt, err := dao.UcgChatMessage.Ctx(ctx).
		Where(cols.ConversationId, convID).
		Where(cols.Id, msg.ID).
		Count()
	if err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	status := strings.TrimSpace(msg.Status)
	if status == "" {
		status = "delivered"
	}
	_, err = dao.UcgChatMessage.Ctx(ctx).Data(g.Map{
		cols.Id:             msg.ID,
		cols.ConversationId: convID,
		cols.ClientMsgId:    strings.TrimSpace(msg.ClientMsgID),
		cols.SenderWxId:     msg.SenderWxID,
		cols.Content:        msg.Content,
		cols.ImageKey:       strings.TrimSpace(msg.ImageKey),
		cols.VideoKey:       strings.TrimSpace(msg.VideoKey),
		cols.MediaCdnUrl:    strings.TrimSpace(msg.MediaCdnUrl),
		cols.CreatedAt:      msg.CreatedAt,
		cols.Status:         status,
	}).Insert()
	return err
}

func maxChatMessageIDMySQL(ctx context.Context, convID uint64) (uint64, error) {
	var row entity.UcgChatMessage
	err := dao.UcgChatMessage.Ctx(ctx).
		Where(dao.UcgChatMessage.Columns().ConversationId, convID).
		OrderDesc(dao.UcgChatMessage.Columns().Id).
		Limit(1).
		Scan(&row)
	if err != nil {
		return 0, err
	}
	return row.Id, nil
}

// alignChatSeqFromMySQL 将 Redis seq 对齐至 MySQL MAX(id)，避免丢 Redis 后新消息 ID 冲突。
func alignChatSeqFromMySQL(ctx context.Context, convID uint64, maxID uint64) {
	if maxID == 0 {
		return
	}
	seqKey := redisChatMsgSeqKey(convID)
	raw, err := g.Redis().Do(ctx, "GET", seqKey)
	cur := uint64(0)
	if err == nil && raw != nil {
		cur = uint64(raw.Int64())
	}
	if cur >= maxID {
		return
	}
	if _, err = g.Redis().Do(ctx, "SET", seqKey, maxID); err != nil {
		glog.Warningf(ctx, "[ucg-chat] 对齐 Redis seq 失败 conv=%d maxId=%d err=%v", convID, maxID, err)
	}
}

// warmChatMessagesToRedis 将会话消息批量 RPUSH 回 Redis（调用方应保证 LIST 为空或持 rebuild 锁）。
func warmChatMessagesToRedis(ctx context.Context, convID uint64, msgs []ChatMessage) error {
	if len(msgs) == 0 {
		return nil
	}
	listKey := redisChatMsgListKey(convID)
	var maxID uint64
	for _, msg := range msgs {
		raw, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		if _, err = g.Redis().Do(ctx, "RPUSH", listKey, string(raw)); err != nil {
			return err
		}
		if msg.ID > maxID {
			maxID = msg.ID
		}
	}
	_, _ = g.Redis().Do(ctx, "PERSIST", listKey)
	alignChatSeqFromMySQL(ctx, convID, maxID)
	return nil
}

// tryWarmChatFromMySQL Redis LIST 为空且 MySQL 有数据时，按阈值全量 warm 或仅 MySQL 读。
func tryWarmChatFromMySQL(ctx context.Context, convID uint64) error {
	mysqlCount, err := countChatMessagesMySQL(ctx, convID)
	if err != nil || mysqlCount == 0 {
		return err
	}
	if mysqlCount > chatWarmMaxMessages() {
		return nil
	}
	lockKey := redisChatRebuildLockKey(convID)
	got, err := g.Redis().Do(ctx, "SET", lockKey, "1", "NX", "EX", chatRebuildLockSeconds)
	if err != nil {
		return err
	}
	if got == nil || got.IsEmpty() {
		return nil
	}
	defer func() { _, _ = g.Redis().Do(ctx, "DEL", lockKey) }()

	lenRaw, err := g.Redis().Do(ctx, "LLEN", redisChatMsgListKey(convID))
	if err != nil {
		return err
	}
	if lenRaw.Int() > 0 {
		return nil
	}
	var rows []entity.UcgChatMessage
	err = dao.UcgChatMessage.Ctx(ctx).
		Where(dao.UcgChatMessage.Columns().ConversationId, convID).
		OrderAsc(dao.UcgChatMessage.Columns().Id).
		Scan(&rows)
	if err != nil {
		return err
	}
	msgs := make([]ChatMessage, 0, len(rows))
	for _, row := range rows {
		msgs = append(msgs, chatMessageFromEntity(row))
	}
	return warmChatMessagesToRedis(ctx, convID, msgs)
}
