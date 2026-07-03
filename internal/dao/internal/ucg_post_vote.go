// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UcgPostVoteDao is the data access object for table ucg_post_vote.
type UcgPostVoteDao struct {
	table   string               // table is the underlying table name of the DAO.
	group   string               // group is the database configuration group name of current DAO.
	columns UcgPostVoteColumns // columns contains all the column names of Table for convenient usage.
}

// UcgPostVoteColumns defines and stores column names for table ucg_post_vote.
type UcgPostVoteColumns struct {
	Id        string //
	PostId    string //
	VoterWxId string //
	Side      string //
	CreatedAt string //
	UpdatedAt string //
}

// ucgPostVoteColumns holds the columns for table ucg_post_vote.
var ucgPostVoteColumns = UcgPostVoteColumns{
	Id:        "id",
	PostId:    "post_id",
	VoterWxId: "voter_wx_id",
	Side:      "side",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewUcgPostVoteDao creates and returns a new DAO object for table data access.
func NewUcgPostVoteDao() *UcgPostVoteDao {
	return &UcgPostVoteDao{
		group:   "default",
		table:   "ucg_post_vote",
		columns: ucgPostVoteColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *UcgPostVoteDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *UcgPostVoteDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *UcgPostVoteDao) Columns() UcgPostVoteColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *UcgPostVoteDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *UcgPostVoteDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *UcgPostVoteDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
