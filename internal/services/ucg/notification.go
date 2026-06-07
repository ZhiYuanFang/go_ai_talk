package ucg

import (
	"context"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const (
	NotificationTypeCommentOnPost   = "comment_on_post"
	NotificationTypeMentionInComment = "mention_in_comment"
)

var mentionPattern = regexp.MustCompile(`@([\p{Han}\w]+)`)

// NotificationDTO inbox 通知视图。
type NotificationDTO struct {
	Id        uint64       `json:"id"`
	Type      string       `json:"type"`
	PostId    uint64       `json:"postId"`
	CommentId uint64       `json:"commentId"`
	Actor     *ProfileDTO  `json:"actor,omitempty"`
	Preview   string       `json:"preview"`
	Read      bool         `json:"read"`
	CreatedAt int64        `json:"createdAt"`
}

// NotificationPageResult 分页 + 未读计数。
type NotificationPageResult struct {
	List        []*NotificationDTO `json:"list"`
	Total       int                `json:"total"`
	Page        int                `json:"page"`
	PageSize    int                `json:"pageSize"`
	UnreadCount int                `json:"unreadCount"`
}

func truncatePreview(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	if utf8.RuneCountInString(content) <= 200 {
		return content
	}
	return string([]rune(content)[:200])
}

// InsertNotification 写入一条 inbox 通知并可选 WS 推送。
func InsertNotification(ctx context.Context, recipientWxID int64, nType string, postID, commentID, actorWxID uint64, preview string) (uint64, error) {
	if recipientWxID <= 0 || postID == 0 || commentID == 0 || actorWxID == 0 {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "通知参数无效")
	}
	now := time.Now().Unix()
	res, err := dao.UcgNotification.Ctx(ctx).Data(g.Map{
		dao.UcgNotification.Columns().RecipientWxId: recipientWxID,
		dao.UcgNotification.Columns().Type:          nType,
		dao.UcgNotification.Columns().PostId:        postID,
		dao.UcgNotification.Columns().CommentId:     commentID,
		dao.UcgNotification.Columns().ActorWxId:     actorWxID,
		dao.UcgNotification.Columns().Preview:       truncatePreview(preview),
		dao.UcgNotification.Columns().CreatedAt:     now,
	}).Insert()
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	notifyID := uint64(id)
	ChatWSHub().SendJSON(recipientWxID, map[string]interface{}{
		"type":           "comment_notification",
		"notificationId": notifyID,
	})
	return notifyID, nil
}

// CountUnreadNotifications 未读互动消息数。
func CountUnreadNotifications(ctx context.Context, recipientWxID int64) (int, error) {
	if recipientWxID <= 0 {
		return 0, nil
	}
	n, err := dao.UcgNotification.Ctx(ctx).
		Where(dao.UcgNotification.Columns().RecipientWxId, recipientWxID).
		WhereNull(dao.UcgNotification.Columns().ReadAt).
		Count()
	return int(n), err
}

// ListCommentNotifications 分页拉取评论/@ 通知。
func ListCommentNotifications(ctx context.Context, recipientWxID int64, page, pageSize int) (*NotificationPageResult, error) {
	if recipientWxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	p := NormalizePage(page, pageSize)
	model := dao.UcgNotification.Ctx(ctx).Where(dao.UcgNotification.Columns().RecipientWxId, recipientWxID)
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	unread, err := CountUnreadNotifications(ctx, recipientWxID)
	if err != nil {
		return nil, err
	}
	rows, err := model.OrderDesc(dao.UcgNotification.Columns().CreatedAt).
		Limit(p.PageSize).Offset(pageOffset(p)).All()
	if err != nil {
		return nil, err
	}
	list := make([]*NotificationDTO, 0, len(rows))
	for _, row := range rows {
		var n entity.UcgNotification
		if err = row.Struct(&n); err != nil {
			return nil, err
		}
		dto := notificationToDTO(ctx, n)
		list = append(list, dto)
	}
	return &NotificationPageResult{
		List: list, Total: total, Page: p.Page, PageSize: p.PageSize, UnreadCount: unread,
	}, nil
}

