package ucg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/platform/eventkit"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// EnqueueProfileAuditJob HTTP 侧：写入/更新 ucg_profile_audit_job 并同事务写 audit outbox。
// 注意：再次提审会把 moderation_verdict 重置为 0，MQ 消费时会重新走 Phase1 Green。
// 返回值语义：jobID=0 无 job；auditVersion=0 无 patch；outboxID=0 无 outbox。
func EnqueueProfileAuditJob(ctx context.Context, wxID int64, nickname, avatarKey, bio string) (jobID uint64, auditVersion int, outboxID uint64, err error) {
	if wxID <= 0 {
		return 0, 0, 0, fmt.Errorf("wxId 无效")
	}
	nickname = strings.TrimSpace(nickname)
	avatarKey = strings.TrimSpace(avatarKey)
	bio = strings.TrimSpace(bio)
	now := time.Now().Unix()

	// 事务写 audit job 与 outbox
	err = dao.UcgProfileAuditJob.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 获取当前用户最新 pending job 用于生成 audit version 与重置 moderation_verdict 为 0
		var pending entity.UcgProfileAuditJob
		_ = tx.Model(dao.UcgProfileAuditJob.Table()).Ctx(ctx).
			Where(dao.UcgProfileAuditJob.Columns().WxId, wxID).
			Where(dao.UcgProfileAuditJob.Columns().Status, ProfileJobStatusPending).
			OrderDesc(dao.UcgProfileAuditJob.Columns().Id).
			Limit(1).
			Scan(&pending)

		// 生成 audit version
		auditVersion = 1
		// 如果 pending job 存在，则生成新的 audit version
		if pending.Id > 0 {
			// 如果 pending job 存在，则生成新的 audit version
			auditVersion = pending.AuditVersion + 1
			// 更新 pending job
			_, txErr := tx.Model(dao.UcgProfileAuditJob.Table()).Ctx(ctx).
				Where(dao.UcgProfileAuditJob.Columns().Id, pending.Id).
				Data(g.Map{
					dao.UcgProfileAuditJob.Columns().Nickname:          nickname,
					dao.UcgProfileAuditJob.Columns().AvatarKey:         avatarKey,
					dao.UcgProfileAuditJob.Columns().Bio:               bio,
					dao.UcgProfileAuditJob.Columns().AuditVersion:      auditVersion,
					dao.UcgProfileAuditJob.Columns().RejectReason:      "",
					dao.UcgProfileAuditJob.Columns().ModerationVerdict: ModerationVerdictNone, // 重置 → 下次消费必调 Green
					dao.UcgProfileAuditJob.Columns().ModerationReason:  "",
					dao.UcgProfileAuditJob.Columns().ModerationAt:      0,
					dao.UcgProfileAuditJob.Columns().ApplyAttempts:     0,
					dao.UcgProfileAuditJob.Columns().ApplyFailedAt:     0,
					dao.UcgProfileAuditJob.Columns().UpdatedAt:         now,
				}).Update()
			if txErr != nil {
				return txErr
			}
			// 设置 jobID 为 pending job 的 id 用于后续事务写 outbox
			jobID = pending.Id
		} else {
			// 如果 pending job 不存在，则创建新的 pending job
			res, txErr := tx.Model(dao.UcgProfileAuditJob.Table()).Ctx(ctx).Data(g.Map{
				dao.UcgProfileAuditJob.Columns().WxId:         wxID,
				dao.UcgProfileAuditJob.Columns().Nickname:     nickname,
				dao.UcgProfileAuditJob.Columns().AvatarKey:    avatarKey,
				dao.UcgProfileAuditJob.Columns().Bio:          bio,
				dao.UcgProfileAuditJob.Columns().Status:       ProfileJobStatusPending,
				dao.UcgProfileAuditJob.Columns().AuditVersion: auditVersion,
				dao.UcgProfileAuditJob.Columns().CreatedAt:    now,
				dao.UcgProfileAuditJob.Columns().UpdatedAt:    now,
			}).Insert()
			if txErr != nil {
				return txErr
			}
			// 设置 jobID 为新创建的 pending job 的 id 用于后续事务写 outbox
			id, _ := res.LastInsertId()
			jobID = uint64(id)
		}
		// outbox relay 会 publish 到 ucg.profile.patch.submitted → 队列 ucg.profile.patch.submitted.q
		outboxID, err = enqueueAuditPublishOutboxTx(ctx, tx, eventkit.RoutingUcgProfilePatchSubmitted.String(),
			auditPublishProfilePayload(jobID, auditVersion))
		return err
	})
	return jobID, auditVersion, outboxID, err
}

