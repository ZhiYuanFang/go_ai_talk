package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// WxDao wx 表访问对象。
type WxDao struct {
	table   string
	group   string
	columns WxColumns
}

type WxColumns struct {
	Id       string
	WxCode   string
	DeviceNo string
	Platform string
}

var wxColumns = WxColumns{
	Id:       "id",
	WxCode:   "wx_code",
	DeviceNo: "device_no",
	Platform: "platform",
}

func NewWxDao() *WxDao {
	return &WxDao{
		group:   "default",
		table:   "wx",
		columns: wxColumns,
	}
}

func (dao *WxDao) DB() gdb.DB {
	return g.DB(dao.group)
}

func (dao *WxDao) Table() string {
	return dao.table
}

func (dao *WxDao) Columns() WxColumns {
	return dao.columns
}

func (dao *WxDao) Group() string {
	return dao.group
}

func (dao *WxDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

func (dao *WxDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
