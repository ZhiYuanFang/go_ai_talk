// Package ucg 资料/帖子内容审核两阶段编排（Phase1 Green 机审 → Phase2 apply 落库）。
//
// 【Green 风暴排查 — 对照日志与 DB moderation_verdict / moderation_failed 终态】
//
// 路径 A（Phase1 Green/API 失败 — 已修复，不再 requeue Green）：
//
//	run*GreenChecks 返回 err → mark*ModerationFailed → status=moderation_failed → handler Ack
//	→ 后续 delivery 因 status≠pending 跳过 Green
//
// 路径 B（Phase1 成功，Phase2 apply 失败，有界重试、不再调 Green）：
//
//	persistModerationVerdictProfile 写入 verdict=1/2
//	→ runProfileApplyPhase / approveProfileJobCAS 失败
//	→ handleProfileApplyFailure：attempts<max 返回 applyErr（requeue 仅 retry apply）
//	→ attempts>=max 返回 nil（Ack，标记 apply_failed）
//
// 路径 C（enabled=false / noop）：不调 Green，直接 pass 并 persist → 不会风暴。
package ucg

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// ucgAuditApplyMaxAttemptsEnv apply 阶段最大重试次数；Phase1 Green 无对应上限 env。
const ucgAuditApplyMaxAttemptsEnv = "UCG_AUDIT_APPLY_MAX_ATTEMPTS"

const defaultAuditApplyMaxAttempts = 5

// auditApplyMaxAttempts 读取 apply 重试上限；仅 Phase2 使用，Phase1 Green 失败不受此限制。
func auditApplyMaxAttempts() int {
	v := strings.TrimSpace(os.Getenv(ucgAuditApplyMaxAttemptsEnv))
	if v == "" {
		return defaultAuditApplyMaxAttempts
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return defaultAuditApplyMaxAttempts
	}
	return n
}

// persistModerationVerdictProfile Phase1 落库：将 Green 结论写入 job.moderation_verdict。
// 成功写入后 MQ 重投时 runProfileModerationPhase 会 skip Green。
// 返回 error → 上层 handler error → MQ Nack requeue；若 verdict 仍为 0 则下次再调 Green。
func persistModerationVerdictProfile(ctx context.Context, jobID uint64, auditVersion, verdict int, reason string) error {
	now := time.Now().Unix()
	cols := dao.UcgProfileAuditJob.Columns()
	// CAS：仅 pending + 匹配 audit_version + moderation_verdict=0 时可写，避免并发双调 Green
	result, err := dao.UcgProfileAuditJob.Ctx(ctx).
		Where(cols.Id, jobID).
		Where(cols.Status, ProfileJobStatusPending).
		Where(cols.AuditVersion, auditVersion).
		Where(cols.ModerationVerdict, ModerationVerdictNone).
		Data(g.Map{
			cols.ModerationVerdict: verdict,
			cols.ModerationReason:  reason,
			cols.ModerationAt:      now,
			cols.UpdatedAt:         now,
		}).Update()
	if err != nil {
		// DB 错误（如缺列）→ 写不进 verdict → 路径 A 风暴
		return err
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		return nil // 首次写入 verdict 成功
	}
	// 并发另一 consumer 可能已写入，读回确认
	var job entity.UcgProfileAuditJob
	if scanErr := dao.UcgProfileAuditJob.Ctx(ctx).Where(cols.Id, jobID).Scan(&job); scanErr != nil {
		return scanErr
	}
	if job.ModerationVerdict == ModerationVerdictNone {
		// status/version 不匹配导致 UPDATE 0 行且 verdict 仍 0 → requeue 后会再调 Green
		return fmt.Errorf("profile moderation verdict cas lost id=%d version=%d", jobID, auditVersion)
	}
	return nil // 他人已写入 verdict，本 delivery 可继续 Phase2
}

// persistModerationVerdictPost 帖子 Phase1 落库，语义同 persistModerationVerdictProfile。
func persistModerationVerdictPost(ctx context.Context, postID uint64, auditVersion, verdict int, reason string) error {
	now := time.Now().Unix()
	cols := dao.UcgPost.Columns()
	result, err := dao.UcgPost.Ctx(ctx).
		Where(cols.Id, postID).
		Where(cols.Status, PostStatusPendingAudit).
		Where(cols.AuditVersion, auditVersion).
		Where(cols.ModerationVerdict, ModerationVerdictNone).
		Data(g.Map{
			cols.ModerationVerdict: verdict,
			cols.ModerationReason:  reason,
			cols.ModerationAt:      now,
			cols.UpdatedAt:         now,
		}).Update()
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		return nil
	}
	var post entity.UcgPost
	if scanErr := dao.UcgPost.Ctx(ctx).Where(cols.Id, postID).Scan(&post); scanErr != nil {
		return scanErr
	}
	if post.ModerationVerdict == ModerationVerdictNone {
		return fmt.Errorf("post moderation verdict cas lost id=%d version=%d", postID, auditVersion)
	}
	return nil
}

