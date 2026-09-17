package v1

import "github.com/gogf/gf/v2/frame/g"

// PushAdminDevicesListReq GET 管理端推送设备分页列表（须 X-Admin-Password）。
type PushAdminDevicesListReq struct {
	g.Meta      `path:"/push/admin/api/devices" method:"get" tags:"push-admin" summary:"管理端推送设备列表"`
	Page        int    `json:"page" in:"query" d:"1" dc:"页码，从 1 起"`
	PageSize    int    `json:"pageSize" in:"query" d:"20" dc:"每页条数，最大 200"`
	WxId        int64  `json:"wxId" in:"query" dc:"可选，精确过滤 wx 主键"`
	Channel     string `json:"channel" in:"query" dc:"可选：apns|hms|mipush"`
	TokenPrefix string `json:"tokenPrefix" in:"query" dc:"可选，token 前缀（至少 4 字符）"`
}

// PushAdminDeviceItem 管理端推送设备行。
type PushAdminDeviceItem struct {
	Id        uint64 `json:"id"`
	WxId      int64  `json:"wxId"`
	Channel   string `json:"channel"`
	DeviceKey string `json:"deviceKey"`
	Token     string `json:"token" dc:"完整 token；页面默认截断展示"`
	UpdatedAt int64  `json:"updatedAt" dc:"unix 秒"`
}

// PushAdminDevicesListRes 管理端推送设备分页 data。
type PushAdminDevicesListRes struct {
	List     []PushAdminDeviceItem `json:"list"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"pageSize"`
	Total    int                   `json:"total"`
}

// PushAdminDeviceDeleteReq DELETE 按 id 删除推送设备行。
type PushAdminDeviceDeleteReq struct {
	g.Meta `path:"/push/admin/api/devices/{id}" method:"delete" tags:"push-admin" summary:"管理端删除推送设备"`
	Id     uint64 `json:"id" in:"path" v:"required|min:1" dc:"push_device.id"`
}

// PushAdminDeviceDeleteRes 删除成功。
type PushAdminDeviceDeleteRes struct{}
