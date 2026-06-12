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

// EnqueueProfileAuditJob 写入 MySQL 待审 job 并同事务入队 audit publish outbox。
func EnqueueProfileAuditJob(ctx context.Context, wxID int64, nickname, avatarKey, bio string) (jobID uint64, auditVersion int, outboxID uint64, err error) {
	if wxID <= 0 {
		return 0, 0, 0, fmt.Errorf("wxId 无效")
	}
	nickname = strings.TrimSpace(nickname)
	avatarKey = strings.TrimSpace(avatarKey)
	bio = strings.TrimSpace(bio)
	now := time.Now().Unix()

	err = dao.UcgProfileAuditJob.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var pending entity.UcgProfileAuditJob
		_ = tx.Model(dao.UcgProfileAuditJob.Table()).Ctx(ctx).
			Where(dao.UcgProfileAuditJob.Columns().WxId, wxID).
			Where(dao.UcgProfileAuditJob.Columns().Status, ProfileJobStatusPending).
			OrderDesc(dao.UcgProfileAuditJob.Columns().Id).
			Limit(1).
			Scan(&pending)

		auditVersion = 1
		if pending.Id > 0 {
			auditVersion = pending.AuditVersion + 1
			_, txErr := tx.Model(dao.UcgProfileAuditJob.Table()).Ctx(ctx).
				Where(dao.UcgProfileAuditJob.Columns().Id, pending.Id).
				Data(g.Map{
					dao.UcgProfileAuditJob.Columns().Nickname:          nickname,
					dao.UcgProfileAuditJob.Columns().AvatarKey:         avatarKey,
					dao.UcgProfileAuditJob.Columns().Bio:               bio,
					dao.UcgProfileAuditJob.Columns().AuditVersion:      auditVersion,
					dao.UcgProfileAuditJob.Columns().RejectReason:      "",
					dao.UcgProfileAuditJob.Columns().ModerationVerdict: ModerationVerdictNone,
					dao.UcgProfileAuditJob.Columns().ModerationReason:  "",
					dao.UcgProfileAuditJob.Columns().ModerationAt:      0,
					dao.UcgProfileAuditJob.Columns().ApplyAttempts:     0,
					dao.UcgProfileAuditJob.Columns().ApplyFailedAt:     0,
					dao.UcgProfileAuditJob.Columns().UpdatedAt:         now,
				}).Update()
			if txErr != nil {
				return txErr
			}
			jobID = pending.Id
		} else {
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
			id, _ := res.LastInsertId()
			jobID = uint64(id)
		}
		outboxID, err = enqueueAuditPublishOutboxTx(ctx, tx, eventkit.RoutingUcgProfilePatchSubmitted.String(),
			auditPublishProfilePayload(jobID, auditVersion))
		return err
	})
	return jobID, auditVersion, outboxID, err
}

// LoadLatestPendingProfileJob 读取用户最新 pending job（作者预览用）。
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

// auditProfileJobFromEvent MQ 消费者：两阶段 Green 机审 → apply CAS。
func auditProfileJobFromEvent(ctx context.Context, jobID uint64, auditVersion int) error {
	job, err := loadProfileAuditJob(ctx, jobID)
	if err != nil {
		return err
	}
	if job.Id == 0 {
		glog.Infof(ctx, "[ucg-audit-mq] profile job skip missing id=%d version=%d", jobID, auditVersion)
		return nil
	}
	if job.Status != ProfileJobStatusPending || job.AuditVersion != auditVersion {
		glog.Infof(ctx, "[ucg-audit-mq] profile job skip stale id=%d curStatus=%d curVersion=%d eventVersion=%d",
			jobID, job.Status, job.AuditVersion, auditVersion)
		return nil
	}

	if err = runProfileModerationPhase(ctx, ucgProfileQueue, job, auditVersion); err != nil {
		return err
	}
	job, err = loadProfileAuditJob(ctx, jobID)
	if err != nil {
		return err
	}
	return runProfileApplyPhase(ctx, ucgProfileQueue, job, auditVersion)
}

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
