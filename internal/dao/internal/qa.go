// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// QaDao is the data access object for table qa.
type QaDao struct {
	table   string    // table is the underlying table name of the DAO.
	group   string    // group is the database configuration group name of current DAO.
	columns QaColumns // columns contains all the column names of Table for convenient usage.
}

// QaColumns defines and stores column names for table qa.
type QaColumns struct {
	Id              string //
	Question        string // 问题
	IntentionId     string // 意图id
	IntentionAnswer string // 意图下的回答
	Replay          string // 回复
	Attack          string // 命中次数
}

// qaColumns holds the columns for table qa.
var qaColumns = QaColumns{
	Id:              "id",
	Question:        "question",
	IntentionId:     "intention_id",
	IntentionAnswer: "intention_answer",
	Replay:          "replay",
	Attack:          "attack",
}

// NewQaDao creates and returns a new DAO object for table data access.
func NewQaDao() *QaDao {
	return &QaDao{
		group:   "default",
		table:   "qa",
		columns: qaColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *QaDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *QaDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *QaDao) Columns() QaColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *QaDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *QaDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *QaDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