// runProfileGreenChecks 资料 Phase1：按 job 非空字段依次调 Green（串行，任一步 err 则整段失败不落库）。
// 与 ucg_profile 已发布值相同的字段跳过（兼容 MQ 积压全量 job + 省 API）。
// 返回 err → 路径 A；返回 pass=false → 会 persist reject，不会无限 Green。
func runProfileGreenChecks(ctx context.Context, job entity.UcgProfileAuditJob) (pass bool, reason string, err error) {
	moderator := EffectiveGreen() // enabled=false 时为 noop，不调阿里云
	cfg := LoadOSSConfig(ctx)
	var published entity.UcgProfile
	_ = dao.UcgProfile.Ctx(ctx).Where(dao.UcgProfile.Columns().WxId, job.WxId).Scan(&published)
	pubNick := strings.TrimSpace(published.Nickname)
	pubBio := strings.TrimSpace(published.Bio)
	pubAvatar := strings.TrimSpace(published.AvatarKey)

	nick := strings.TrimSpace(job.Nickname)
	if nick != "" && nick != pubNick {
		var verdict AuditVerdict
		verdict, err = moderator.ModerateText(ctx, "nickname_detection", nick)
		if err != nil {
			// SDK/额度/body.code≠200 → err，verdict 不写库 → 风暴
			return false, "", err
		}
		if !verdict.Pass {
			return false, verdict.Reason, nil // 违规：后续 persist reject，handler 成功 Ack
		}
	}
	bioText := strings.TrimSpace(job.Bio)
	if bioText != "" && bioText != pubBio {
		var verdict AuditVerdict
		verdict, err = moderator.ModerateText(ctx, "comment_detection", bioText)
		if err != nil {
			return false, "", err
		}
		if !verdict.Pass {
			return false, verdict.Reason, nil
		}
	}
	avatarKey := strings.TrimSpace(job.AvatarKey)
	if avatarKey != "" && avatarKey != pubAvatar {
		url := cfg.CdnBaseURL + "/" + avatarKey
		var verdict AuditVerdict
		verdict, err = moderator.ModerateImageURL(ctx, url)
		if err != nil {
			return false, "", err
		}
		if !verdict.Pass {
			return false, verdict.Reason, nil
		}
	}
	return true, "", nil // 全部通过或跳过
}

// runPostGreenChecks 帖子 Phase1 Green；结构同资料，无「对比已发布 profile」跳过逻辑。
func runPostGreenChecks(ctx context.Context, post entity.UcgPost) (pass bool, reason string, err error) {
	moderator := EffectiveGreen()
	cfg := LoadOSSConfig(ctx)
	if verdict, mErr := moderator.ModerateText(ctx, "comment_detection", post.Content); mErr != nil {
		return false, "", mErr
	} else if !verdict.Pass {
		return false, verdict.Reason, nil
	}
	media, err := loadPostMedia(ctx, post.Id)
	if err != nil {
		return false, "", err
	}
	for _, m := range media {
		url := m.CdnUrl
		if url == "" && m.ObjectKey != "" {
			url = cfg.CdnBaseURL + "/" + m.ObjectKey
		}
		var verdict AuditVerdict
		switch {
		case m.MediaKind == 2:
			verdict, err = moderator.ModerateVideoURL(ctx, url)
		default:
			verdict, err = moderator.ModerateImageURL(ctx, url)
		}
		if err != nil {
			return false, "", err
		}
		if !verdict.Pass {
			return false, verdict.Reason, nil
		}
	}
	return true, "", nil
}

