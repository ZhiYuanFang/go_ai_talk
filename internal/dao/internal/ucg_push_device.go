// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UcgPushDeviceDao is the data access object for table ucg_push_device.
type UcgPushDeviceDao struct {
	table   string
	group   string
	columns UcgPushDeviceColumns
}

// UcgPushDeviceColumns defines and stores column names for table ucg_push_device.
type UcgPushDeviceColumns struct {
	Id        string
	WxId      string
	Channel   string
	Token     string
	DeviceKey string
	UpdatedAt string
}

var ucgPushDeviceColumns = UcgPushDeviceColumns{
	Id:        "id",
	WxId:      "wx_id",
	Channel:   "channel",
	Token:     "token",
	DeviceKey: "device_key",
	UpdatedAt: "updated_at",
}

// NewUcgPushDeviceDao creates and returns a new DAO object for table data access.
func NewUcgPushDeviceDao() *UcgPushDeviceDao {
	return &UcgPushDeviceDao{
		group:   "default",
		table:   "ucg_push_device",
		columns: ucgPushDeviceColumns,
	}
}

func (dao *UcgPushDeviceDao) DB() gdb.DB {
	return g.DB(dao.group)
}

func (dao *UcgPushDeviceDao) Table() string {
	return dao.table
}

func (dao *UcgPushDeviceDao) Columns() UcgPushDeviceColumns {
	return dao.columns
}

func (dao *UcgPushDeviceDao) Group() string {
	return dao.group
}

func (dao *UcgPushDeviceDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

func (dao *UcgPushDeviceDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
