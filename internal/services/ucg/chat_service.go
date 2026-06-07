package ucg

import (
	"context"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// ConversationDTO 会话列表项。
type ConversationDTO struct {
	Id             uint64 `json:"id"`
	PeerWxId       uint64 `json:"peerWxId"`
	PeerNickname   string `json:"peerNickname,omitempty"`
	PeerAvatarKey  string `json:"peerAvatarKey,omitempty"`
	PeerAvatarUrl  string `json:"peerAvatarUrl,omitempty"`
	Pinned         int    `json:"pinned"`
	UnreadCount    int    `json:"unreadCount"`
	UpdatedAt      int64  `json:"updatedAt"`
	LastPreview    string `json:"lastPreview,omitempty"`
	Deleted        bool   `json:"deleted,omitempty"`
}

// GetOrCreateDirectConversation 获取或创建 1:1 会话。
func GetOrCreateDirectConversation(ctx context.Context, wxID, targetWxID int64) (*ConversationDTO, error) {
	if wxID <= 0 || targetWxID <= 0 || wxID == targetWxID {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "targetWxId 无效")
	}
	exists, _, err := Device().ValidateWx(ctx, targetWxID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, gerror.NewCode(gcode.CodeNotFound, "用户不存在")
	}
	convID, err := findDirectConversation(ctx, wxID, targetWxID)
	if err != nil {
		return nil, err
	}
	if convID == 0 {
		convID, err = createDirectConversation(ctx, wxID, targetWxID)
		if err != nil {
			return nil, err
		}
	}
	return loadConversationDTO(ctx, convID, wxID)
}

func findDirectConversation(ctx context.Context, a, b int64) (uint64, error) {
	// 查找双方均存在的 direct 会话（type=1）。
	rows, err := dao.UcgConversationMember.Ctx(ctx).
		Where(dao.UcgConversationMember.Columns().WxId, a).
		Fields(dao.UcgConversationMember.Columns().ConversationId).
		All()
	if err != nil {
		return 0, err
	}
	for _, row := range rows {
		var m entity.UcgConversationMember
		if err = row.Struct(&m); err != nil {
			return 0, err
		}
		cnt, cErr := dao.UcgConversationMember.Ctx(ctx).
			Where(dao.UcgConversationMember.Columns().ConversationId, m.ConversationId).
			Where(dao.UcgConversationMember.Columns().WxId, b).
			Count()
		if cErr != nil {
			return 0, cErr
		}
		if cnt > 0 {
			return m.ConversationId, nil
		}
	}
	return 0, nil
}

func createDirectConversation(ctx context.Context, a, b int64) (uint64, error) {
	now := time.Now().Unix()
	res, err := dao.UcgConversation.Ctx(ctx).Data(g.Map{
		dao.UcgConversation.Columns().Type:      1,
		dao.UcgConversation.Columns().CreatedAt: now,
		dao.UcgConversation.Columns().UpdatedAt: now,
	}).Insert()
	if err != nil {
		return 0, err
	}
	convID, _ := res.LastInsertId()
	for _, wxID := range []int64{a, b} {
		if _, err = dao.UcgConversationMember.Ctx(ctx).Data(g.Map{
			dao.UcgConversationMember.Columns().ConversationId: convID,
			dao.UcgConversationMember.Columns().WxId:           wxID,
			dao.UcgConversationMember.Columns().UpdatedAt:      now,
		}).Insert(); err != nil {
			return 0, err
		}
	}
	return uint64(convID), nil
}

