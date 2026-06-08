// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UcgMediaBlobDao is the data access object for table ucg_media_blob.
type UcgMediaBlobDao struct {
	table   string              // table is the underlying table name of the DAO.
	group   string              // group is the database configuration group name of current DAO.
	columns UcgMediaBlobColumns // columns contains all the column names of Table for convenient usage.
}

// UcgMediaBlobColumns defines and stores column names for table ucg_media_blob.
type UcgMediaBlobColumns struct {
	Id               string //
	ContentHash      string // SHA-256 hex lowercase
	TransformVersion string // client transform pipeline version
	ObjectKey        string // OSS object key
	MediaKind        string // 1=image 2=video
	RefCount         string // ownership registrations
	CreatedAt        string //
}

// ucgMediaBlobColumns holds the columns for table ucg_media_blob.
var ucgMediaBlobColumns = UcgMediaBlobColumns{
	Id:               "id",
	ContentHash:      "content_hash",
	TransformVersion: "transform_version",
	ObjectKey:        "object_key",
	MediaKind:        "media_kind",
	RefCount:         "ref_count",
	CreatedAt:        "created_at",
}

// NewUcgMediaBlobDao creates and returns a new DAO object for table data access.
func NewUcgMediaBlobDao() *UcgMediaBlobDao {
	return &UcgMediaBlobDao{
		group:   "default",
		table:   "ucg_media_blob",
		columns: ucgMediaBlobColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *UcgMediaBlobDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *UcgMediaBlobDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *UcgMediaBlobDao) Columns() UcgMediaBlobColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *UcgMediaBlobDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *UcgMediaBlobDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *UcgMediaBlobDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
