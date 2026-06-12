// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UcgProfileAuditJobDao is the data access object for table ucg_profile_audit_job.
type UcgProfileAuditJobDao struct {
	table   string                      // table is the underlying table name of the DAO.
	group   string                      // group is the database configuration group name of current DAO.
	columns UcgProfileAuditJobColumns   // columns contains all the column names of Table for convenient usage.
}

// UcgProfileAuditJobColumns defines and stores column names for table ucg_profile_audit_job.
type UcgProfileAuditJobColumns struct {
	Id           string //
	WxId         string //
	Nickname     string //
	AvatarKey    string //
	Bio          string //
	Status       string // 1 pending 2 approved 3 rejected
	AuditVersion string // 审核轮次
	RejectReason string //
	CreatedAt    string //
	UpdatedAt    string //
}

var ucgProfileAuditJobColumns = UcgProfileAuditJobColumns{
	Id:           "id",
	WxId:         "wx_id",
	Nickname:     "nickname",
	AvatarKey:    "avatar_key",
	Bio:          "bio",
	Status:       "status",
	AuditVersion: "audit_version",
	RejectReason: "reject_reason",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewUcgProfileAuditJobDao creates and returns a new DAO object for table data access.
func NewUcgProfileAuditJobDao() *UcgProfileAuditJobDao {
	return &UcgProfileAuditJobDao{
		group:   "default",
		table:   "ucg_profile_audit_job",
		columns: ucgProfileAuditJobColumns,
	}
}

func (dao *UcgProfileAuditJobDao) DB() gdb.DB {
	return g.DB(dao.group)
}

func (dao *UcgProfileAuditJobDao) Table() string {
	return dao.table
}

func (dao *UcgProfileAuditJobDao) Columns() UcgProfileAuditJobColumns {
	return dao.columns
}

func (dao *UcgProfileAuditJobDao) Group() string {
	return dao.group
}

func (dao *UcgProfileAuditJobDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

func (dao *UcgProfileAuditJobDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