// ListConversations 当前用户会话列表。
func ListConversations(ctx context.Context, wxID int64, page, pageSize int) (*PageResult, error) {
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "缺少 X-Internal-Wx-Id")
	}
	p := NormalizePage(page, pageSize)
	model := dao.UcgConversationMember.Ctx(ctx).Where(dao.UcgConversationMember.Columns().WxId, wxID)
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	rows, err := model.
		OrderDesc(dao.UcgConversationMember.Columns().Pinned).
		OrderDesc(dao.UcgConversationMember.Columns().UpdatedAt).
		Limit(p.PageSize).Offset(pageOffset(p)).All()
	if err != nil {
		return nil, err
	}
	list := make([]*ConversationDTO, 0, len(rows))
	for _, row := range rows {
		var m entity.UcgConversationMember
		if err = row.Struct(&m); err != nil {
			return nil, err
		}
		if m.DeletedAt > 0 {
			continue
		}
		dto, lErr := loadConversationDTO(ctx, m.ConversationId, wxID)
		if lErr != nil {
			return nil, lErr
		}
		list = append(list, dto)
	}
	return &PageResult{List: list, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

func loadConversationDTO(ctx context.Context, convID uint64, wxID int64) (*ConversationDTO, error) {
	selfRow, err := dao.UcgConversationMember.Ctx(ctx).
		Where(dao.UcgConversationMember.Columns().ConversationId, convID).
		Where(dao.UcgConversationMember.Columns().WxId, wxID).One()
	if err != nil || selfRow.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotFound, "会话不存在")
	}
	var self entity.UcgConversationMember
	if err = selfRow.Struct(&self); err != nil {
		return nil, err
	}
	peerID, err := peerWxID(ctx, convID, wxID)
	if err != nil {
		return nil, err
	}
	unread, _ := getUnread(ctx, convID, wxID)
	if unread == 0 {
		unread = int(self.UnreadCount)
	}
	preview := lastMessagePreview(ctx, convID)
	dto := &ConversationDTO{
		Id:          convID,
		PeerWxId:    peerID,
		Pinned:      self.Pinned,
		UnreadCount: unread,
		UpdatedAt:   self.UpdatedAt,
		LastPreview: preview,
		Deleted:     self.DeletedAt > 0,
	}
	if prof, pErr := GetPublicProfile(ctx, peerID); pErr == nil && prof != nil {
		dto.PeerNickname = prof.Nickname
		dto.PeerAvatarKey = prof.AvatarKey
		dto.PeerAvatarUrl = prof.AvatarUrl
	}
	return dto, nil
}

func peerWxID(ctx context.Context, convID uint64, wxID int64) (uint64, error) {
	row, err := dao.UcgConversationMember.Ctx(ctx).
		Where(dao.UcgConversationMember.Columns().ConversationId, convID).
		WhereNot(dao.UcgConversationMember.Columns().WxId, wxID).One()
	if err != nil {
		return 0, err
	}
	if row.IsEmpty() {
		return 0, gerror.NewCode(gcode.CodeNotFound, "会话成员不存在")
	}
	var m entity.UcgConversationMember
	if err = row.Struct(&m); err != nil {
		return 0, err
	}
	return m.WxId, nil
}

func lastMessagePreview(ctx context.Context, convID uint64) string {
	_, msgs, err := listChatMessages(ctx, convID, 1, 1)
	if err != nil || len(msgs) == 0 {
		return ""
	}
	last := msgs[len(msgs)-1]
	if strings.TrimSpace(last.VideoKey) != "" {
		if t := strings.TrimSpace(last.Content); t != "" {
			return "[视频] " + previewTrim(t, 48)
		}
		return "[视频]"
	}
	if strings.TrimSpace(last.ImageKey) != "" {
		if t := strings.TrimSpace(last.Content); t != "" {
			return "[图片] " + previewTrim(t, 48)
		}
		return "[图片]"
	}
	return previewTrim(last.Content, 64)
}

func previewTrim(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	if len([]rune(s)) > maxRunes {
		return string([]rune(s)[:maxRunes])
	}
	return s
}

