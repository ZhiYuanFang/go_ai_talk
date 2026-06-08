package ucg

import (
	"context"
	"regexp"
	"strconv"
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

// @nickname、@nickname#wxId、@wxId（纯数字）— 与客户端长按评论预填格式一致。
var mentionPattern = regexp.MustCompile(`@([^\s@]+?)(?:#(\d+))?`)

// NotificationDTO inbox 通知视图。
type NotificationDTO struct {
	Id            uint64      `json:"id"`
	Type          string      `json:"type"`
	PostId        uint64      `json:"postId"`
	CommentId     uint64      `json:"commentId"`
	Actor         *ProfileDTO `json:"actor,omitempty"`
	Preview       string      `json:"preview"`
	PostThumbUrl  string      `json:"postThumbUrl"`
	PostMediaKind int         `json:"postMediaKind"`
	Read          bool        `json:"read"`
	CreatedAt     int64       `json:"createdAt"`
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
func InsertNotification(ctx context.Context, recipientWxID int64, nType string, postID, commentID, actorWxID uint64, preview, postThumbURL string, postMediaKind int) (uint64, error) {
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
		dao.UcgNotification.Columns().PostThumbUrl:  strings.TrimSpace(postThumbURL),
		dao.UcgNotification.Columns().PostMediaKind: postMediaKind,
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
		Id:            n.Id,
		Type:          n.Type,
		PostId:        n.PostId,
		CommentId:     n.CommentId,
		Preview:       n.Preview,
		PostThumbUrl:  n.PostThumbUrl,
		PostMediaKind: n.PostMediaKind,
		Read:          n.ReadAt > 0,
		CreatedAt:     n.CreatedAt,
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

type mentionTarget struct {
	nickname string
	wxID     uint64
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

func verifyProfileWxID(ctx context.Context, wxID uint64) uint64 {
	if wxID == 0 {
		return 0
	}
	if _, err := GetPublicProfile(ctx, wxID); err != nil {
		return 0
	}
	return wxID
}

func resolveMentionWxID(ctx context.Context, target mentionTarget) uint64 {
	if target.wxID > 0 {
		if id := verifyProfileWxID(ctx, target.wxID); id > 0 {
			return id
		}
	}
	nick := strings.TrimSpace(target.nickname)
	if nick == "" {
		return 0
	}
	if id, err := strconv.ParseUint(nick, 10, 64); err == nil && id > 0 {
		if verified := verifyProfileWxID(ctx, id); verified > 0 {
			return verified
		}
	}
	return resolveNicknameToWxID(ctx, nick)
}

// parseMentionTargets 解析 @昵称 / @昵称#wxId / @wxId 列表（去重保序）。
func parseMentionTargets(content string) []mentionTarget {
	matches := mentionPattern.FindAllStringSubmatch(content, -1)
	seen := make(map[uint64]struct{})
	seenNick := make(map[string]struct{})
	out := make([]mentionTarget, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		nick := strings.TrimSpace(m[1])
		if nick == "" {
			continue
		}
		var wxID uint64
		if len(m) >= 3 && strings.TrimSpace(m[2]) != "" {
			if id, err := strconv.ParseUint(strings.TrimSpace(m[2]), 10, 64); err == nil {
				wxID = id
			}
		}
		if wxID > 0 {
			if _, ok := seen[wxID]; ok {
				continue
			}
			seen[wxID] = struct{}{}
			out = append(out, mentionTarget{nickname: nick, wxID: wxID})
			continue
		}
		if _, ok := seenNick[nick]; ok {
			continue
		}
		seenNick[nick] = struct{}{}
		out = append(out, mentionTarget{nickname: nick})
	}
	return out
}

// resolvePostCoverSnapshot 写入通知前一次 loadPostMedia，取首条媒体生成封面快照。
func resolvePostCoverSnapshot(ctx context.Context, postID uint64) (thumbURL string, mediaKind int) {
	media, err := loadPostMedia(ctx, postID)
	if err != nil || len(media) == 0 {
		return "", 0
	}
	first := media[0]
	key := strings.TrimSpace(first.ObjectKey)
	if key == "" {
		return "", 0
	}
	switch first.MediaKind {
	case 1:
		return BuildImageThumbnailURL(key), 1
	case 2:
		return BuildVideoSnapshotURL(key), 2
	default:
		return "", 0
	}
}

// NotifyOnComment AddComment 成功后：通知帖主 + @提及（Option A：仅 inbox，不发 DM）。
func NotifyOnComment(ctx context.Context, commenterWxID int64, postAuthorWxID uint64, postID, commentID uint64, content string) {
	preview := truncatePreview(content)
	actor := uint64(commenterWxID)
	thumbURL, mediaKind := resolvePostCoverSnapshot(ctx, postID)

	insert := func(recipient int64, nType string) {
		if _, err := InsertNotification(ctx, recipient, nType, postID, commentID, actor, preview, thumbURL, mediaKind); err != nil {
			g.Log().Warningf(ctx, "[ucg-notification] %s insert failed recipient=%d post=%d comment=%d err=%v", nType, recipient, postID, commentID, err)
		}
	}

	if postAuthorWxID > 0 && int64(postAuthorWxID) != commenterWxID {
		insert(int64(postAuthorWxID), NotificationTypeCommentOnPost)
	}

	notified := make(map[int64]struct{})
	if postAuthorWxID > 0 {
		notified[int64(postAuthorWxID)] = struct{}{}
	}
	notified[commenterWxID] = struct{}{}

	for _, target := range parseMentionTargets(content) {
		wxID := resolveMentionWxID(ctx, target)
		if wxID == 0 {
			continue
		}
		if _, skip := notified[int64(wxID)]; skip {
			continue
		}
		notified[int64(wxID)] = struct{}{}
		insert(int64(wxID), NotificationTypeMentionInComment)
	}
}
