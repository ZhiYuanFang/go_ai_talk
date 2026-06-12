// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UcgPostCommentDao is the data access object for table ucg_post_comment.
type UcgPostCommentDao struct {
	table   string                // table is the underlying table name of the DAO.
	group   string                // group is the database configuration group name of current DAO.
	columns UcgPostCommentColumns // columns contains all the column names of Table for convenient usage.
}

// UcgPostCommentColumns defines and stores column names for table ucg_post_comment.
type UcgPostCommentColumns struct {
	Id         string //
	PostId     string //
	AuthorWxId string //
	Content      string //
	Status       string // 0 draft 1 pending_audit 2 published 3 rejected
	AuditVersion string // 审核轮次
	RejectReason string //
	CreatedAt    string //
}

// ucgPostCommentColumns holds the columns for table ucg_post_comment.
var ucgPostCommentColumns = UcgPostCommentColumns{
	Id:         "id",
	PostId:     "post_id",
	AuthorWxId: "author_wx_id",
	Content:      "content",
	Status:       "status",
	AuditVersion: "audit_version",
	RejectReason: "reject_reason",
	CreatedAt:    "created_at",
}

// NewUcgPostCommentDao creates and returns a new DAO object for table data access.
func NewUcgPostCommentDao() *UcgPostCommentDao {
	return &UcgPostCommentDao{
		group:   "default",
		table:   "ucg_post_comment",
		columns: ucgPostCommentColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *UcgPostCommentDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *UcgPostCommentDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *UcgPostCommentDao) Columns() UcgPostCommentColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *UcgPostCommentDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *UcgPostCommentDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *UcgPostCommentDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