// ListConversationMessages 会话消息分页（Redis）。
func ListConversationMessages(ctx context.Context, wxID int64, convID uint64, page, pageSize int) (*PageResult, error) {
	if err := ensureConversationMember(ctx, convID, wxID); err != nil {
		return nil, err
	}
	total, msgs, err := listChatMessages(ctx, convID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &PageResult{List: msgs, Total: total, Page: NormalizePage(page, pageSize).Page, PageSize: NormalizePage(page, pageSize).PageSize}, nil
}

// MarkConversationRead 标记已读。
func MarkConversationRead(ctx context.Context, wxID int64, convID uint64, lastMsgID uint64) error {
	if err := ensureConversationMember(ctx, convID, wxID); err != nil {
		return err
	}
	_ = resetUnread(ctx, convID, wxID)
	data := g.Map{
		dao.UcgConversationMember.Columns().UnreadCount: 0,
	}
	if lastMsgID > 0 {
		data[dao.UcgConversationMember.Columns().LastReadMsgId] = lastMsgID
	}
	_, err := dao.UcgConversationMember.Ctx(ctx).
		Where(dao.UcgConversationMember.Columns().ConversationId, convID).
		Where(dao.UcgConversationMember.Columns().WxId, wxID).
		Data(data).Update()
	return err
}

// SetConversationPinned 置顶/取消置顶。
func SetConversationPinned(ctx context.Context, wxID int64, convID uint64, pinned bool) error {
	if err := ensureConversationMember(ctx, convID, wxID); err != nil {
		return err
	}
	p := 0
	if pinned {
		p = 1
	}
	_, err := dao.UcgConversationMember.Ctx(ctx).
		Where(dao.UcgConversationMember.Columns().ConversationId, convID).
		Where(dao.UcgConversationMember.Columns().WxId, wxID).
		Data(g.Map{
			dao.UcgConversationMember.Columns().Pinned:     p,
			dao.UcgConversationMember.Columns().UpdatedAt: time.Now().Unix(),
		}).Update()
	return err
}

// SoftDeleteConversation 用户侧软删除会话。
func SoftDeleteConversation(ctx context.Context, wxID int64, convID uint64) error {
	if err := ensureConversationMember(ctx, convID, wxID); err != nil {
		return err
	}
	_, err := dao.UcgConversationMember.Ctx(ctx).
		Where(dao.UcgConversationMember.Columns().ConversationId, convID).
		Where(dao.UcgConversationMember.Columns().WxId, wxID).
		Data(g.Map{
			dao.UcgConversationMember.Columns().DeletedAt: time.Now().Unix(),
		}).Update()
	return err
}

func ensureConversationMember(ctx context.Context, convID uint64, wxID int64) error {
	cnt, err := dao.UcgConversationMember.Ctx(ctx).
		Where(dao.UcgConversationMember.Columns().ConversationId, convID).
		Where(dao.UcgConversationMember.Columns().WxId, wxID).
		Where(dao.UcgConversationMember.Columns().DeletedAt, 0).
		Count()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return gerror.NewCode(gcode.CodeNotFound, "会话不存在")
	}
	return nil
}

func bumpMemberActivity(ctx context.Context, convID uint64, wxIDs ...int64) {
	now := time.Now().Unix()
	for _, id := range wxIDs {
		_, _ = dao.UcgConversationMember.Ctx(ctx).
			Where(dao.UcgConversationMember.Columns().ConversationId, convID).
			Where(dao.UcgConversationMember.Columns().WxId, id).
			Data(g.Map{dao.UcgConversationMember.Columns().UpdatedAt: now}).Update()
		_ = touchUserConversation(ctx, id, convID, now)
	}
	_, _ = dao.UcgConversation.Ctx(ctx).Where(dao.UcgConversation.Columns().Id, convID).
		Data(g.Map{dao.UcgConversation.Columns().UpdatedAt: now}).Update()
}

