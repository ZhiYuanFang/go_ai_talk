// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UcgPostDao is the data access object for table ucg_post.
type UcgPostDao struct {
	table   string         // table is the underlying table name of the DAO.
	group   string         // group is the database configuration group name of current DAO.
	columns UcgPostColumns // columns contains all the column names of Table for convenient usage.
}

// UcgPostColumns defines and stores column names for table ucg_post.
type UcgPostColumns struct {
	Id                string //
	AuthorWxId        string //
	Type              string // moment|debate
	Content           string //
	DebateLeftLabel   string //
	DebateRightLabel  string //
	IpLocation        string // 发帖时 IP属地快照
	Lat               string // 可选发帖纬度 WGS84
	Lng               string // 可选发帖经度 WGS84
	Status            string // 0 draft 1 pending_audit 2 published 3 rejected 4 apply_failed
	AuditVersion      string // 审核轮次
	ModerationVerdict string //
	ModerationReason  string //
	ModerationAt      string //
	ApplyAttempts     string //
	ApplyFailedAt     string //
	RejectReason      string //
	MediaType         string // 0 none 1 images 2 video
	LikeCount         string //
	CommentCount      string //
	CreatedAt         string //
	UpdatedAt         string //
	PublishedAt       string //
}

// ucgPostColumns holds the columns for table ucg_post.
var ucgPostColumns = UcgPostColumns{
	Id:                "id",
	AuthorWxId:        "author_wx_id",
	Type:              "type",
	Content:           "content",
	DebateLeftLabel:   "debate_left_label",
	DebateRightLabel:  "debate_right_label",
	IpLocation:        "ip_location",
	Lat:               "lat",
	Lng:               "lng",
	Status:            "status",
	AuditVersion:      "audit_version",
	ModerationVerdict: "moderation_verdict",
	ModerationReason:  "moderation_reason",
	ModerationAt:      "moderation_at",
	ApplyAttempts:     "apply_attempts",
	ApplyFailedAt:     "apply_failed_at",
	RejectReason:      "reject_reason",
	MediaType:         "media_type",
	LikeCount:         "like_count",
	CommentCount:      "comment_count",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
	PublishedAt:       "published_at",
}

// NewUcgPostDao creates and returns a new DAO object for table data access.
func NewUcgPostDao() *UcgPostDao {
	return &UcgPostDao{
		group:   "default",
		table:   "ucg_post",
		columns: ucgPostColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *UcgPostDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *UcgPostDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *UcgPostDao) Columns() UcgPostColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *UcgPostDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *UcgPostDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *UcgPostDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
