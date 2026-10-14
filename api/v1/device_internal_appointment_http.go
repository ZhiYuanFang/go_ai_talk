package v1

import "github.com/gogf/gf/v2/frame/g"

// DeviceInternalAppointmentNextGetReq 内部读取宝宝某事件下次约定（供 Python 预测兜底）。
// 须经 InternalSecretMiddleware；与 App GET 字段对齐；无行 nextAt=0。
type DeviceInternalAppointmentNextGetReq struct {
	g.Meta   `path:"/device/internal/api/appointment/next" method:"get" tags:"device" summary:"内部-预约事件下次约定"`
	DeviceNo string `json:"deviceNo" in:"query" v:"required" dc:"宝宝设备号"`
	EventId  int64  `json:"eventId" in:"query" v:"required|min:1" dc:"事件字典 ID（根或叶均可，调用方归根）"`
}

// DeviceInternalAppointmentNextGetRes 内部读响应。
type DeviceInternalAppointmentNextGetRes struct {
	NextAt int64 `json:"nextAt" dc:"下次约定 unix 秒；0=无约定"`
}
