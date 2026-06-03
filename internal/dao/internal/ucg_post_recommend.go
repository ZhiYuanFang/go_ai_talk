// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UcgPostRecommendDao is the data access object for table ucg_post_recommend.
type UcgPostRecommendDao struct {
	table   string                  // table is the underlying table name of the DAO.
	group   string                  // group is the database configuration group name of current DAO.
	columns UcgPostRecommendColumns // columns contains all the column names of Table for convenient usage.
}

// UcgPostRecommendColumns defines and stores column names for table ucg_post_recommend.
type UcgPostRecommendColumns struct {
	PostId     string //
	Score      string //
	ComputedAt string //
}

// ucgPostRecommendColumns holds the columns for table ucg_post_recommend.
var ucgPostRecommendColumns = UcgPostRecommendColumns{
	PostId:     "post_id",
	Score:      "score",
	ComputedAt: "computed_at",
}

// NewUcgPostRecommendDao creates and returns a new DAO object for table data access.
func NewUcgPostRecommendDao() *UcgPostRecommendDao {
	return &UcgPostRecommendDao{
		group:   "default",
		table:   "ucg_post_recommend",
		columns: ucgPostRecommendColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *UcgPostRecommendDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *UcgPostRecommendDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *UcgPostRecommendDao) Columns() UcgPostRecommendColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *UcgPostRecommendDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *UcgPostRecommendDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *UcgPostRecommendDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
