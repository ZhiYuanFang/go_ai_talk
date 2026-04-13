// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SuggestDao is the data access object for table suggest.
type SuggestDao struct {
	table   string         // table is the underlying table name of the DAO.
	group   string         // group is the database configuration group name of current DAO.
	columns SuggestColumns // columns contains all the column names of Table for convenient usage.
}

// SuggestColumns defines and stores column names for table suggest.
type SuggestColumns struct {
	Id       string //
	DeviceNo string //
	Suggest  string //
	Time     string //
}

// suggestColumns holds the columns for table suggest.
var suggestColumns = SuggestColumns{
	Id:       "id",
	DeviceNo: "device_no",
	Suggest:  "suggest",
	Time:     "time",
}

// NewSuggestDao creates and returns a new DAO object for table data access.
func NewSuggestDao() *SuggestDao {
	return &SuggestDao{
		group:   "default",
		table:   "suggest",
		columns: suggestColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *SuggestDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *SuggestDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *SuggestDao) Columns() SuggestColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *SuggestDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *SuggestDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *SuggestDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
