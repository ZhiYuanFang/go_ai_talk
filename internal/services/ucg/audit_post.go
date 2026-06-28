package ucg

import (
	"context"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// auditPostFromEvent 帖子 MQ 审核入口：两阶段（Phase1 Green → Phase2 publish/reject CAS）。
// Phase1 失败写 moderation_failed 并 Ack，不再 requeue Green。
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

	runPostModerationPhase(ctx, ucgPostQueue, post, auditVersion)
	post, err = loadPostForAudit(ctx, postID)
	if err != nil {
		return err
	}
	return runPostApplyPhase(ctx, ucgPostQueue, post, auditVersion)
}

// publishPostCAS Phase2 通过：pending_audit → published。
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
		return err // CAS/DB 失败 → handlePostApplyFailure → 有界 requeue
	}
	if affected == 0 {
		glog.Infof(ctx, "[ucg-audit-mq] post publish cas skip id=%d version=%d", postID, auditVersion)
		return nil // 已被其他 delivery 处理
	}
	// 推荐分与 Redis 索引：publish 后 ZADD/GEO/snapshot。
	return syncPublishedPostRedis(ctx, postID)
}

// rejectPostCAS Phase2 驳回：pending_audit → rejected。
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

// rejectPostByIDAdmin 管理端驳回：允许 draft/pending/published/apply_failed/moderation_failed，仍带 audit_version CAS。
func rejectPostByIDAdmin(ctx context.Context, post entity.UcgPost, reason string) error {
	ver := post.AuditVersion
	if ver <= 0 {
		ver = 1
	}
	now := time.Now().Unix()
	wasPublished := post.Status == PostStatusPublished
	affected, err := CasAuditTransitionInStatuses(ctx, dao.UcgPost.Table(), post.Id,
		[]int{
			PostStatusDraft,
			PostStatusPendingAudit,
			PostStatusPublished,
			PostStatusApplyFailed,
			PostStatusModerationFailed,
		}, ver, PostStatusRejected, g.Map{
			dao.UcgPost.Columns().RejectReason: reason,
			dao.UcgPost.Columns().UpdatedAt:    now,
		})
	if err != nil {
		return err
	}
	if affected == 0 {
		glog.Infof(ctx, "[ucg-admin] post reject cas skip id=%d version=%d", post.Id, ver)
		return nil
	}
	if wasPublished {
		_ = RemoveRecommendScore(ctx, post.Id)
	}
	return nil
}

// approvePostByIDAdmin 管理端人工通过：pending/apply_failed/moderation_failed → published，不调 Green。
func approvePostByIDAdmin(ctx context.Context, post entity.UcgPost) error {
	ver := post.AuditVersion
	if ver <= 0 {
		ver = 1
	}
	now := time.Now().Unix()
	cols := dao.UcgPost.Columns()

	var fromStatus int
	var extra g.Map
	switch post.Status {
	case PostStatusPendingAudit:
		fromStatus = PostStatusPendingAudit
		extra = g.Map{
			cols.PublishedAt:  now,
			cols.UpdatedAt:    now,
			cols.RejectReason: "",
		}
		// 人工放行时若尚无机审 verdict，补写 pass 便于审计一致。
		if post.ModerationVerdict == ModerationVerdictNone {
			extra[cols.ModerationVerdict] = ModerationVerdictPass
			extra[cols.ModerationAt] = now
		}
	case PostStatusApplyFailed:
		fromStatus = PostStatusApplyFailed
		extra = g.Map{
			cols.PublishedAt:   now,
			cols.UpdatedAt:     now,
			cols.RejectReason:  "",
			cols.ApplyFailedAt: 0,
		}
	case PostStatusModerationFailed:
		fromStatus = PostStatusModerationFailed
		extra = g.Map{
			cols.PublishedAt:       now,
			cols.UpdatedAt:         now,
			cols.RejectReason:      "",
			cols.ModerationVerdict: ModerationVerdictPass,
			cols.ModerationAt:      now,
		}
	default:
		glog.Infof(ctx, "[ucg-admin] post approve unsupported status id=%d status=%d", post.Id, post.Status)
		return gerror.NewCodef(gcode.CodeInvalidParameter, "status=%d 不可人工通过", post.Status)
	}

	affected, err := CasAuditTransition(ctx, CasAuditInput{
		Table:       dao.UcgPost.Table(),
		ID:          post.Id,
		Kind:        AuditCasKindStatus,
		FromStatus:  fromStatus,
		ToStatus:    PostStatusPublished,
		FromVersion: ver,
		Extra:       extra,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		glog.Infof(ctx, "[ucg-admin] post approve cas skip id=%d version=%d status=%d", post.Id, ver, post.Status)
		return nil
	}
	glog.Infof(ctx, "[ucg-admin] post approve ok id=%d version=%d", post.Id, ver)
	return syncPublishedPostRedis(ctx, post.Id)
}
