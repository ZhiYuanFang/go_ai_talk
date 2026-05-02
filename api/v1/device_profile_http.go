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

// DeviceProfileSaveReq 保存设备画像（内部/网关可调）。
type DeviceProfileSaveReq struct {
	g.Meta   `path:"/device/profile/api/save" method:"post" tags:"device" summary:"保存设备画像"`
	DeviceNo string `json:"deviceNo" dc:"设备号"`
	Birthday string `json:"birthday" dc:"生日"`
	Sex      int    `json:"sex" dc:"性别"`
}

// DeviceProfileSaveRes 保存成功。
type DeviceProfileSaveRes struct{}
