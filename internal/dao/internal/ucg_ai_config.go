// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UcgAiConfigDao is the data access object for table ucg_ai_config.
type UcgAiConfigDao struct {
	table   string             // table is the underlying table name of the DAO.
	group   string             // group is the database configuration group name of current DAO.
	columns UcgAiConfigColumns // columns contains all the column names of Table for convenient usage.
}

// UcgAiConfigColumns defines and stores column names for table ucg_ai_config.
type UcgAiConfigColumns struct {
	Id                  string // singleton row id=1
	VisionModel         string //
	MaxImagesPerRequest string //
	UpdatedAt           string // unix seconds
	UpdatedBy           string // admin operator label
}

// ucgAiConfigColumns holds the columns for table ucg_ai_config.
var ucgAiConfigColumns = UcgAiConfigColumns{
	Id:                  "id",
	VisionModel:         "vision_model",
	MaxImagesPerRequest: "max_images_per_request",
	UpdatedAt:           "updated_at",
	UpdatedBy:           "updated_by",
}

// NewUcgAiConfigDao creates and returns a new DAO object for table data access.
func NewUcgAiConfigDao() *UcgAiConfigDao {
	return &UcgAiConfigDao{
		group:   "default",
		table:   "ucg_ai_config",
		columns: ucgAiConfigColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *UcgAiConfigDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *UcgAiConfigDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *UcgAiConfigDao) Columns() UcgAiConfigColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *UcgAiConfigDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *UcgAiConfigDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *UcgAiConfigDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
