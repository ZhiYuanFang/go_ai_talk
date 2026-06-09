// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UcgChatMessageOutboxDao is the data access object for table ucg_chat_message_outbox.
type UcgChatMessageOutboxDao struct {
	table   string                      // table is the underlying table name of the DAO.
	group   string                      // group is the database configuration group name of current DAO.
	columns UcgChatMessageOutboxColumns // columns contains all the column names of Table for convenient usage.
}

// UcgChatMessageOutboxColumns defines and stores column names for table ucg_chat_message_outbox.
type UcgChatMessageOutboxColumns struct {
	Id             string //
	ConversationId string //
	Payload        string // ChatMessage JSON
	Status         string // pending|done|failed
	Attempts       string //
	LastError      string //
	CreatedAt      string //
}

// ucgChatMessageOutboxColumns holds the columns for table ucg_chat_message_outbox.
var ucgChatMessageOutboxColumns = UcgChatMessageOutboxColumns{
	Id:             "id",
	ConversationId: "conversation_id",
	Payload:        "payload",
	Status:         "status",
	Attempts:       "attempts",
	LastError:      "last_error",
	CreatedAt:      "created_at",
}

// NewUcgChatMessageOutboxDao creates and returns a new DAO object for table data access.
func NewUcgChatMessageOutboxDao() *UcgChatMessageOutboxDao {
	return &UcgChatMessageOutboxDao{
		group:   "default",
		table:   "ucg_chat_message_outbox",
		columns: ucgChatMessageOutboxColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *UcgChatMessageOutboxDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *UcgChatMessageOutboxDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *UcgChatMessageOutboxDao) Columns() UcgChatMessageOutboxColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *UcgChatMessageOutboxDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *UcgChatMessageOutboxDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *UcgChatMessageOutboxDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