// runProfileModerationPhase Phase1 入口：verdict≠0 则 skip Green；否则 Green + persist。
// 调用内容安全检查 API 进行审核
func runProfileModerationPhase(ctx context.Context, queueName string, job entity.UcgProfileAuditJob, auditVersion int) {
	// 如果 verdict 已落库，则跳过 Green
	if job.ModerationVerdict != ModerationVerdictNone {
		// 【关键】verdict 已落库 → 重投不再调 Green（路径 B 的 retry apply）
		glog.Infof(ctx, "[ucg-audit-mq] profile moderation skip green queue=%s jobId=%d wxId=%d auditVersion=%d verdict=%d apply_attempts=%d",
			queueName, job.Id, job.WxId, auditVersion, job.ModerationVerdict, job.ApplyAttempts)
		return
	}
	pass, reason, err := runProfileGreenChecks(ctx, job)
	if err != nil {
		fromReason := fmt.Sprintf("[ucg-audit-mq] profile moderation green checks failed jobId=%d auditVersion=%d err=%v", job.Id, auditVersion, err)
		markProfileModerationFailed(ctx, job.Id, fromReason)
		return
	}
	verdict := ModerationVerdictPass
	if !pass {
		verdict = ModerationVerdictReject
		if reason == "" {
			reason = rejectReasonDefault
		}
	}
	if err := persistModerationVerdictProfile(ctx, job.Id, auditVersion, verdict, reason); err != nil {
		fromReason := fmt.Sprintf("[ucg-audit-mq] profile moderation persist verdict failed jobId=%d auditVersion=%d err=%v", job.Id, auditVersion, err)
		markProfileModerationFailed(ctx, job.Id, fromReason)
	}
}

func markProfileModerationFailed(ctx context.Context, jobID uint64, fromReason string) {
	now := time.Now().Unix()
	cols := dao.UcgProfileAuditJob.Columns()
	_, _ = dao.UcgProfileAuditJob.Ctx(ctx).
		Where(cols.Id, jobID).
		Where(cols.Status, ProfileJobStatusPending).
		Data(g.Map{
			cols.Status:       ProfileJobStatusModerationFailed,
			cols.RejectReason: fromReason,
			cols.UpdatedAt:    now,
		}).Update()
}

// runProfileApplyPhase Phase2：基于已持久化 verdict 执行 approve/reject CAS；失败走有界重试。
func runProfileApplyPhase(ctx context.Context, queueName string, job entity.UcgProfileAuditJob, auditVersion int) error {
	if job.Status != ProfileJobStatusPending {
		return nil // 已终态，Ack
	}
	// 如果 verdict 未落库，则返回错误
	if job.ModerationVerdict == ModerationVerdictNone {
		glog.Errorf(ctx, "[ucg-audit-mq] profile apply without verdict queue=%s jobId=%d auditVersion=%d", queueName, job.Id, auditVersion)
		return fmt.Errorf("profile apply: moderation verdict missing jobId=%d", job.Id)
	}
	var applyErr error
	if job.ModerationVerdict == ModerationVerdictReject {
		applyErr = rejectProfileJobCAS(ctx, job.Id, auditVersion, job.ModerationReason)
	} else {
		applyErr = approveProfileJobCAS(ctx, job, auditVersion) // 常见风暴历史原因：此处 DB 失败
	}
	if applyErr != nil {
		return handleProfileApplyFailure(ctx, queueName, job, auditVersion, applyErr)
	}
	return nil // 成功 → handler nil → Ack
}

// runPostModerationPhase 帖子 Phase1；Green/persist 失败写 moderation_failed 并 Ack，不再 requeue Green。
func runPostModerationPhase(ctx context.Context, queueName string, post entity.UcgPost, auditVersion int) {
	if post.ModerationVerdict != ModerationVerdictNone {
		glog.Infof(ctx, "[ucg-audit-mq] post moderation skip green queue=%s postId=%d authorWxId=%d auditVersion=%d verdict=%d apply_attempts=%d",
			queueName, post.Id, post.AuthorWxId, auditVersion, post.ModerationVerdict, post.ApplyAttempts)
		return
	}
	pass, reason, err := runPostGreenChecks(ctx, post)
	if err != nil {
		fromReason := fmt.Sprintf("[ucg-audit-mq] post moderation green checks failed postId=%d auditVersion=%d err=%v", post.Id, auditVersion, err)
		markPostModerationFailed(ctx, post.Id, auditVersion, fromReason)
		return
	}
	verdict := ModerationVerdictPass
	if !pass {
		verdict = ModerationVerdictReject
		if reason == "" {
			reason = rejectReasonDefault
		}
	}
	if err := persistModerationVerdictPost(ctx, post.Id, auditVersion, verdict, reason); err != nil {
		fromReason := fmt.Sprintf("[ucg-audit-mq] post moderation persist verdict failed postId=%d auditVersion=%d err=%v", post.Id, auditVersion, err)
		markPostModerationFailed(ctx, post.Id, auditVersion, fromReason)
	}
}

