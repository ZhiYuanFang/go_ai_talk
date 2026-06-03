// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UcgConversationMemberDao is the data access object for table ucg_conversation_member.
type UcgConversationMemberDao struct {
	table   string                       // table is the underlying table name of the DAO.
	group   string                       // group is the database configuration group name of current DAO.
	columns UcgConversationMemberColumns // columns contains all the column names of Table for convenient usage.
}

// UcgConversationMemberColumns defines and stores column names for table ucg_conversation_member.
type UcgConversationMemberColumns struct {
	Id             string //
	ConversationId string //
	WxId           string //
	Pinned         string //
	DeletedAt      string // user soft delete
	LastReadMsgId  string //
	UnreadCount    string //
	UpdatedAt      string // last activity; drives idx_wx_list sort
}

// ucgConversationMemberColumns holds the columns for table ucg_conversation_member.
var ucgConversationMemberColumns = UcgConversationMemberColumns{
	Id:             "id",
	ConversationId: "conversation_id",
	WxId:           "wx_id",
	Pinned:         "pinned",
	DeletedAt:      "deleted_at",
	LastReadMsgId:  "last_read_msg_id",
	UnreadCount:    "unread_count",
	UpdatedAt:      "updated_at",
}

// NewUcgConversationMemberDao creates and returns a new DAO object for table data access.
func NewUcgConversationMemberDao() *UcgConversationMemberDao {
	return &UcgConversationMemberDao{
		group:   "default",
		table:   "ucg_conversation_member",
		columns: ucgConversationMemberColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *UcgConversationMemberDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *UcgConversationMemberDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *UcgConversationMemberDao) Columns() UcgConversationMemberColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *UcgConversationMemberDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *UcgConversationMemberDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *UcgConversationMemberDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
