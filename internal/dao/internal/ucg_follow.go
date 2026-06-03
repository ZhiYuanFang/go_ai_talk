// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UcgFollowDao is the data access object for table ucg_follow.
type UcgFollowDao struct {
	table   string           // table is the underlying table name of the DAO.
	group   string           // group is the database configuration group name of current DAO.
	columns UcgFollowColumns // columns contains all the column names of Table for convenient usage.
}

// UcgFollowColumns defines and stores column names for table ucg_follow.
type UcgFollowColumns struct {
	Id           string //
	FollowerWxId string //
	FolloweeWxId string //
	CreatedAt    string //
}

// ucgFollowColumns holds the columns for table ucg_follow.
var ucgFollowColumns = UcgFollowColumns{
	Id:           "id",
	FollowerWxId: "follower_wx_id",
	FolloweeWxId: "followee_wx_id",
	CreatedAt:    "created_at",
}

// NewUcgFollowDao creates and returns a new DAO object for table data access.
func NewUcgFollowDao() *UcgFollowDao {
	return &UcgFollowDao{
		group:   "default",
		table:   "ucg_follow",
		columns: ucgFollowColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *UcgFollowDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *UcgFollowDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *UcgFollowDao) Columns() UcgFollowColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *UcgFollowDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *UcgFollowDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *UcgFollowDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