func markPostModerationFailed(ctx context.Context, postID uint64, auditVersion int, fromReason string) {
	now := time.Now().Unix()
	cols := dao.UcgPost.Columns()
	_, _ = dao.UcgPost.Ctx(ctx).
		Where(cols.Id, postID).
		Where(cols.Status, PostStatusPendingAudit).
		Where(cols.AuditVersion, auditVersion).
		Data(g.Map{
			cols.Status:        PostStatusModerationFailed,
			cols.RejectReason:  fromReason,
			cols.UpdatedAt:     now,
		}).Update()
}

// runPostApplyPhase 帖子 Phase2。
func runPostApplyPhase(ctx context.Context, queueName string, post entity.UcgPost, auditVersion int) error {
	if post.Status != PostStatusPendingAudit {
		return nil
	}
	if post.ModerationVerdict == ModerationVerdictNone {
		glog.Errorf(ctx, "[ucg-audit-mq] post apply without verdict queue=%s postId=%d auditVersion=%d", queueName, post.Id, auditVersion)
		return fmt.Errorf("post apply: moderation verdict missing postId=%d", post.Id)
	}
	var applyErr error
	if post.ModerationVerdict == ModerationVerdictReject {
		applyErr = rejectPostCAS(ctx, post.Id, auditVersion, post.ModerationReason)
	} else {
		applyErr = publishPostCAS(ctx, post.Id, auditVersion)
	}
	if applyErr != nil {
		return handlePostApplyFailure(ctx, queueName, post, auditVersion, applyErr)
	}
	return nil
}

func incrementProfileApplyAttempts(ctx context.Context, jobID uint64, auditVersion int) (int, error) {
	cols := dao.UcgProfileAuditJob.Columns()
	_, err := dao.UcgProfileAuditJob.Ctx(ctx).
		Where(cols.Id, jobID).
		Where(cols.AuditVersion, auditVersion).
		Increment(cols.ApplyAttempts, 1)
	if err != nil {
		return 0, err
	}
	var job entity.UcgProfileAuditJob
	if scanErr := dao.UcgProfileAuditJob.Ctx(ctx).Where(cols.Id, jobID).Scan(&job); scanErr != nil {
		return 0, scanErr
	}
	return job.ApplyAttempts, nil
}

func incrementPostApplyAttempts(ctx context.Context, postID uint64, auditVersion int) (int, error) {
	cols := dao.UcgPost.Columns()
	_, err := dao.UcgPost.Ctx(ctx).
		Where(cols.Id, postID).
		Where(cols.AuditVersion, auditVersion).
		Increment(cols.ApplyAttempts, 1)
	if err != nil {
		return 0, err
	}
	var post entity.UcgPost
	if scanErr := dao.UcgPost.Ctx(ctx).Where(cols.Id, postID).Scan(&post); scanErr != nil {
		return 0, scanErr
	}
	return post.ApplyAttempts, nil
}

func markProfileApplyFailed(ctx context.Context, jobID uint64, auditVersion int) error {
	now := time.Now().Unix()
	cols := dao.UcgProfileAuditJob.Columns()
	_, err := dao.UcgProfileAuditJob.Ctx(ctx).
		Where(cols.Id, jobID).
		Where(cols.Status, ProfileJobStatusPending).
		Where(cols.AuditVersion, auditVersion).
		Data(g.Map{
			cols.Status:        ProfileJobStatusApplyFailed,
			cols.RejectReason:  applyFailedSystemReason,
			cols.ApplyFailedAt: now,
			cols.UpdatedAt:     now,
		}).Update()
	return err
}

func markPostApplyFailed(ctx context.Context, postID uint64, auditVersion int) error {
	now := time.Now().Unix()
	cols := dao.UcgPost.Columns()
	_, err := dao.UcgPost.Ctx(ctx).
		Where(cols.Id, postID).
		Where(cols.Status, PostStatusPendingAudit).
		Where(cols.AuditVersion, auditVersion).
		Data(g.Map{
			cols.Status:        PostStatusApplyFailed,
			cols.RejectReason:  applyFailedSystemReason,
			cols.ApplyFailedAt: now,
			cols.UpdatedAt:     now,
		}).Update()
	return err
}