// LoadLatestPendingProfileJob 作者预览：读取 pending job 合并到 ProfileDTO。
func LoadLatestPendingProfileJob(ctx context.Context, wxID int64) (entity.UcgProfileAuditJob, bool, error) {
	var job entity.UcgProfileAuditJob
	err := dao.UcgProfileAuditJob.Ctx(ctx).
		Where(dao.UcgProfileAuditJob.Columns().WxId, wxID).
		Where(dao.UcgProfileAuditJob.Columns().Status, ProfileJobStatusPending).
		OrderDesc(dao.UcgProfileAuditJob.Columns().Id).
		Limit(1).
		Scan(&job)
	if err != nil {
		return job, false, err
	}
	if job.Id == 0 {
		return job, false, nil
	}
	return job, true, nil
}

// auditProfileJobFromEvent MQ 资料审核入口：Phase1 Green → Phase2 apply。
// 返回值语义：nil=Ack 删消息；非 nil=eventkit Nack(requeue=true) 消息回队。
func auditProfileJobFromEvent(ctx context.Context, jobID uint64, auditVersion int) error {
	job, err := loadProfileAuditJob(ctx, jobID)
	if err != nil {
		return err // DB 读失败 → requeue
	}
	if job.Id == 0 {
		glog.Infof(ctx, "[ucg-audit-mq] profile job skip missing id=%d version=%d", jobID, auditVersion)
		return nil // 无 job → Ack，避免毒消息
	}
	if job.Status != ProfileJobStatusPending || job.AuditVersion != auditVersion {
		glog.Infof(ctx, "[ucg-audit-mq] profile job skip stale id=%d curStatus=%d curVersion=%d eventVersion=%d",
			jobID, job.Status, job.AuditVersion, auditVersion)
		return nil // 过期消息（版本已变）→ Ack
	}

	// Phase1：可能调 Green；失败返回 err → 【路径 A 风暴】
	runProfileModerationPhase(ctx, ucgProfileQueue, job, auditVersion)

	job, err = loadProfileAuditJob(ctx, jobID)
	if err != nil || job.Id == 0 {
		return nil // 防脏数据
	}

	if err := runProfileApplyPhase(ctx, ucgProfileQueue, job, auditVersion); err != nil {
		glog.Errorf(ctx, "profile apply failed, skip retry jobId=%d err=%v", job.Id, err)
		// TODO：可在此记录补偿标记
	}
	return nil
}

// approveProfileJobCAS Phase2 通过：job pending→approved，并按 job 非空字段更新 ucg_profile。
// 此处失败且 verdict 已=1 → 路径 B（只 retry apply，不 retry Green）。
func approveProfileJobCAS(ctx context.Context, job entity.UcgProfileAuditJob, auditVersion int) error {
	now := time.Now().Unix()
	return dao.UcgProfileAuditJob.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		affected, err := CasAuditTransitionTx(ctx, tx, CasAuditInput{
			Table:       dao.UcgProfileAuditJob.Table(),
			ID:          job.Id,
			Kind:        AuditCasKindStatus,
			FromStatus:  ProfileJobStatusPending,
			ToStatus:    ProfileJobStatusApproved,
			FromVersion: auditVersion,
			Extra: g.Map{
				dao.UcgProfileAuditJob.Columns().RejectReason: "",
				dao.UcgProfileAuditJob.Columns().UpdatedAt:    now,
			},
		})
		if err != nil {
			return err
		}
		if affected == 0 {
			glog.Infof(ctx, "[ucg-audit-mq] profile job cas skip id=%d version=%d", job.Id, auditVersion)
			return nil
		}
		data := g.Map{dao.UcgProfile.Columns().UpdatedAt: now}
		if job.Nickname != "" {
			data[dao.UcgProfile.Columns().Nickname] = job.Nickname
		}
		if job.AvatarKey != "" {
			data[dao.UcgProfile.Columns().AvatarKey] = job.AvatarKey
		}
		if job.Bio != "" {
			data[dao.UcgProfile.Columns().Bio] = job.Bio
		}
		_, err = tx.Model(dao.UcgProfile.Table()).Ctx(ctx).
			Where(dao.UcgProfile.Columns().WxId, job.WxId).
			Data(data).Update()
		return err
	})
}

func rejectProfileJobCAS(ctx context.Context, jobID uint64, auditVersion int, reason string) error {
	if reason == "" {
		reason = rejectReasonDefault
	}
	now := time.Now().Unix()
	affected, err := CasAuditTransition(ctx, CasAuditInput{
		Table:       dao.UcgProfileAuditJob.Table(),
		ID:          jobID,
		Kind:        AuditCasKindStatus,
		FromStatus:  ProfileJobStatusPending,
		ToStatus:    ProfileJobStatusRejected,
		FromVersion: auditVersion,
		Extra: g.Map{
			dao.UcgProfileAuditJob.Columns().RejectReason: reason,
			dao.UcgProfileAuditJob.Columns().UpdatedAt:    now,
		},
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		glog.Infof(ctx, "[ucg-audit-mq] profile job reject cas skip id=%d version=%d", jobID, auditVersion)
	}
	return nil
}
