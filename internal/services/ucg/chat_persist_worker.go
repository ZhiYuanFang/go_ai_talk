package ucg

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// StartChatPersistWorker 轮询 ucg_chat_message_outbox，异步落库 ucg_chat_message。
func StartChatPersistWorker(ctx context.Context) {
	interval := chatPersistPollInterval()
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := flushOneChatOutbox(ctx); err != nil {
					glog.Warningf(ctx, "[ucg-chat-persist] flush failed: %v", err)
				}
			}
		}
	}()
	glog.Infof(ctx, "[ucg-chat-persist] worker started interval=%s", interval)
}

func flushOneChatOutbox(ctx context.Context) error {
	cols := dao.UcgChatMessageOutbox.Columns()
	var item entity.UcgChatMessageOutbox
	err := dao.UcgChatMessageOutbox.Ctx(ctx).
		WhereIn(cols.Status, []string{chatOutboxStatusPending, chatOutboxStatusFailed}).
		WhereLT(cols.Attempts, chatPersistMaxAttempts).
		OrderAsc(cols.Id).
		Limit(1).
		Scan(&item)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if item.Id == 0 {
		return nil
	}
	var msg ChatMessage
	if uErr := json.Unmarshal([]byte(item.Payload), &msg); uErr != nil {
		_ = markChatOutboxFailed(ctx, item.Id, item.Attempts, "invalid payload json")
		return uErr
	}
	if pErr := persistChatMessageRow(ctx, item.ConversationId, msg); pErr != nil {
		_ = markChatOutboxFailed(ctx, item.Id, item.Attempts, pErr.Error())
		return pErr
	}
	_, err = dao.UcgChatMessageOutbox.Ctx(ctx).
		Where(cols.Id, item.Id).
		Data(g.Map{
			cols.Status:     chatOutboxStatusDone,
			cols.Attempts:   item.Attempts + 1,
			cols.LastError:  "",
		}).Update()
	return err
}

func markChatOutboxFailed(ctx context.Context, id uint64, attempts uint, errMsg string) error {
	errMsg = truncateChatError(errMsg, 512)
	cols := dao.UcgChatMessageOutbox.Columns()
	_, err := dao.UcgChatMessageOutbox.Ctx(ctx).
		Where(cols.Id, id).
		Data(g.Map{
			cols.Status:    chatOutboxStatusFailed,
			cols.Attempts:  attempts + 1,
			cols.LastError: errMsg,
		}).Update()
	return err
}

func truncateChatError(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max]
}
