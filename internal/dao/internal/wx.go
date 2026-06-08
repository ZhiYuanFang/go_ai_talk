// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// WxDao is the data access object for table wx.
type WxDao struct {
	table   string    // table is the underlying table name of the DAO.
	group   string    // group is the database configuration group name of current DAO.
	columns WxColumns // columns contains all the column names of Table for convenient usage.
}

// WxColumns defines and stores column names for table wx.
type WxColumns struct {
	Id         string //
	DeviceNo   string //
	Unionid    string //
	AppleSub   string // Apple JWT sub
	Platform   string // 平台来源
	IpLocation string // IP属地展示文案（省/市，客户端上报）
	Account    string // 账户
	Password   string // 密码哈希（bcrypt，不可逆）
}

// wxColumns holds the columns for table wx.
var wxColumns = WxColumns{
	Id:         "id",
	DeviceNo:   "device_no",
	Unionid:    "unionid",
	AppleSub:   "apple_sub",
	Platform:   "platform",
	IpLocation: "ip_location",
	Account:    "account",
	Password:   "password",
}

// NewWxDao creates and returns a new DAO object for table data access.
func NewWxDao() *WxDao {
	return &WxDao{
		group:   "default",
		table:   "wx",
		columns: wxColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *WxDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *WxDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *WxDao) Columns() WxColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *WxDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *WxDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *WxDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
