// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// VersionDao is the data access object for table version.
type VersionDao struct {
	table   string         // table is the underlying table name of the DAO.
	group   string         // group is the database configuration group name of current DAO.
	columns VersionColumns // columns contains all the column names of Table for convenient usage.
}

// VersionColumns defines and stores column names for table version.
type VersionColumns struct {
	Id            string //
	LatestVersion string // 线上最新版本号
	ForceUpdate   string // 整包强制更新开关
	ReleaseNotes  string // 更新内容
	DownloadUrl   string // 下载链接
	MinVersion    string // 最低支持版本
	ReleaseDate   string // 发布时间戳
}

// versionColumns holds the columns for table version.
var versionColumns = VersionColumns{
	Id:            "id",
	LatestVersion: "latest_version",
	ForceUpdate:   "force_update",
	ReleaseNotes:  "release_notes",
	DownloadUrl:   "download_url",
	MinVersion:    "min_version",
	ReleaseDate:   "release_date",
}

// NewVersionDao creates and returns a new DAO object for table data access.
func NewVersionDao() *VersionDao {
	return &VersionDao{
		group:   "default",
		table:   "version",
		columns: versionColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *VersionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *VersionDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *VersionDao) Columns() VersionColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *VersionDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *VersionDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *VersionDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
