package ucg

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
)

const (
	simUnreadSampleMessageLimitDefault = 20
	simUnreadSampleMessageLimitMax     = 50
)

// SimUnreadChatMessageItem 供 sim T5 LLM 使用的最小消息视图。
type SimUnreadChatMessageItem struct {
	SenderWxId int64  `json:"senderWxId"`
	Content    string `json:"content"`
}

// SimUnreadChatSampleResult 未读会话抽样结果。
type SimUnreadChatSampleResult struct {
	Found          bool                       `json:"found"`
	ConversationId uint64                     `json:"conversationId,omitempty"`
	SimWxId        int64                      `json:"simWxId,omitempty"`
	PeerWxId       int64                      `json:"peerWxId,omitempty"`
	UnreadCount    int                        `json:"unreadCount,omitempty"`
	Messages       []SimUnreadChatMessageItem `json:"messages,omitempty"`
}

// SampleRandomSimUnreadChat 在 simWxIds 全集上随机抽一条 eligible 未读会话（真人 peer），并附带最近消息。
func SampleRandomSimUnreadChat(ctx context.Context, simWxIds []int64, messageLimit int) (SimUnreadChatSampleResult, error) {
	if len(simWxIds) == 0 {
		return SimUnreadChatSampleResult{Found: false}, nil
	}
	if messageLimit <= 0 {
		messageLimit = simUnreadSampleMessageLimitDefault
	}
	if messageLimit > simUnreadSampleMessageLimitMax {
		messageLimit = simUnreadSampleMessageLimitMax
	}

	bounds, err := simUnreadEligibleModel(ctx, simWxIds).
		Fields("MIN(m."+dao.UcgConversationMember.Columns().Id+") as min_id",
			"MAX(m."+dao.UcgConversationMember.Columns().Id+") as max_id").
		One()
	if err != nil {
		return SimUnreadChatSampleResult{}, err
	}
	if bounds.IsEmpty() {
		return SimUnreadChatSampleResult{Found: false}, nil
	}
	minID := bounds["min_id"].Uint64()
	maxID := bounds["max_id"].Uint64()
	if minID == 0 || maxID == 0 {
		return SimUnreadChatSampleResult{Found: false}, nil
	}

	anchor := minID
	if minID < maxID {
		u, uErr := simUnreadSampleRandomUnit()
		if uErr != nil {
			return SimUnreadChatSampleResult{}, uErr
		}
		span := float64(maxID - minID)
		anchor = minID + uint64(math.Floor(span*u))
		if anchor > maxID {
			anchor = maxID
		}
	}

	row, err := simUnreadEligiblePickFields(ctx, simWxIds).
		Where("m."+dao.UcgConversationMember.Columns().Id+" >= ?", anchor).
		OrderAsc("m." + dao.UcgConversationMember.Columns().Id).
		Limit(1).
		One()
	if err != nil {
		return SimUnreadChatSampleResult{}, err
	}
	if row.IsEmpty() {
		row, err = simUnreadEligiblePickFields(ctx, simWxIds).
			Where("m."+dao.UcgConversationMember.Columns().Id, minID).
			Limit(1).
			One()
		if err != nil {
			return SimUnreadChatSampleResult{}, err
		}
	}
	if row.IsEmpty() {
		return SimUnreadChatSampleResult{Found: false}, nil
	}

	convID := row["conversation_id"].Uint64()
	simWxID := row["sim_wx_id"].Int64()
	peerWxID := row["peer_wx_id"].Int64()
	unread := row["unread_count"].Int()

	msgs, err := loadRecentMessagesForSimReply(ctx, convID, simWxID, messageLimit)
	if err != nil {
		return SimUnreadChatSampleResult{}, err
	}

	return SimUnreadChatSampleResult{
		Found:          true,
		ConversationId: convID,
		SimWxId:        simWxID,
		PeerWxId:       peerWxID,
		UnreadCount:    unread,
		Messages:       msgs,
	}, nil
}

func simUnreadEligibleModel(ctx context.Context, simWxIds []int64) *gdb.Model {
	cols := dao.UcgConversationMember.Columns()
	peerTable := dao.UcgConversationMember.Table()
	wxArgs := int64SliceToInterface(simWxIds)
	return dao.UcgConversationMember.Ctx(ctx).As("m").
		InnerJoin(peerTable+" peer", fmt.Sprintf(
			"peer.%s=m.%s AND peer.%s!=m.%s",
			cols.ConversationId, cols.ConversationId, cols.WxId, cols.WxId,
		)).
		Where("m."+cols.UnreadCount+" > ?", 0).
		Where("m."+cols.DeletedAt, 0).
		WhereIn("m."+cols.WxId, wxArgs).
		WhereNotIn("peer."+cols.WxId, wxArgs).
		Where("peer."+cols.DeletedAt, 0)
}

func simUnreadEligiblePickFields(ctx context.Context, simWxIds []int64) *gdb.Model {
	cols := dao.UcgConversationMember.Columns()
	return simUnreadEligibleModel(ctx, simWxIds).Fields(
		"m."+cols.ConversationId+" as conversation_id",
		"m."+cols.WxId+" as sim_wx_id",
		"peer."+cols.WxId+" as peer_wx_id",
		"m."+cols.UnreadCount+" as unread_count",
	)
}

func loadRecentMessagesForSimReply(ctx context.Context, convID uint64, simWxID int64, limit int) ([]SimUnreadChatMessageItem, error) {
	var rows []entity.UcgChatMessage
	err := dao.UcgChatMessage.Ctx(ctx).
		Where(dao.UcgChatMessage.Columns().ConversationId, convID).
		OrderDesc(dao.UcgChatMessage.Columns().Id).
		Limit(limit).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []SimUnreadChatMessageItem{}, nil
	}
	// 反转为时间正序，供 LLM 阅读。
	filtered := make([]SimUnreadChatMessageItem, 0, len(rows))
	for i := len(rows) - 1; i >= 0; i-- {
		msg := chatMessageFromEntity(rows[i])
		if !chatMessageVisibleToViewer(msg, simWxID) {
			continue
		}
		filtered = append(filtered, SimUnreadChatMessageItem{
			SenderWxId: msg.SenderWxID,
			Content:    msg.Content,
		})
	}
	return filtered, nil
}

func int64SliceToInterface(ids []int64) []interface{} {
	out := make([]interface{}, len(ids))
	for i, id := range ids {
		out[i] = id
	}
	return out
}

func simUnreadSampleRandomUnit() (float64, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return float64(binary.BigEndian.Uint64(b[:])>>11) / float64(1<<53), nil
}
