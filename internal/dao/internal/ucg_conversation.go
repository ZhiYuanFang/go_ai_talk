// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UcgConversationDao is the data access object for table ucg_conversation.
type UcgConversationDao struct {
	table   string                 // table is the underlying table name of the DAO.
	group   string                 // group is the database configuration group name of current DAO.
	columns UcgConversationColumns // columns contains all the column names of Table for convenient usage.
}

// UcgConversationColumns defines and stores column names for table ucg_conversation.
type UcgConversationColumns struct {
	Id        string //
	Type      string // 1 direct
	CreatedAt string //
	UpdatedAt string //
}

// ucgConversationColumns holds the columns for table ucg_conversation.
var ucgConversationColumns = UcgConversationColumns{
	Id:        "id",
	Type:      "type",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewUcgConversationDao creates and returns a new DAO object for table data access.
func NewUcgConversationDao() *UcgConversationDao {
	return &UcgConversationDao{
		group:   "default",
		table:   "ucg_conversation",
		columns: ucgConversationColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *UcgConversationDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *UcgConversationDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *UcgConversationDao) Columns() UcgConversationColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *UcgConversationDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *UcgConversationDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *UcgConversationDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