func notificationToDTO(ctx context.Context, n entity.UcgNotification) *NotificationDTO {
	dto := &NotificationDTO{
		Id:        n.Id,
		Type:      n.Type,
		PostId:    n.PostId,
		CommentId: n.CommentId,
		Preview:   n.Preview,
		Read:      n.ReadAt > 0,
		CreatedAt: n.CreatedAt,
	}
	if prof, err := GetPublicProfile(ctx, n.ActorWxId); err == nil {
		dto.Actor = prof
	} else {
		dto.Actor = &ProfileDTO{WxId: n.ActorWxId}
	}
	return dto
}

// MarkNotificationsRead 按 ids 或全部标记已读。
func MarkNotificationsRead(ctx context.Context, recipientWxID int64, ids []uint64, all bool) error {
	if recipientWxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	if !all && len(ids) == 0 {
		return nil
	}
	now := time.Now().Unix()
	model := dao.UcgNotification.Ctx(ctx).
		Where(dao.UcgNotification.Columns().RecipientWxId, recipientWxID).
		WhereNull(dao.UcgNotification.Columns().ReadAt)
	if !all {
		model = model.WhereIn(dao.UcgNotification.Columns().Id, ids)
	}
	_, err := model.Data(g.Map{dao.UcgNotification.Columns().ReadAt: now}).Update()
	return err
}

// resolveNicknameToWxID 按昵称查 profile；重名时返回 0（跳过）。
func resolveNicknameToWxID(ctx context.Context, nickname string) uint64 {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return 0
	}
	rows, err := dao.UcgProfile.Ctx(ctx).
		Where(dao.UcgProfile.Columns().Nickname, nickname).
		Limit(2).
		All()
	if err != nil || len(rows) != 1 {
		return 0
	}
	var p entity.UcgProfile
	if err = rows[0].Struct(&p); err != nil {
		return 0
	}
	return p.WxId
}

// parseMentionNicknames 解析 @昵称 列表（去重保序）。
func parseMentionNicknames(content string) []string {
	matches := mentionPattern.FindAllStringSubmatch(content, -1)
	seen := make(map[string]struct{})
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		nick := strings.TrimSpace(m[1])
		if nick == "" {
			continue
		}
		if _, ok := seen[nick]; ok {
			continue
		}
		seen[nick] = struct{}{}
		out = append(out, nick)
	}
	return out
}

// NotifyOnComment AddComment 成功后：通知帖主 + @提及（Option A：仅 inbox，不发 DM）。
func NotifyOnComment(ctx context.Context, commenterWxID int64, postAuthorWxID uint64, postID, commentID uint64, content string) {
	preview := truncatePreview(content)
	actor := uint64(commenterWxID)

	if postAuthorWxID > 0 && int64(postAuthorWxID) != commenterWxID {
		if _, err := InsertNotification(ctx, int64(postAuthorWxID), NotificationTypeCommentOnPost, postID, commentID, actor, preview); err != nil {
			g.Log().Warningf(ctx, "[ucg-notification] comment_on_post insert failed post=%d comment=%d err=%v", postID, commentID, err)
		}
	}

	notified := make(map[int64]struct{})
	if postAuthorWxID > 0 {
		notified[int64(postAuthorWxID)] = struct{}{}
	}
	notified[commenterWxID] = struct{}{}

	for _, nick := range parseMentionNicknames(content) {
		wxID := resolveNicknameToWxID(ctx, nick)
		if wxID == 0 {
			continue
		}
		if _, skip := notified[int64(wxID)]; skip {
			continue
		}
		notified[int64(wxID)] = struct{}{}
		if _, err := InsertNotification(ctx, int64(wxID), NotificationTypeMentionInComment, postID, commentID, actor, preview); err != nil {
			g.Log().Warningf(ctx, "[ucg-notification] mention_in_comment insert failed recipient=%d err=%v", wxID, err)
		}
	}
}
