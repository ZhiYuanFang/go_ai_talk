// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UcgNotificationDao is the data access object for table ucg_notification.
type UcgNotificationDao struct {
	table   string                 // table is the underlying table name of the DAO.
	group   string                 // group is the database configuration group name of current DAO.
	columns UcgNotificationColumns // columns contains all the column names of Table for convenient usage.
}

// UcgNotificationColumns defines and stores column names for table ucg_notification.
type UcgNotificationColumns struct {
	Id            string //
	RecipientWxId string // 接收者 wxId
	Type          string // comment_on_post | mention_in_comment
	PostId        string //
	CommentId     string //
	ActorWxId     string // 评论者
	Preview       string // 评论摘要
	PostThumbUrl  string // 写入时快照的帖子封面 URL
	PostMediaKind string // 0=none,1=image,2=video
	ReadAt        string // NULL=未读
	CreatedAt     string //
}

// ucgNotificationColumns holds the columns for table ucg_notification.
var ucgNotificationColumns = UcgNotificationColumns{
	Id:            "id",
	RecipientWxId: "recipient_wx_id",
	Type:          "type",
	PostId:        "post_id",
	CommentId:     "comment_id",
	ActorWxId:     "actor_wx_id",
	Preview:       "preview",
	PostThumbUrl:  "post_thumb_url",
	PostMediaKind: "post_media_kind",
	ReadAt:        "read_at",
	CreatedAt:     "created_at",
}

// NewUcgNotificationDao creates and returns a new DAO object for table data access.
func NewUcgNotificationDao() *UcgNotificationDao {
	return &UcgNotificationDao{
		group:   "default",
		table:   "ucg_notification",
		columns: ucgNotificationColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *UcgNotificationDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *UcgNotificationDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *UcgNotificationDao) Columns() UcgNotificationColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *UcgNotificationDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *UcgNotificationDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *UcgNotificationDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
