package v1

import "github.com/gogf/gf/v2/frame/g"
import "hello/internal/model/entity"

// 管理端接口均需在 Header 携带 X-Admin-Password（口令不在 JSON 体内）。

// DeviceAdminRegisterReq 注册设备。
type DeviceAdminRegisterReq struct {
	g.Meta   `path:"/device/admin/api/register" method:"post" tags:"admin" summary:"注册设备"`
	DeviceNo string `json:"deviceNo" dc:"设备号"`
}

// DeviceAdminRegisterRes 注册成功返回。
type DeviceAdminRegisterRes struct {
	DeviceNo   string `json:"deviceNo"`
	ActiveTime string `json:"activeTime"`
}

// DeviceAdminListReq 设备列表（仅需 Header 口令）。
type DeviceAdminListReq struct {
	g.Meta `path:"/device/admin/api/list" method:"get" tags:"admin" summary:"设备列表"`
}

// DeviceAdminListRes 设备列表响应。
type DeviceAdminListRes struct {
	List []entity.User `json:"list"`
}

// DeviceAdminEventListReq 事件字典列表。
type DeviceAdminEventListReq struct {
	g.Meta `path:"/device/admin/api/event/list" method:"get" tags:"admin" summary:"事件列表"`
}

// DeviceAdminEventListRes 事件列表响应。
type DeviceAdminEventListRes struct {
	List []entity.Event `json:"list"`
}

// DeviceAdminEventAddReq 新增事件。
type DeviceAdminEventAddReq struct {
	g.Meta       `path:"/device/admin/api/event/add" method:"post" tags:"admin" summary:"新增事件"`
	Name         string `json:"name" dc:"事件名称"`
	NeedTime     int    `json:"needTime" dc:"是否需要计时(0否1是)"`
	NeedQuantity int    `json:"needQuantity" dc:"是否需要计数(0否1是)"`
}

// DeviceAdminEventAddRes 新增成功。
type DeviceAdminEventAddRes struct{}

// DeviceAdminEventUpdateReq 更新事件名称。
type DeviceAdminEventUpdateReq struct {
	g.Meta       `path:"/device/admin/api/event/update" method:"post" tags:"admin" summary:"更新事件"`
	Id           int64  `json:"id" dc:"事件ID"`
	Name         string `json:"name" dc:"新名称"`
	NeedTime     int    `json:"needTime" dc:"是否需要计时(0否1是)"`
	NeedQuantity int    `json:"needQuantity" dc:"是否需要计数(0否1是)"`
}

// DeviceAdminEventUpdateRes 更新成功。
type DeviceAdminEventUpdateRes struct{}

// DeviceAdminIntentionListReq 意图列表。
type DeviceAdminIntentionListReq struct {
	g.Meta `path:"/device/admin/api/intention/list" method:"get" tags:"admin" summary:"意图列表"`
}

// DeviceAdminIntentionListRes 意图列表响应。
type DeviceAdminIntentionListRes struct {
	List []entity.Intention `json:"list"`
}

// DeviceAdminIntentionUpdateReq 更新意图动态历史上限。
type DeviceAdminIntentionUpdateReq struct {
	g.Meta     `path:"/device/admin/api/intention/update" method:"post" tags:"admin" summary:"更新意图历史上限"`
	Id         int64 `json:"id" dc:"意图ID"`
	UpperLimit int   `json:"upperLimit" dc:"动态历史消息上限(>=0)"`
}

// DeviceAdminIntentionUpdateRes 更新成功。
type DeviceAdminIntentionUpdateRes struct{}