// handleProfileApplyFailure Phase2 失败：递增 apply_attempts；超限返回 nil（Ack 停止 requeue）。
// 未超限时返回 applyErr → Nack requeue，但 Phase1 已 skip Green。
func handleProfileApplyFailure(ctx context.Context, queueName string, job entity.UcgProfileAuditJob, auditVersion int, applyErr error) error {
	attempts, incErr := incrementProfileApplyAttempts(ctx, job.Id, auditVersion)
	if incErr != nil {
		return incErr
	}
	max := auditApplyMaxAttempts()
	if attempts >= max {
		if mErr := markProfileApplyFailed(ctx, job.Id, auditVersion); mErr != nil {
			glog.Errorf(ctx, "[ucg-audit-mq] profile apply failed mark err queue=%s jobId=%d auditVersion=%d attempts=%d err=%v",
				queueName, job.Id, auditVersion, attempts, mErr)
			return mErr
		}
		glog.Errorf(ctx, "[ucg-audit-mq] profile apply max exceeded queue=%s jobId=%d wxId=%d auditVersion=%d apply_attempts=%d apply_err=%v",
			queueName, job.Id, job.WxId, auditVersion, attempts, applyErr)
		return nil // 【关键】返回 nil → Ack，不再 requeue
	}
	glog.Warningf(ctx, "[ucg-audit-mq] profile apply retry queue=%s jobId=%d wxId=%d auditVersion=%d apply_attempts=%d err=%v",
		queueName, job.Id, job.WxId, auditVersion, attempts, applyErr)
	return applyErr // requeue，但 moderation_verdict 已有 → 不再调 Green
}

func handlePostApplyFailure(ctx context.Context, queueName string, post entity.UcgPost, auditVersion int, applyErr error) error {
	attempts, incErr := incrementPostApplyAttempts(ctx, post.Id, auditVersion)
	if incErr != nil {
		return incErr
	}
	max := auditApplyMaxAttempts()
	if attempts >= max {
		if mErr := markPostApplyFailed(ctx, post.Id, auditVersion); mErr != nil {
			glog.Errorf(ctx, "[ucg-audit-mq] post apply failed mark err queue=%s postId=%d auditVersion=%d attempts=%d err=%v",
				queueName, post.Id, auditVersion, attempts, mErr)
			return mErr
		}
		glog.Errorf(ctx, "[ucg-audit-mq] post apply max exceeded queue=%s postId=%d authorWxId=%d auditVersion=%d apply_attempts=%d apply_err=%v",
			queueName, post.Id, post.AuthorWxId, auditVersion, attempts, applyErr)
		return nil
	}
	glog.Warningf(ctx, "[ucg-audit-mq] post apply retry queue=%s postId=%d authorWxId=%d auditVersion=%d apply_attempts=%d err=%v",
		queueName, post.Id, post.AuthorWxId, auditVersion, attempts, applyErr)
	return applyErr
}

// --- 评论 Phase1/Phase2 ---

func persistModerationVerdictComment(ctx context.Context, commentID uint64, auditVersion, verdict int, reason string) error {
	now := time.Now().Unix()
	cols := dao.UcgPostComment.Columns()
	result, err := dao.UcgPostComment.Ctx(ctx).
		Where(cols.Id, commentID).
		Where(cols.Status, CommentStatusPendingAudit).
		Where(cols.AuditVersion, auditVersion).
		Where(cols.ModerationVerdict, ModerationVerdictNone).
		Data(g.Map{
			cols.ModerationVerdict: verdict,
			cols.ModerationReason:  reason,
			cols.ModerationAt:      now,
		}).Update()
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		return nil
	}
	var comment entity.UcgPostComment
	if scanErr := dao.UcgPostComment.Ctx(ctx).Where(cols.Id, commentID).Scan(&comment); scanErr != nil {
		return scanErr
	}
	if comment.ModerationVerdict == ModerationVerdictNone {
		return fmt.Errorf("comment moderation verdict cas lost id=%d version=%d", commentID, auditVersion)
	}
	return nil
}

func runCommentGreenChecks(ctx context.Context, comment entity.UcgPostComment) (pass bool, reason string, err error) {
	moderator := EffectiveGreen()
	verdict, mErr := moderator.ModerateText(ctx, "comment_detection", comment.Content)
	if mErr != nil {
		return false, "", mErr
	}
	if !verdict.Pass {
		return false, verdict.Reason, nil
	}
	return true, "", nil
}

func runCommentModerationPhase(ctx context.Context, queueName string, comment entity.UcgPostComment, auditVersion int) {
	if comment.ModerationVerdict != ModerationVerdictNone {
		glog.Infof(ctx, "[ucg-audit-mq] comment moderation skip green queue=%s commentId=%d auditVersion=%d verdict=%d",
			queueName, comment.Id, auditVersion, comment.ModerationVerdict)
		return
	}
	pass, reason, err := runCommentGreenChecks(ctx, comment)
	if err != nil {
		fromReason := fmt.Sprintf("[ucg-audit-mq] comment moderation green failed commentId=%d auditVersion=%d err=%v", comment.Id, auditVersion, err)
		markCommentModerationFailed(ctx, comment.Id, auditVersion, fromReason)
		return
	}
	verdict := ModerationVerdictPass
	if !pass {
		verdict = ModerationVerdictReject
		if reason == "" {
			reason = rejectReasonDefault
		}
	}
	if err := persistModerationVerdictComment(ctx, comment.Id, auditVersion, verdict, reason); err != nil {
		fromReason := fmt.Sprintf("[ucg-audit-mq] comment moderation persist failed commentId=%d auditVersion=%d err=%v", comment.Id, auditVersion, err)
		markCommentModerationFailed(ctx, comment.Id, auditVersion, fromReason)
	}
}

