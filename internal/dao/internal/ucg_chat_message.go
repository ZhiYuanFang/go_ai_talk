// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UcgChatMessageDao is the data access object for table ucg_chat_message.
type UcgChatMessageDao struct {
	table   string                // table is the underlying table name of the DAO.
	group   string                // group is the database configuration group name of current DAO.
	columns UcgChatMessageColumns // columns contains all the column names of Table for convenient usage.
}

// UcgChatMessageColumns defines and stores column names for table ucg_chat_message.
type UcgChatMessageColumns struct {
	Id             string //
	ConversationId string //
	ClientMsgId    string // 客户端幂等 ID，空表示无
	SenderWxId     string //
	Content        string //
	ImageKey       string //
	VideoKey       string //
	MediaCdnUrl    string //
	CreatedAt      string //
	Status         string //
}

var ucgChatMessageColumns = UcgChatMessageColumns{
	Id:             "id",
	ConversationId: "conversation_id",
	ClientMsgId:    "client_msg_id",
	SenderWxId:     "sender_wx_id",
	Content:        "content",
	ImageKey:       "image_key",
	VideoKey:       "video_key",
	MediaCdnUrl:    "media_cdn_url",
	CreatedAt:      "created_at",
	Status:         "status",
}

// NewUcgChatMessageDao creates and returns a new DAO object for table data access.
func NewUcgChatMessageDao() *UcgChatMessageDao {
	return &UcgChatMessageDao{
		group:   "default",
		table:   "ucg_chat_message",
		columns: ucgChatMessageColumns,
	}
}

func (dao *UcgChatMessageDao) DB() gdb.DB {
	return g.DB(dao.group)
}

func (dao *UcgChatMessageDao) Table() string {
	return dao.table
}

func (dao *UcgChatMessageDao) Columns() UcgChatMessageColumns {
	return dao.columns
}

func (dao *UcgChatMessageDao) Group() string {
	return dao.group
}

func (dao *UcgChatMessageDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

func (dao *UcgChatMessageDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
