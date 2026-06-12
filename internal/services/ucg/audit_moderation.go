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

const ucgAuditApplyMaxAttemptsEnv = "UCG_AUDIT_APPLY_MAX_ATTEMPTS"

const defaultAuditApplyMaxAttempts = 5

// auditApplyMaxAttempts 读取 apply 阶段最大重试次数；超限后 Ack 停止 MQ 风暴。
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

// persistModerationVerdictProfile Phase1：CAS 写入机审结论；moderation_verdict=0 条件避免并发双调 Green。
func persistModerationVerdictProfile(ctx context.Context, jobID uint64, auditVersion, verdict int, reason string) error {
	now := time.Now().Unix()
	cols := dao.UcgProfileAuditJob.Columns()
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
		return err
	}
	affected, _ := result.RowsAffected()
	if affected > 0 {
		return nil
	}
	// 并发 consumer 已写入 verdict：读回确认，不再调 Green。
	var job entity.UcgProfileAuditJob
	if scanErr := dao.UcgProfileAuditJob.Ctx(ctx).Where(cols.Id, jobID).Scan(&job); scanErr != nil {
		return scanErr
	}
	if job.ModerationVerdict == ModerationVerdictNone {
		return fmt.Errorf("profile moderation verdict cas lost id=%d version=%d", jobID, auditVersion)
	}
	return nil
}

// persistModerationVerdictPost Phase1：帖子机审结论 CAS 落库。
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

func runProfileGreenChecks(ctx context.Context, job entity.UcgProfileAuditJob) (pass bool, reason string, err error) {
	moderator := EffectiveGreen()
	cfg := LoadOSSConfig(ctx)
	if job.Nickname != "" {
		var verdict AuditVerdict
		verdict, err = moderator.ModerateText(ctx, "nickname_detection", job.Nickname)
		if err != nil {
			return false, "", err
		}
		if !verdict.Pass {
			return false, verdict.Reason, nil
		}
	}
	if job.Bio != "" {
		var verdict AuditVerdict
		verdict, err = moderator.ModerateText(ctx, "comment_detection", job.Bio)
		if err != nil {
			return false, "", err
		}
		if !verdict.Pass {
			return false, verdict.Reason, nil
		}
	}
	if job.AvatarKey != "" {
		url := cfg.CdnBaseURL + "/" + job.AvatarKey
		var verdict AuditVerdict
		verdict, err = moderator.ModerateImageURL(ctx, url)
		if err != nil {
			return false, "", err
		}
		if !verdict.Pass {
			return false, verdict.Reason, nil
		}
	}
	return true, "", nil
}

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

// runProfileModerationPhase Phase1：Green 机审并持久化 verdict；verdict 已存在则跳过 Green（MQ 重投幂等）。
func runProfileModerationPhase(ctx context.Context, queueName string, job entity.UcgProfileAuditJob, auditVersion int) error {
	if job.ModerationVerdict != ModerationVerdictNone {
		glog.Infof(ctx, "[ucg-audit-mq] profile moderation skip green queue=%s jobId=%d wxId=%d auditVersion=%d verdict=%d apply_attempts=%d",
			queueName, job.Id, job.WxId, auditVersion, job.ModerationVerdict, job.ApplyAttempts)
		return nil
	}
	pass, reason, err := runProfileGreenChecks(ctx, job)
	if err != nil {
		return err
	}
	verdict := ModerationVerdictPass
	if !pass {
		verdict = ModerationVerdictReject
		if reason == "" {
			reason = rejectReasonDefault
		}
	}
	return persistModerationVerdictProfile(ctx, job.Id, auditVersion, verdict, reason)
}

// runProfileApplyPhase Phase2：基于已持久化 verdict 执行 CAS apply；失败有界重试。
func runProfileApplyPhase(ctx context.Context, queueName string, job entity.UcgProfileAuditJob, auditVersion int) error {
	if job.Status != ProfileJobStatusPending {
		return nil
	}
	if job.ModerationVerdict == ModerationVerdictNone {
		glog.Errorf(ctx, "[ucg-audit-mq] profile apply without verdict queue=%s jobId=%d auditVersion=%d", queueName, job.Id, auditVersion)
		return fmt.Errorf("profile apply: moderation verdict missing jobId=%d", job.Id)
	}
	var applyErr error
	if job.ModerationVerdict == ModerationVerdictReject {
		applyErr = rejectProfileJobCAS(ctx, job.Id, auditVersion, job.ModerationReason)
	} else {
		applyErr = approveProfileJobCAS(ctx, job, auditVersion)
	}
	if applyErr != nil {
		return handleProfileApplyFailure(ctx, queueName, job, auditVersion, applyErr)
	}
	return nil
}

// runPostModerationPhase Phase1：帖子 Green 机审；verdict 已落库则跳过重投时的 Green 调用。
func runPostModerationPhase(ctx context.Context, queueName string, post entity.UcgPost, auditVersion int) error {
	if post.ModerationVerdict != ModerationVerdictNone {
		glog.Infof(ctx, "[ucg-audit-mq] post moderation skip green queue=%s postId=%d authorWxId=%d auditVersion=%d verdict=%d apply_attempts=%d",
			queueName, post.Id, post.AuthorWxId, auditVersion, post.ModerationVerdict, post.ApplyAttempts)
		return nil
	}
	pass, reason, err := runPostGreenChecks(ctx, post)
	if err != nil {
		return err
	}
	verdict := ModerationVerdictPass
	if !pass {
		verdict = ModerationVerdictReject
		if reason == "" {
			reason = rejectReasonDefault
		}
	}
	return persistModerationVerdictPost(ctx, post.Id, auditVersion, verdict, reason)
}

// runPostApplyPhase Phase2：帖子 apply（publish/reject CAS）；失败有界重试。
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
			cols.Status:       ProfileJobStatusApplyFailed,
			cols.RejectReason: applyFailedSystemReason,
			cols.ApplyFailedAt: now,
			cols.UpdatedAt:    now,
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
			cols.Status:       PostStatusApplyFailed,
			cols.RejectReason: applyFailedSystemReason,
			cols.ApplyFailedAt: now,
			cols.UpdatedAt:    now,
		}).Update()
	return err
}

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
		return nil
	}
	glog.Warningf(ctx, "[ucg-audit-mq] profile apply retry queue=%s jobId=%d wxId=%d auditVersion=%d apply_attempts=%d err=%v",
		queueName, job.Id, job.WxId, auditVersion, attempts, applyErr)
	return applyErr
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
