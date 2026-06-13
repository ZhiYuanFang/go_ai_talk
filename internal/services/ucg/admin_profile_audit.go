package ucg

import (
	"context"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// ProfileAuditJobAdminItem 管理端资料机审失败 job 列表项。
type ProfileAuditJobAdminItem struct {
	JobId        uint64 `json:"jobId"`
	WxId         int64  `json:"wxId"`
	AuditVersion int    `json:"auditVersion"`
	Status       int    `json:"status"`
	Nickname     string `json:"nickname,omitempty"`
	AvatarKey    string `json:"avatarKey,omitempty"`
	Bio          string `json:"bio,omitempty"`
	RejectReason string `json:"rejectReason,omitempty"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
}

// ListProfileAuditJobsForAdmin 分页列出 profile audit job；默认 status=moderation_failed。
func ListProfileAuditJobsForAdmin(ctx context.Context, page, pageSize int, statusFilter *int) (*PageResult, error) {
	p := NormalizeAdminPage(page, pageSize)
	status := ProfileJobStatusModerationFailed
	if statusFilter != nil {
		status = *statusFilter
	}
	cols := dao.UcgProfileAuditJob.Columns()
	model := dao.UcgProfileAuditJob.Ctx(ctx).Where(cols.Status, status)
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	var jobs []entity.UcgProfileAuditJob
	err = model.OrderDesc(cols.Id).Limit(p.PageSize).Offset(pageOffset(p)).Scan(&jobs)
	if err != nil {
		return nil, err
	}
	items := make([]ProfileAuditJobAdminItem, 0, len(jobs))
	for _, job := range jobs {
		items = append(items, ProfileAuditJobAdminItem{
			JobId:        job.Id,
			WxId:         int64(job.WxId),
			AuditVersion: job.AuditVersion,
			Status:       job.Status,
			Nickname:     job.Nickname,
			AvatarKey:    job.AvatarKey,
			Bio:          job.Bio,
			RejectReason: job.RejectReason,
			CreatedAt:    job.CreatedAt,
			UpdatedAt:    job.UpdatedAt,
		})
	}
	return &PageResult{List: items, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// ResolveProfileAuditJobAdmin 人工通过/驳回 moderation_failed job。
func ResolveProfileAuditJobAdmin(ctx context.Context, jobID uint64, action, reason string) error {
	if jobID == 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "jobId 无效")
	}
	action = strings.TrimSpace(strings.ToLower(action))
	reason = strings.TrimSpace(reason)
	switch action {
	case "approve":
		return approveProfileJobFromModerationFailedAdmin(ctx, jobID)
	case "reject":
		if reason == "" {
			return gerror.NewCode(gcode.CodeInvalidParameter, "驳回须填写 reason")
		}
		return rejectProfileJobFromModerationFailedAdmin(ctx, jobID, reason)
	default:
		return gerror.NewCode(gcode.CodeInvalidParameter, "action 须为 approve 或 reject")
	}
}

func approveProfileJobFromModerationFailedAdmin(ctx context.Context, jobID uint64) error {
	job, err := loadProfileAuditJob(ctx, jobID)
	if err != nil {
		return err
	}
	if job.Id == 0 || job.Status != ProfileJobStatusModerationFailed {
		return gerror.NewCode(gcode.CodeInvalidParameter, "job 非机审失败态或不存在")
	}
	now := time.Now().Unix()
	return dao.UcgProfileAuditJob.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		cols := dao.UcgProfileAuditJob.Columns()
		affected, txErr := CasAuditTransitionTx(ctx, tx, CasAuditInput{
			Table:       dao.UcgProfileAuditJob.Table(),
			ID:          job.Id,
			Kind:        AuditCasKindStatus,
			FromStatus:  ProfileJobStatusModerationFailed,
			ToStatus:    ProfileJobStatusApproved,
			FromVersion: job.AuditVersion,
			Extra: g.Map{
				cols.RejectReason:      "",
				cols.ModerationVerdict: ModerationVerdictPass,
				cols.ModerationAt:      now,
				cols.UpdatedAt:         now,
			},
		})
		if txErr != nil {
			return txErr
		}
		if affected == 0 {
			return nil // 幂等：已被处理
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
		_, txErr = tx.Model(dao.UcgProfile.Table()).Ctx(ctx).
			Where(dao.UcgProfile.Columns().WxId, job.WxId).
			Data(data).Update()
		return txErr
	})
}

func rejectProfileJobFromModerationFailedAdmin(ctx context.Context, jobID uint64, reason string) error {
	job, err := loadProfileAuditJob(ctx, jobID)
	if err != nil {
		return err
	}
	if job.Id == 0 || job.Status != ProfileJobStatusModerationFailed {
		return gerror.NewCode(gcode.CodeInvalidParameter, "job 非机审失败态或不存在")
	}
	now := time.Now().Unix()
	affected, err := CasAuditTransition(ctx, CasAuditInput{
		Table:       dao.UcgProfileAuditJob.Table(),
		ID:          job.Id,
		Kind:        AuditCasKindStatus,
		FromStatus:  ProfileJobStatusModerationFailed,
		ToStatus:    ProfileJobStatusRejected,
		FromVersion: job.AuditVersion,
		Extra: g.Map{
			dao.UcgProfileAuditJob.Columns().RejectReason: reason,
			dao.UcgProfileAuditJob.Columns().UpdatedAt:    now,
		},
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return nil
	}
	return nil
}
