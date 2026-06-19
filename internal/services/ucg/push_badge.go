package ucg

import (
	"context"
	"strings"

	"hello/internal/dao"

	"github.com/gogf/gf/v2/frame/g"
)

// ComputeTotalUnread returns Σ(conversation unread_count) + unread ucg_notification count.
func ComputeTotalUnread(ctx context.Context, wxID int64) (int, error) {
	if wxID <= 0 {
		return 0, nil
	}
	memberCols := dao.UcgConversationMember.Columns()
	sumVal, err := dao.UcgConversationMember.Ctx(ctx).
		Fields("COALESCE(SUM("+memberCols.UnreadCount+"),0) AS s").
		Where(memberCols.WxId, wxID).
		Where(memberCols.DeletedAt, 0).
		Value()
	if err != nil {
		return 0, err
	}
	chatUnread := sumVal.Int()
	notifUnread, err := CountUnreadNotifications(ctx, wxID)
	if err != nil {
		return 0, err
	}
	total := chatUnread + notifUnread
	if total < 0 {
		total = 0
	}
	return total, nil
}

// PushVisibleAlert sends a visible notification with absolute badge to all registered devices.
func PushVisibleAlert(ctx context.Context, recipientWxID int64, alertBody string) {
	if recipientWxID <= 0 || strings.TrimSpace(alertBody) == "" {
		return
	}
	asyncPush(recipientWxID, func(bg context.Context) {
		total, err := ComputeTotalUnread(bg, recipientWxID)
		if err != nil {
			g.Log().Warningf(bg, "[ucg-push] ComputeTotalUnread failed wxId=%d err=%v", recipientWxID, err)
			return
		}
		if total < 1 {
			total = 1
		}
		dispatchPush(bg, recipientWxID, PushPayload{
			Alert: alertBody,
			Badge: total,
			Silent: false,
		})
	})
}

// PushSilentBadge sends a silent badge-only update after read degradation.
func PushSilentBadge(ctx context.Context, recipientWxID int64) {
	if recipientWxID <= 0 {
		return
	}
	asyncPush(recipientWxID, func(bg context.Context) {
		total, err := ComputeTotalUnread(bg, recipientWxID)
		if err != nil {
			g.Log().Warningf(bg, "[ucg-push] ComputeTotalUnread failed wxId=%d err=%v", recipientWxID, err)
			return
		}
		if total < 0 {
			total = 0
		}
		dispatchPush(bg, recipientWxID, PushPayload{
			Badge:  total,
			Silent: true,
		})
	})
}

// PushVisibleDM notifies recipient of a new direct message.
func PushVisibleDM(ctx context.Context, recipientWxID, senderWxID int64) {
	nick := resolvePushNickname(ctx, senderWxID)
	PushVisibleAlert(ctx, recipientWxID, nick+"发来一条私信")
}

// PushVisibleComment notifies recipient of a new comment/mention notification.
func PushVisibleComment(ctx context.Context, recipientWxID, actorWxID int64) {
	nick := resolvePushNickname(ctx, actorWxID)
	PushVisibleAlert(ctx, recipientWxID, nick+"评论了你的动态")
}

func resolvePushNickname(ctx context.Context, wxID int64) string {
	if wxID <= 0 {
		return "有人"
	}
	if prof, err := GetPublicProfile(ctx, uint64(wxID)); err == nil && prof != nil {
		nick := strings.TrimSpace(prof.Nickname)
		if nick != "" {
			return nick
		}
	}
	return "有人"
}
