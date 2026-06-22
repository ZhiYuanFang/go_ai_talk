package ucg

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/database/gdb"
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

// enqueueChatMessageOutbox 同步写入 MySQL outbox，供 persist worker 异步落 ucg_chat_message。
func enqueueChatMessageOutbox(ctx context.Context, convID uint64, msg ChatMessage) error {
	return enqueueChatMessageOutboxTx(ctx, nil, convID, msg)
}

func enqueueChatMessageOutboxTx(ctx context.Context, tx gdb.TX, convID uint64, msg ChatMessage) error {
	raw, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	data := g.Map{
		dao.UcgChatMessageOutbox.Columns().ConversationId: convID,
		dao.UcgChatMessageOutbox.Columns().Payload:        string(raw),
		dao.UcgChatMessageOutbox.Columns().Status:         chatOutboxStatusPending,
		dao.UcgChatMessageOutbox.Columns().Attempts:       0,
		dao.UcgChatMessageOutbox.Columns().LastError:      "",
		dao.UcgChatMessageOutbox.Columns().CreatedAt:      now,
	}
	if tx != nil {
		_, err = tx.Model(dao.UcgChatMessageOutbox.Table()).Ctx(ctx).Data(data).Insert()
	} else {
		_, err = dao.UcgChatMessageOutbox.Ctx(ctx).Data(data).Insert()
	}
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
	auditStatus := strings.TrimSpace(row.AuditStatus)
	if auditStatus == "" {
		auditStatus = ChatAuditStatusApproved
	}
	auditVersion := row.AuditVersion
	if auditVersion <= 0 {
		auditVersion = 1
	}
	return ChatMessage{
		ID:           row.Id,
		ClientMsgID:  row.ClientMsgId,
		SenderWxID:   int64(row.SenderWxId),
		Content:      row.Content,
		ImageKey:     row.ImageKey,
		VideoKey:     row.VideoKey,
		MediaCdnUrl:  row.MediaCdnUrl,
		CreatedAt:    row.CreatedAt,
		Status:       row.Status,
		AuditStatus:  auditStatus,
		AuditVersion: auditVersion,
		RejectReason: row.RejectReason,
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
	auditStatus := strings.TrimSpace(msg.AuditStatus)
	if auditStatus == "" {
		auditStatus = ChatAuditStatusPending
	}
	auditVersion := msg.AuditVersion
	if auditVersion <= 0 {
		auditVersion = 1
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
		cols.AuditStatus:    auditStatus,
		cols.AuditVersion:   auditVersion,
		cols.RejectReason:   strings.TrimSpace(msg.RejectReason),
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
	seqKey := cachekit.UCGChatMsgSeqKey(convID)
	curRaw, ok, err := ucgCache.Get(ctx, seqKey)
	cur := uint64(0)
	if err == nil && ok {
		cur, _ = strconv.ParseUint(strings.TrimSpace(curRaw), 10, 64)
	}
	if cur >= maxID {
		return
	}
	if err := ucgCache.Set(ctx, seqKey, strconv.FormatUint(maxID, 10)); err != nil {
		glog.Warningf(ctx, "[ucg-chat] 对齐 Redis seq 失败 conv=%d maxId=%d err=%v", convID, maxID, err)
	}
}

// warmChatMessagesToRedis 将会话消息批量 RPUSH 回 Redis（调用方应保证 LIST 为空或持 rebuild 锁）。
func warmChatMessagesToRedis(ctx context.Context, convID uint64, msgs []ChatMessage) error {
	if len(msgs) == 0 {
		return nil
	}
	listKey := cachekit.UCGChatMsgListKey(convID)
	var maxID uint64
	for _, msg := range msgs {
		raw, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		if err = ucgCache.ListPush(ctx, listKey, string(raw)); err != nil {
			return err
		}
		if msg.ID > maxID {
			maxID = msg.ID
		}
	}
	_ = ucgCache.Persist(ctx, listKey)
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
	lockKey := cachekit.UCGChatRebuildLockKey(convID)
	got, err := ucgCache.SetNXEX(ctx, lockKey, "1", chatRebuildLockSeconds*time.Second)
	if err != nil {
		return err
	}
	if !got {
		return nil
	}
	defer func() { _ = ucgCache.Del(ctx, lockKey) }()

	redisLen, err := ucgCache.ListLen(ctx, cachekit.UCGChatMsgListKey(convID))
	if err != nil {
		return err
	}
	if redisLen > 0 {
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
