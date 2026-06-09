// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// HistoryDao is the data access object for table history.
type HistoryDao struct {
	table   string         // table is the underlying table name of the DAO.
	group   string         // group is the database configuration group name of current DAO.
	columns HistoryColumns // columns contains all the column names of Table for convenient usage.
}

// HistoryColumns defines and stores column names for table history.
type HistoryColumns struct {
	Id          string //
	DeviceNo    string // 设备号
	EventId     string // 事件id
	EventName   string // 事件名
	EventNumber string // 奶量等
	StartTime   string // 开始时间戳
	EndTime     string // 结束时间戳
	Remark      string // 备注
	PostId      string // 关联 UCG 帖子
	MediaType   string // 媒体类型
	ImageKeys   string // 图片 objectKey JSON 数组
	VideoKey    string // 视频 objectKey
}

// historyColumns holds the columns for table history.
var historyColumns = HistoryColumns{
	Id:          "id",
	DeviceNo:    "device_no",
	EventId:     "event_id",
	EventName:   "event_name",
	EventNumber: "event_number",
	StartTime:   "start_time",
	EndTime:     "end_time",
	Remark:      "remark",
	PostId:      "post_id",
	MediaType:   "media_type",
	ImageKeys:   "image_keys",
	VideoKey:    "video_key",
}

// NewHistoryDao creates and returns a new DAO object for table data access.
func NewHistoryDao() *HistoryDao {
	return &HistoryDao{
		group:   "default",
		table:   "history",
		columns: historyColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *HistoryDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *HistoryDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *HistoryDao) Columns() HistoryColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *HistoryDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *HistoryDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *HistoryDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
