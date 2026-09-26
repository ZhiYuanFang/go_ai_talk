// =================================================================================
package dao

import (
	"hello/internal/dao/internal"
)

type internalAppointmentNextDao = *internal.AppointmentNextDao

type appointmentNextDao struct {
	internalAppointmentNextDao
}

var (
	// AppointmentNext 宝宝预约事件下次约定表。
	AppointmentNext = appointmentNextDao{
		internal.NewAppointmentNextDao(),
	}
)
