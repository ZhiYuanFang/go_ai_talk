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

func (dao *UcgChatMessageOutboxDao) DB() gdb.DB {
	return g.DB(dao.group)
}

func (dao *UcgChatMessageOutboxDao) Table() string {
	return dao.table
}

func (dao *UcgChatMessageOutboxDao) Columns() UcgChatMessageOutboxColumns {
	return dao.columns
}

func (dao *UcgChatMessageOutboxDao) Group() string {
	return dao.group
}

func (dao *UcgChatMessageOutboxDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

func (dao *UcgChatMessageOutboxDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
