// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// EventDao is the data access object for table event.
type EventDao struct {
	table   string       // table is the underlying table name of the DAO.
	group   string       // group is the database configuration group name of current DAO.
	columns EventColumns // columns contains all the column names of Table for convenient usage.
}

// EventColumns defines and stores column names for table event.
type EventColumns struct {
	Id           string //
	Name         string // 吃奶/睡觉/尿/屎等
	NeedQuantity string // 是否需要计数1要
	ExtraNames   string // name的其它表达方式
	Color        string //
	Logo         string //
}

// eventColumns holds the columns for table event.
var eventColumns = EventColumns{
	Id:           "id",
	Name:         "name",
	NeedQuantity: "need_quantity",
	ExtraNames:   "extra_names",
	Color:        "color",
	Logo:         "logo",
}

// NewEventDao creates and returns a new DAO object for table data access.
func NewEventDao() *EventDao {
	return &EventDao{
		group:   "default",
		table:   "event",
		columns: eventColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *EventDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *EventDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *EventDao) Columns() EventColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *EventDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *EventDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *EventDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
