// ==========================================================================
// 预约下次约定表 appointment_next 的 DAO（手写，与 event 同属 device 库）。
// 业务：按 (device_no, event_id) 持久化客户端「下次约定」unix 秒；next_at=0 表示无约定。
// 设计：保留行值为 0（不清行），便于 upsert 覆盖；与 predict Redis 无关。
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AppointmentNextDao is the data access object for table appointment_next.
type AppointmentNextDao struct {
	table   string
	group   string
	columns AppointmentNextColumns
}

// AppointmentNextColumns 列名。
type AppointmentNextColumns struct {
	Id       string
	DeviceNo string
	EventId  string
	NextAt   string
}

var appointmentNextColumns = AppointmentNextColumns{
	Id:       "id",
	DeviceNo: "device_no",
	EventId:  "event_id",
	NextAt:   "next_at",
}

// NewAppointmentNextDao creates and returns a new DAO object.
func NewAppointmentNextDao() *AppointmentNextDao {
	return &AppointmentNextDao{
		group:   "default",
		table:   "appointment_next",
		columns: appointmentNextColumns,
	}
}

func (dao *AppointmentNextDao) DB() gdb.DB {
	return g.DB(dao.group)
}

func (dao *AppointmentNextDao) Table() string {
	return dao.table
}

func (dao *AppointmentNextDao) Columns() AppointmentNextColumns {
	return dao.columns
}

func (dao *AppointmentNextDao) Group() string {
	return dao.group
}

func (dao *AppointmentNextDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

func (dao *AppointmentNextDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
