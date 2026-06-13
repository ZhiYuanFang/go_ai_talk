package ucg

import (
	"context"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// auditCommentFromEvent 评论 MQ 审核：两阶段 Green → apply CAS。
func auditCommentFromEvent(ctx context.Context, commentID uint64, auditVersion int) error {
	comment, err := loadCommentForAudit(ctx, commentID)
	if err != nil {
		return err
	}
	if comment.Id == 0 {
		glog.Infof(ctx, "[ucg-audit-mq] comment skip missing id=%d version=%d", commentID, auditVersion)
		return nil
	}
	if comment.Status != CommentStatusPendingAudit || comment.AuditVersion != auditVersion {
		glog.Infof(ctx, "[ucg-audit-mq] comment skip stale id=%d curStatus=%d curVersion=%d eventVersion=%d",
			commentID, comment.Status, comment.AuditVersion, auditVersion)
		return nil
	}

	runCommentModerationPhase(ctx, ucgCommentQueue, comment, auditVersion)
	comment, err = loadCommentForAudit(ctx, commentID)
	if err != nil {
		return err
	}
	return runCommentApplyPhase(ctx, ucgCommentQueue, comment, auditVersion)
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
