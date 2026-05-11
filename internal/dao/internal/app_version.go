package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AppVersionDao ai_voice_app.version 表。
type AppVersionDao struct {
	table   string
	group   string
	columns AppVersionColumns
}

type AppVersionColumns struct {
	Id            string
	LatestVersion string
	ReleaseNotes  string
	DownloadUrl   string
	ForceUpdate   string
	MinVersion    string
	ReleaseDate   string // 当前版本上线时间：Unix 秒时间戳（库列 release_date，bigint）
}

var appVersionColumns = AppVersionColumns{
	Id:            "id",
	LatestVersion: "latest_version",
	ReleaseNotes:  "release_notes",
	DownloadUrl:   "download_url",
	ForceUpdate:   "force_update",
	MinVersion:    "min_version",
	ReleaseDate:   "release_date",
}

func NewAppVersionDao() *AppVersionDao {
	return &AppVersionDao{
		group:   "app",
		table:   "version",
		columns: appVersionColumns,
	}
}

func (dao *AppVersionDao) DB() gdb.DB {
	return g.DB(dao.group)
}

func (dao *AppVersionDao) Table() string {
	return dao.table
}

func (dao *AppVersionDao) Columns() AppVersionColumns {
	return dao.columns
}

func (dao *AppVersionDao) Group() string {
	return dao.group
}

func (dao *AppVersionDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}
