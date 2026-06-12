package ucg

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AuditCasKind 审态 CAS 条件列类型：整型 status 或字符串 audit_status。
type AuditCasKind int

const (
	AuditCasKindStatus      AuditCasKind = iota // status + audit_version（帖/评/job）
	AuditCasKindAuditStatus                     // audit_status + audit_version（私信）
)

// CasAuditInput 条件更新入参；SET 子句禁止递增 audit_version。
type CasAuditInput struct {
	Table       string
	IDColumn    string
	ID          uint64
	Kind        AuditCasKind
	FromStatus  any // int 或 string
	ToStatus    any
	FromVersion int
	Extra       g.Map // 额外 SET 字段（reject_reason、published_at 等）
	ExtraWhere  g.Map // 额外 WHERE 条件（如 conversation_id）
}

// CasAuditTransition 执行 audit_version CAS 更新；返回影响行数（0 表示过期/重复消息）。
func CasAuditTransition(ctx context.Context, in CasAuditInput) (int64, error) {
	if in.Table == "" || in.ID == 0 {
		return 0, fmt.Errorf("cas: 无效表或 id")
	}
	idCol := in.IDColumn
	if idCol == "" {
		idCol = "id"
	}
	versionCol := "audit_version"
	data := g.Map{}
	switch in.Kind {
	case AuditCasKindAuditStatus:
		data["audit_status"] = in.ToStatus
	default:
		data["status"] = in.ToStatus
	}
	for k, v := range in.Extra {
		data[k] = v
	}
	model := g.DB().Model(in.Table).Ctx(ctx).Where(idCol, in.ID)
	for k, v := range in.ExtraWhere {
		model = model.Where(k, v)
	}
	switch in.Kind {
	case AuditCasKindAuditStatus:
		model = model.Where("audit_status", in.FromStatus)
	default:
		model = model.Where("status", in.FromStatus)
	}
	model = model.Where(versionCol, in.FromVersion)
	result, err := model.Data(data).Update()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CasAuditTransitionInStatuses 管理端驳回等场景：from 为 status 集合（如 pending+published）。
func CasAuditTransitionInStatuses(ctx context.Context, table string, id uint64, fromStatuses []int, fromVersion int, toStatus int, extra g.Map) (int64, error) {
	if table == "" || id == 0 || len(fromStatuses) == 0 {
		return 0, fmt.Errorf("cas: 无效参数")
	}
	data := g.Map{"status": toStatus}
	for k, v := range extra {
		data[k] = v
	}
	result, err := g.DB().Model(table).Ctx(ctx).
		Where("id", id).
		WhereIn("status", fromStatuses).
		Where("audit_version", fromVersion).
		Data(data).Update()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CasAuditTransitionTx 事务内 CAS（资料 job apply 等同事务路径）。
func CasAuditTransitionTx(ctx context.Context, tx gdb.TX, in CasAuditInput) (int64, error) {
	if in.Table == "" || in.ID == 0 {
		return 0, fmt.Errorf("cas: 无效表或 id")
	}
	idCol := in.IDColumn
	if idCol == "" {
		idCol = "id"
	}
	data := g.Map{}
	switch in.Kind {
	case AuditCasKindAuditStatus:
		data["audit_status"] = in.ToStatus
	default:
		data["status"] = in.ToStatus
	}
	for k, v := range in.Extra {
		data[k] = v
	}
	model := tx.Model(in.Table).Ctx(ctx).Where(idCol, in.ID)
	switch in.Kind {
	case AuditCasKindAuditStatus:
		model = model.Where("audit_status", in.FromStatus)
	default:
		model = model.Where("status", in.FromStatus)
	}
	model = model.Where("audit_version", in.FromVersion)
	result, err := model.Data(data).Update()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
