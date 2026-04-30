package v1

import "github.com/gogf/gf/v2/frame/g"

// DeviceProfileGetReq 查询设备画像（内部服务调用）。
type DeviceProfileGetReq struct {
	g.Meta   `path:"/device/profile/api/get" method:"get" tags:"device" summary:"查询设备画像"`
	DeviceNo string `json:"deviceNo" p:"deviceNo" dc:"设备号"`
}

// DeviceProfileGetRes 设备画像响应。
type DeviceProfileGetRes struct {
	DeviceNo string `json:"deviceNo"`
	Birthday string `json:"birthday"`
	Sex      int    `json:"sex"`
}
