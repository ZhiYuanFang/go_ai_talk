package ucg

import (
	"context"
	"strings"

	pushclient "hello/internal/clients/push"
	"hello/internal/dao"

	"github.com/gogf/gf/v2/frame/g"
)

// ComputeTotalUnread returns Σ(conversation unread_count) + unread ucg_notification count.
// 业务：UCG 调用 push-service 时在请求中带绝对 badge。
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

// asyncPushClient 异步经 clients/push 下发，避免阻塞 UCG 写路径。
func asyncPushClient(recipientWxID int64, fn func(context.Context)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// best-effort；不得拖垮进程
			}
		}()
		fn(context.Background())
	}()
	_ = recipientWxID
}

// PushVisibleAlert 可见通知；badge 为本域未读总数（至少 1）。
func PushVisibleAlert(ctx context.Context, recipientWxID int64, alertBody string) {
	if recipientWxID <= 0 || strings.TrimSpace(alertBody) == "" {
		return
	}
	alert := strings.TrimSpace(alertBody)
	asyncPushClient(recipientWxID, func(bg context.Context) {
		total, err := ComputeTotalUnread(bg, recipientWxID)
		if err != nil {
			g.Log().Warningf(bg, "[ucg-push] ComputeTotalUnread failed wxId=%d err=%v", recipientWxID, err)
			return
		}
		if total < 1 {
			total = 1
		}
		if err = pushclient.PushByBizType(bg, recipientWxID, pushclient.BizUcgAlert, alert, total, false, nil); err != nil {
			g.Log().Warningf(bg, "[ucg-push] visible alert failed wxId=%d err=%v", recipientWxID, err)
		}
	})
}

// PushSilentBadge 静默角标更新（读后降级）。
func PushSilentBadge(ctx context.Context, recipientWxID int64) {
	if recipientWxID <= 0 {
		return
	}
	asyncPushClient(recipientWxID, func(bg context.Context) {
		total, err := ComputeTotalUnread(bg, recipientWxID)
		if err != nil {
			g.Log().Warningf(bg, "[ucg-push] ComputeTotalUnread failed wxId=%d err=%v", recipientWxID, err)
			return
		}
		if total < 0 {
			total = 0
		}
		if err = pushclient.PushByBizType(bg, recipientWxID, pushclient.BizUcgSilentBadge, "", total, true, nil); err != nil {
			g.Log().Warningf(bg, "[ucg-push] silent badge failed wxId=%d err=%v", recipientWxID, err)
		}
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