func markCommentModerationFailed(ctx context.Context, commentID uint64, auditVersion int, fromReason string) {
	cols := dao.UcgPostComment.Columns()
	_, _ = dao.UcgPostComment.Ctx(ctx).
		Where(cols.Id, commentID).
		Where(cols.Status, CommentStatusPendingAudit).
		Where(cols.AuditVersion, auditVersion).
		Data(g.Map{
			cols.Status:       CommentStatusModerationFailed,
			cols.RejectReason: fromReason,
		}).Update()
}

func runCommentApplyPhase(ctx context.Context, queueName string, comment entity.UcgPostComment, auditVersion int) error {
	if comment.Status != CommentStatusPendingAudit {
		return nil
	}
	if comment.ModerationVerdict == ModerationVerdictNone {
		glog.Errorf(ctx, "[ucg-audit-mq] comment apply without verdict queue=%s commentId=%d auditVersion=%d", queueName, comment.Id, auditVersion)
		return fmt.Errorf("comment apply: moderation verdict missing commentId=%d", comment.Id)
	}
	var applyErr error
	if comment.ModerationVerdict == ModerationVerdictReject {
		applyErr = rejectCommentCAS(ctx, comment.Id, auditVersion, comment.ModerationReason)
	} else {
		applyErr = publishCommentCAS(ctx, comment)
	}
	if applyErr != nil {
		return handleCommentApplyFailure(ctx, queueName, comment, auditVersion, applyErr)
	}
	return nil
}

func incrementCommentApplyAttempts(ctx context.Context, commentID uint64, auditVersion int) (int, error) {
	cols := dao.UcgPostComment.Columns()
	_, err := dao.UcgPostComment.Ctx(ctx).
		Where(cols.Id, commentID).
		Where(cols.AuditVersion, auditVersion).
		Increment(cols.ApplyAttempts, 1)
	if err != nil {
		return 0, err
	}
	var comment entity.UcgPostComment
	if scanErr := dao.UcgPostComment.Ctx(ctx).Where(cols.Id, commentID).Scan(&comment); scanErr != nil {
		return 0, scanErr
	}
	return comment.ApplyAttempts, nil
}

func markCommentApplyFailed(ctx context.Context, commentID uint64, auditVersion int) error {
	now := time.Now().Unix()
	cols := dao.UcgPostComment.Columns()
	_, err := dao.UcgPostComment.Ctx(ctx).
		Where(cols.Id, commentID).
		Where(cols.Status, CommentStatusPendingAudit).
		Where(cols.AuditVersion, auditVersion).
		Data(g.Map{
			cols.Status:        CommentStatusApplyFailed,
			cols.RejectReason:  applyFailedSystemReason,
			cols.ApplyFailedAt: now,
		}).Update()
	return err
}

func handleCommentApplyFailure(ctx context.Context, queueName string, comment entity.UcgPostComment, auditVersion int, applyErr error) error {
	attempts, incErr := incrementCommentApplyAttempts(ctx, comment.Id, auditVersion)
	if incErr != nil {
		return incErr
	}
	max := auditApplyMaxAttempts()
	if attempts >= max {
		if mErr := markCommentApplyFailed(ctx, comment.Id, auditVersion); mErr != nil {
			return mErr
		}
		glog.Errorf(ctx, "[ucg-audit-mq] comment apply max exceeded queue=%s commentId=%d auditVersion=%d attempts=%d err=%v",
			queueName, comment.Id, auditVersion, attempts, applyErr)
		return nil
	}
	glog.Warningf(ctx, "[ucg-audit-mq] comment apply retry queue=%s commentId=%d auditVersion=%d attempts=%d err=%v",
		queueName, comment.Id, auditVersion, attempts, applyErr)
	return applyErr
}

func loadCommentForAudit(ctx context.Context, commentID uint64) (entity.UcgPostComment, error) {
	var comment entity.UcgPostComment
	err := dao.UcgPostComment.Ctx(ctx).Where(dao.UcgPostComment.Columns().Id, commentID).Scan(&comment)
	return comment, err
}

