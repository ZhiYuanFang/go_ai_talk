package ucg

import (
	"context"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// auditPostFromEvent MQ 消费者：两阶段 Green 机审 → apply CAS（不递增 audit_version）。
func auditPostFromEvent(ctx context.Context, postID uint64, auditVersion int) error {
	post, err := loadPostForAudit(ctx, postID)
	if err != nil {
		return err
	}
	if post.Id == 0 {
		glog.Infof(ctx, "[ucg-audit-mq] post skip missing id=%d version=%d", postID, auditVersion)
		return nil
	}
	if post.Status != PostStatusPendingAudit || post.AuditVersion != auditVersion {
		glog.Infof(ctx, "[ucg-audit-mq] post skip stale id=%d curStatus=%d curVersion=%d eventVersion=%d",
			postID, post.Status, post.AuditVersion, auditVersion)
		return nil
	}

	if err = runPostModerationPhase(ctx, ucgPostQueue, post, auditVersion); err != nil {
		return err
	}
	post, err = loadPostForAudit(ctx, postID)
	if err != nil {
		return err
	}
	return runPostApplyPhase(ctx, ucgPostQueue, post, auditVersion)
}

func publishPostCAS(ctx context.Context, postID uint64, auditVersion int) error {
	now := time.Now().Unix()
	affected, err := CasAuditTransition(ctx, CasAuditInput{
		Table:       dao.UcgPost.Table(),
		ID:          postID,
		Kind:        AuditCasKindStatus,
		FromStatus:  PostStatusPendingAudit,
		ToStatus:    PostStatusPublished,
		FromVersion: auditVersion,
		Extra: g.Map{
			dao.UcgPost.Columns().PublishedAt:  now,
			dao.UcgPost.Columns().UpdatedAt:    now,
			dao.UcgPost.Columns().RejectReason: "",
		},
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		glog.Infof(ctx, "[ucg-audit-mq] post publish cas skip id=%d version=%d", postID, auditVersion)
		return nil
	}
	PublishPostPublished(ctx, postID)
	return nil
}

// rejectPostCAS 机审/管理端驳回：CAS 带 audit_version。
func rejectPostCAS(ctx context.Context, postID uint64, auditVersion int, reason string) error {
	if reason == "" {
		reason = rejectReasonDefault
	}
	now := time.Now().Unix()
	affected, err := CasAuditTransition(ctx, CasAuditInput{
		Table:       dao.UcgPost.Table(),
		ID:          postID,
		Kind:        AuditCasKindStatus,
		FromStatus:  PostStatusPendingAudit,
		ToStatus:    PostStatusRejected,
		FromVersion: auditVersion,
		Extra: g.Map{
			dao.UcgPost.Columns().RejectReason: reason,
			dao.UcgPost.Columns().UpdatedAt:    now,
		},
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		glog.Infof(ctx, "[ucg-audit-mq] post reject cas skip id=%d version=%d", postID, auditVersion)
	}
	return nil
}

// rejectPostByIDAdmin 管理端驳回：允许 pending 或 published，仍带 audit_version CAS。
func rejectPostByIDAdmin(ctx context.Context, post entity.UcgPost, reason string) error {
	if reason == "" {
		reason = rejectReasonDefault
	}
	ver := post.AuditVersion
	if ver <= 0 {
		ver = 1
	}
	now := time.Now().Unix()
	wasPublished := post.Status == PostStatusPublished
	affected, err := CasAuditTransitionInStatuses(ctx, dao.UcgPost.Table(), post.Id,
		[]int{PostStatusPendingAudit, PostStatusPublished}, ver, PostStatusRejected, g.Map{
			dao.UcgPost.Columns().RejectReason: reason,
			dao.UcgPost.Columns().UpdatedAt:    now,
		})
	if err != nil {
		return err
	}
	if affected == 0 {
		glog.Infof(ctx, "[ucg-audit-mq] admin reject cas skip id=%d version=%d", post.Id, ver)
		return nil
	}
	if wasPublished {
		PublishPostUnpublished(ctx, post.Id)
	}
	return nil
}
