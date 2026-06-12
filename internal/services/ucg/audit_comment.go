package ucg

import (
	"context"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// auditCommentFromEvent MQ 消费者：评论 Green + CAS；通过后 increment count 与通知。
func auditCommentFromEvent(ctx context.Context, commentID uint64, auditVersion int) error {
	row, err := dao.UcgPostComment.Ctx(ctx).Where(dao.UcgPostComment.Columns().Id, commentID).One()
	if err != nil {
		return err
	}
	if row.IsEmpty() {
		glog.Infof(ctx, "[ucg-audit-mq] comment skip missing id=%d version=%d", commentID, auditVersion)
		return nil
	}
	var comment entity.UcgPostComment
	if err = row.Struct(&comment); err != nil {
		return err
	}
	if comment.Status != CommentStatusPendingAudit {
		glog.Infof(ctx, "[ucg-audit-mq] comment skip stale id=%d curStatus=%d eventVersion=%d", commentID, comment.Status, auditVersion)
		return nil
	}
	moderator := EffectiveGreen()
	if verdict, mErr := moderator.ModerateText(ctx, "comment_detection", comment.Content); mErr != nil {
		return mErr
	} else if !verdict.Pass {
		return rejectCommentCAS(ctx, commentID, auditVersion, verdict.Reason)
	}
	return publishCommentCAS(ctx, comment)
}

func publishCommentCAS(ctx context.Context, comment entity.UcgPostComment) error {
	now := time.Now().Unix()
	affected, err := CasAuditTransition(ctx, CasAuditInput{
		Table:       dao.UcgPostComment.Table(),
		ID:          comment.Id,
		Kind:        AuditCasKindStatus,
		FromStatus:  CommentStatusPendingAudit,
		ToStatus:    CommentStatusPublished,
		FromVersion: comment.AuditVersion,
		Extra: g.Map{
			dao.UcgPostComment.Columns().RejectReason: "",
		},
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		glog.Infof(ctx, "[ucg-audit-mq] comment cas skip id=%d version=%d", comment.Id, comment.AuditVersion)
		return nil
	}
	_, _ = dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, comment.PostId).Increment(dao.UcgPost.Columns().CommentCount, 1)
	PublishCommentPublished(ctx, comment.PostId, comment.Id)
	postRow, pErr := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, comment.PostId).One()
	if pErr == nil && !postRow.IsEmpty() {
		var post entity.UcgPost
		if sErr := postRow.Struct(&post); sErr == nil {
			NotifyOnComment(ctx, int64(comment.AuthorWxId), post.AuthorWxId, comment.PostId, comment.Id, comment.Content)
		}
	}
	_ = now
	return nil
}

func rejectCommentCAS(ctx context.Context, commentID uint64, auditVersion int, reason string) error {
	if reason == "" {
		reason = rejectReasonDefault
	}
	affected, err := CasAuditTransition(ctx, CasAuditInput{
		Table:       dao.UcgPostComment.Table(),
		ID:          commentID,
		Kind:        AuditCasKindStatus,
		FromStatus:  CommentStatusPendingAudit,
		ToStatus:    CommentStatusRejected,
		FromVersion: auditVersion,
		Extra: g.Map{
			dao.UcgPostComment.Columns().RejectReason: reason,
		},
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		glog.Infof(ctx, "[ucg-audit-mq] comment reject cas skip id=%d version=%d", commentID, auditVersion)
	}
	return nil
}