// --- 私信 Phase1/Phase2 ---

func persistModerationVerdictChat(ctx context.Context, conversationID, messageID uint64, auditVersion, verdict int, reason string) error {
	now := time.Now().Unix()
	cols := dao.UcgChatMessage.Columns()
	result, err := dao.UcgChatMessage.Ctx(ctx).
		Where(cols.ConversationId, conversationID).
		Where(cols.Id, messageID).
		Where(cols.AuditStatus, ChatAuditStatusPending).
		Where(cols.AuditVersion, auditVersion).
		Where(cols.ModerationVerdict, ModerationVerdictNone).
		Data(g.Map{
			cols.ModerationVerdict: verdict,
			cols.ModerationReason:  reason,
			cols.ModerationAt:      now,
		}).Update()
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		return nil
	}
	var msg entity.UcgChatMessage
	if scanErr := dao.UcgChatMessage.Ctx(ctx).
		Where(cols.ConversationId, conversationID).
		Where(cols.Id, messageID).
		Scan(&msg); scanErr != nil {
		return scanErr
	}
	if msg.ModerationVerdict == ModerationVerdictNone {
		return fmt.Errorf("chat moderation verdict cas lost msgId=%d version=%d", messageID, auditVersion)
	}
	return nil
}

func runChatGreenChecks(ctx context.Context, msg entity.UcgChatMessage) (pass bool, reason string, err error) {
	moderator := EffectiveGreen()
	cfg := LoadOSSConfig(ctx)
	if msg.Content != "" {
		var verdict AuditVerdict
		verdict, err = moderator.ModerateText(ctx, "comment_detection", msg.Content)
		if err != nil {
			return false, "", err
		}
		if !verdict.Pass {
			return false, verdict.Reason, nil
		}
	}
	if msg.ImageKey != "" {
		url := msg.MediaCdnUrl
		if url == "" {
			url = cfg.CdnBaseURL + "/" + msg.ImageKey
		}
		var verdict AuditVerdict
		verdict, err = moderator.ModerateImageURL(ctx, url)
		if err != nil {
			return false, "", err
		}
		if !verdict.Pass {
			return false, verdict.Reason, nil
		}
	}
	if msg.VideoKey != "" {
		url := msg.MediaCdnUrl
		if url == "" {
			url = cfg.CdnBaseURL + "/" + msg.VideoKey
		}
		var verdict AuditVerdict
		verdict, err = moderator.ModerateVideoURL(ctx, url)
		if err != nil {
			return false, "", err
		}
		if !verdict.Pass {
			return false, verdict.Reason, nil
		}
	}
	return true, "", nil
}

func runChatModerationPhase(ctx context.Context, queueName string, msg entity.UcgChatMessage, auditVersion int) {
	if msg.ModerationVerdict != ModerationVerdictNone {
		glog.Infof(ctx, "[ucg-audit-mq] chat moderation skip green queue=%s msgId=%d auditVersion=%d verdict=%d",
			queueName, msg.Id, auditVersion, msg.ModerationVerdict)
		return
	}
	pass, reason, err := runChatGreenChecks(ctx, msg)
	if err != nil {
		fromReason := fmt.Sprintf("[ucg-audit-mq] chat moderation green failed msgId=%d auditVersion=%d err=%v", msg.Id, auditVersion, err)
		markChatModerationFailed(ctx, msg.ConversationId, msg.Id, auditVersion, fromReason)
		return
	}
	verdict := ModerationVerdictPass
	if !pass {
		verdict = ModerationVerdictReject
		if reason == "" {
			reason = rejectReasonDefault
		}
	}
	if err := persistModerationVerdictChat(ctx, msg.ConversationId, msg.Id, auditVersion, verdict, reason); err != nil {
		fromReason := fmt.Sprintf("[ucg-audit-mq] chat moderation persist failed msgId=%d auditVersion=%d err=%v", msg.Id, auditVersion, err)
		markChatModerationFailed(ctx, msg.ConversationId, msg.Id, auditVersion, fromReason)
	}
}

func markChatModerationFailed(ctx context.Context, conversationID, messageID uint64, auditVersion int, fromReason string) {
	cols := dao.UcgChatMessage.Columns()
	_, _ = dao.UcgChatMessage.Ctx(ctx).
		Where(cols.ConversationId, conversationID).
		Where(cols.Id, messageID).
		Where(cols.AuditStatus, ChatAuditStatusPending).
		Where(cols.AuditVersion, auditVersion).
		Data(g.Map{
			cols.AuditStatus:   ChatAuditStatusModerationFailed,
			cols.RejectReason:  fromReason,
		}).Update()
}