// DeliverChatMessage Green 通过后持久化并推送（Option C 投递阶段）。
func DeliverChatMessage(ctx context.Context, convID uint64, senderWxID, recipientWxID int64, clientMsgID, content, imageKey, videoKey, mediaCdnURL string) (ChatMessage, error) {
	msg, err := appendChatMessage(ctx, convID, ChatMessage{
		ClientMsgID: clientMsgID,
		SenderWxID:  senderWxID,
		Content:     content,
		ImageKey:    imageKey,
		VideoKey:    videoKey,
		MediaCdnUrl: mediaCdnURL,
		Status:      "delivered",
	})
	if err != nil {
		return msg, err
	}
	_ = incrUnread(ctx, convID, recipientWxID)
	bumpMemberActivity(ctx, convID, senderWxID, recipientWxID)
	_, _ = dao.UcgConversationMember.Ctx(ctx).
		Where(dao.UcgConversationMember.Columns().ConversationId, convID).
		Where(dao.UcgConversationMember.Columns().WxId, recipientWxID).
		Increment(dao.UcgConversationMember.Columns().UnreadCount, 1)
	payload := map[string]interface{}{
		"type":           "message_delivered",
		"conversationId": convID,
		"message":        msg,
	}
	ChatWSHub().SendJSON(recipientWxID, payload)
	return msg, nil
}

// ProcessOutboundChatMessage 发送消息：先 ACK，再 Green，通过后投递。
func ProcessOutboundChatMessage(ctx context.Context, senderWxID int64, convID uint64, clientMsgID, content, imageKey, videoKey string) error {
	content = strings.TrimSpace(content)
	imageKey = strings.TrimSpace(imageKey)
	videoKey = strings.TrimSpace(videoKey)
	if content == "" && imageKey == "" && videoKey == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "content 或媒体 attachment 必填")
	}
	if imageKey != "" && videoKey != "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "imageKey 与 videoKey 不可同时存在")
	}
	if err := ensureConversationMember(ctx, convID, senderWxID); err != nil {
		return err
	}
	recipient, err := peerWxID(ctx, convID, senderWxID)
	if err != nil {
		return err
	}
	ChatWSHub().SendJSON(senderWxID, map[string]interface{}{
		"type":        "message_ack",
		"clientMsgId": clientMsgID,
	})
	moderator := EffectiveGreen()
	if content != "" {
		verdict, mErr := moderator.ModerateText(ctx, "comment_detection", content)
		if mErr != nil {
			return mErr
		}
		if !verdict.Pass {
			sendChatAuditFailed(senderWxID, clientMsgID, verdict.Reason)
			return nil
		}
	}
	cfg := LoadOSSConfig(ctx)
	var mediaCdnURL string
	if imageKey != "" {
		mediaCdnURL = cfg.CdnBaseURL + "/" + strings.TrimPrefix(imageKey, "/")
		if verdict, mErr := moderator.ModerateImageURL(ctx, mediaCdnURL); mErr != nil {
			return mErr
		} else if !verdict.Pass {
			sendChatAuditFailed(senderWxID, clientMsgID, verdict.Reason)
			return nil
		}
	}
	if videoKey != "" {
		mediaCdnURL = cfg.CdnBaseURL + "/" + strings.TrimPrefix(videoKey, "/")
		if verdict, mErr := moderator.ModerateVideoURL(ctx, mediaCdnURL); mErr != nil {
			return mErr
		} else if !verdict.Pass {
			sendChatAuditFailed(senderWxID, clientMsgID, verdict.Reason)
			return nil
		}
	}
	_, err = DeliverChatMessage(ctx, convID, senderWxID, int64(recipient), clientMsgID, content, imageKey, videoKey, mediaCdnURL)
	return err
}

func sendChatAuditFailed(wxID int64, clientMsgID, reason string) {
	if reason == "" {
		reason = rejectReasonDefault
	}
	ChatWSHub().SendJSON(wxID, map[string]interface{}{
		"type":        "audit_failed",
		"clientMsgId": clientMsgID,
		"reason":      reason,
	})
}
