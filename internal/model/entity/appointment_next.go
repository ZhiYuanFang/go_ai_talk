// =================================================================================
// appointment_next 实体：宝宝某事件的下次约定时间（客户端记忆副本）。
// =================================================================================

package entity

// AppointmentNext is the golang structure for table appointment_next.
type AppointmentNext struct {
	Id       int64  `json:"id"       ` //
	DeviceNo string `json:"deviceNo" ` // 宝宝设备号
	EventId  int64  `json:"eventId"  ` // 事件字典 ID
	NextAt   int64  `json:"nextAt"   ` // 下次约定 unix 秒；0=无约定
}