func runChatApplyPhase(ctx context.Context, queueName string, msg entity.UcgChatMessage, auditVersion int) error {
	if msg.AuditStatus != ChatAuditStatusPending {
		return nil
	}
	if msg.ModerationVerdict == ModerationVerdictNone {
		glog.Errorf(ctx, "[ucg-audit-mq] chat apply without verdict queue=%s msgId=%d auditVersion=%d", queueName, msg.Id, auditVersion)
		return fmt.Errorf("chat apply: moderation verdict missing msgId=%d", msg.Id)
	}
	var applyErr error
	if msg.ModerationVerdict == ModerationVerdictReject {
		applyErr = rejectChatMessageCAS(ctx, msg.ConversationId, msg.Id, auditVersion, int64(msg.SenderWxId), msg.ClientMsgId, msg.ModerationReason)
	} else {
		applyErr = approveChatMessageCAS(ctx, msg.ConversationId, msg.Id, auditVersion)
	}
	if applyErr != nil {
		return handleChatApplyFailure(ctx, queueName, msg, auditVersion, applyErr)
	}
	return nil
}

func incrementChatApplyAttempts(ctx context.Context, conversationID, messageID uint64, auditVersion int) (int, error) {
	cols := dao.UcgChatMessage.Columns()
	_, err := dao.UcgChatMessage.Ctx(ctx).
		Where(cols.ConversationId, conversationID).
		Where(cols.Id, messageID).
		Where(cols.AuditVersion, auditVersion).
		Increment(cols.ApplyAttempts, 1)
	if err != nil {
		return 0, err
	}
	var msg entity.UcgChatMessage
	if scanErr := dao.UcgChatMessage.Ctx(ctx).
		Where(cols.ConversationId, conversationID).
		Where(cols.Id, messageID).
		Scan(&msg); scanErr != nil {
		return 0, scanErr
	}
	return msg.ApplyAttempts, nil
}

func markChatApplyFailed(ctx context.Context, conversationID, messageID uint64, auditVersion int) error {
	now := time.Now().Unix()
	cols := dao.UcgChatMessage.Columns()
	_, err := dao.UcgChatMessage.Ctx(ctx).
		Where(cols.ConversationId, conversationID).
		Where(cols.Id, messageID).
		Where(cols.AuditStatus, ChatAuditStatusPending).
		Where(cols.AuditVersion, auditVersion).
		Data(g.Map{
			cols.AuditStatus:   ChatAuditStatusRejected,
			cols.RejectReason:  applyFailedSystemReason,
			cols.ApplyFailedAt: now,
		}).Update()
	return err
}

func handleChatApplyFailure(ctx context.Context, queueName string, msg entity.UcgChatMessage, auditVersion int, applyErr error) error {
	attempts, incErr := incrementChatApplyAttempts(ctx, msg.ConversationId, msg.Id, auditVersion)
	if incErr != nil {
		return incErr
	}
	max := auditApplyMaxAttempts()
	if attempts >= max {
		if mErr := markChatApplyFailed(ctx, msg.ConversationId, msg.Id, auditVersion); mErr != nil {
			return mErr
		}
		glog.Errorf(ctx, "[ucg-audit-mq] chat apply max exceeded queue=%s msgId=%d auditVersion=%d attempts=%d err=%v",
			queueName, msg.Id, auditVersion, attempts, applyErr)
		return nil
	}
	glog.Warningf(ctx, "[ucg-audit-mq] chat apply retry queue=%s msgId=%d auditVersion=%d attempts=%d err=%v",
		queueName, msg.Id, auditVersion, attempts, applyErr)
	return applyErr
}

func loadChatMessageForAudit(ctx context.Context, conversationID, messageID uint64) (entity.UcgChatMessage, error) {
	var msg entity.UcgChatMessage
	err := dao.UcgChatMessage.Ctx(ctx).
		Where(dao.UcgChatMessage.Columns().ConversationId, conversationID).
		Where(dao.UcgChatMessage.Columns().Id, messageID).
		Scan(&msg)
	return msg, err
}

func loadProfileAuditJob(ctx context.Context, jobID uint64) (entity.UcgProfileAuditJob, error) {
	var job entity.UcgProfileAuditJob
	err := dao.UcgProfileAuditJob.Ctx(ctx).Where(dao.UcgProfileAuditJob.Columns().Id, jobID).Scan(&job)
	return job, err
}

func loadPostForAudit(ctx context.Context, postID uint64) (entity.UcgPost, error) {
	var post entity.UcgPost
	err := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).Scan(&post)
	return post, err
}
