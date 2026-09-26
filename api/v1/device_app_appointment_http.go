package v1

import "github.com/gogf/gf/v2/frame/g"

// DeviceAppAppointmentNextGetReq 读取宝宝某事件的下次约定时间。
// 业务：预约事件编辑 sheet 回填；与 history 解耦；无行返回 nextAt=0。
type DeviceAppAppointmentNextGetReq struct {
	g.Meta   `path:"/device/app/api/appointment/next" method:"get" tags:"device" summary:"读取预约事件下次约定"`
	DeviceNo string `json:"deviceNo" v:"required" dc:"宝宝设备号"`
	EventId  int64  `json:"eventId" v:"required|min:1" dc:"事件字典 ID"`
}

// DeviceAppAppointmentNextGetRes 下次约定响应。
type DeviceAppAppointmentNextGetRes struct {
	NextAt int64 `json:"nextAt" dc:"下次约定 unix 秒；0=无约定"`
}

// DeviceAppAppointmentNextPutReq upsert 宝宝某事件下次约定。
// nextAt=0 表示清空约定（保留行）；不驱动 predict Redis。
type DeviceAppAppointmentNextPutReq struct {
	g.Meta   `path:"/device/app/api/appointment/next" method:"put" tags:"device" summary:"写入预约事件下次约定"`
	DeviceNo string `json:"deviceNo" v:"required" dc:"宝宝设备号"`
	EventId  int64  `json:"eventId" v:"required|min:1" dc:"事件字典 ID"`
	NextAt   int64  `json:"nextAt" v:"min:0" dc:"下次约定 unix 秒；0=无约定"`
}

// DeviceAppAppointmentNextPutRes 写入成功。
type DeviceAppAppointmentNextPutRes struct{}
