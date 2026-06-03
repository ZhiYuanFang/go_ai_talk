// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UcgPostMediaDao is the data access object for table ucg_post_media.
type UcgPostMediaDao struct {
	table   string              // table is the underlying table name of the DAO.
	group   string              // group is the database configuration group name of current DAO.
	columns UcgPostMediaColumns // columns contains all the column names of Table for convenient usage.
}

// UcgPostMediaColumns defines and stores column names for table ucg_post_media.
type UcgPostMediaColumns struct {
	Id         string //
	PostId     string //
	SortOrder  string //
	ObjectKey  string //
	MediaKind  string // 1 image 2 video
	DurationMs string //
	SizeBytes  string //
}

// ucgPostMediaColumns holds the columns for table ucg_post_media.
var ucgPostMediaColumns = UcgPostMediaColumns{
	Id:         "id",
	PostId:     "post_id",
	SortOrder:  "sort_order",
	ObjectKey:  "object_key",
	MediaKind:  "media_kind",
	DurationMs: "duration_ms",
	SizeBytes:  "size_bytes",
}

// NewUcgPostMediaDao creates and returns a new DAO object for table data access.
func NewUcgPostMediaDao() *UcgPostMediaDao {
	return &UcgPostMediaDao{
		group:   "default",
		table:   "ucg_post_media",
		columns: ucgPostMediaColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *UcgPostMediaDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *UcgPostMediaDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *UcgPostMediaDao) Columns() UcgPostMediaColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *UcgPostMediaDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *UcgPostMediaDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *UcgPostMediaDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
