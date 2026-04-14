package v1

import (
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

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
	NeedQuantity int    `json:"needQuantity" dc:"是否需要计数(0否1是)"`
	ExtraNames   string `json:"extraNames" dc:"事件扩展（别名等）"`
}

// DeviceAdminEventAddRes 新增成功。
type DeviceAdminEventAddRes struct{}

// DeviceAdminEventUpdateReq 更新事件名称。
type DeviceAdminEventUpdateReq struct {
	g.Meta       `path:"/device/admin/api/event/update" method:"post" tags:"admin" summary:"更新事件"`
	Id           int64  `json:"id" dc:"事件ID"`
	Name         string `json:"name" dc:"新名称"`
	NeedQuantity int    `json:"needQuantity" dc:"是否需要计数(0否1是)"`
	ExtraNames   string `json:"extraNames" dc:"事件扩展（别名等）"`
}

// DeviceAdminEventUpdateRes 更新成功。
type DeviceAdminEventUpdateRes struct{}

// DeviceAdminEventDeleteReq 删除事件。
type DeviceAdminEventDeleteReq struct {
	g.Meta `path:"/device/admin/api/event/delete" method:"post" tags:"admin" summary:"删除事件"`
	Id     int64 `json:"id" dc:"事件ID"`
}

// DeviceAdminEventDeleteRes 删除成功。
type DeviceAdminEventDeleteRes struct{}

// DeviceAdminQaListReq 问答库历史（qa 表）。
type DeviceAdminQaListReq struct {
	g.Meta `path:"/device/admin/api/qa/list" method:"get" tags:"admin" summary:"问答库列表"`
}

// DeviceAdminQaListRes 问答库列表响应。
type DeviceAdminQaListRes struct {
	List []entity.Qa `json:"list"`
}

// DeviceAdminActionListReq 动作预设列表（action 表）。
type DeviceAdminActionListReq struct {
	g.Meta `path:"/device/admin/api/action/list" method:"get" tags:"admin" summary:"动作预设列表"`
}

// DeviceAdminActionItem 动作预设行（含目标类型中文说明）。
type DeviceAdminActionItem struct {
	Id              int64  `json:"id"`
	Name            string `json:"name"`
	TargetType      string `json:"targetType"`
	TargetTypeLabel string `json:"targetTypeLabel"`
}

// DeviceAdminActionListRes 动作预设列表。
type DeviceAdminActionListRes struct {
	List []DeviceAdminActionItem `json:"list"`
}

// DeviceAdminActionUpdateReq 更新动作预设。
type DeviceAdminActionUpdateReq struct {
	g.Meta     `path:"/device/admin/api/action/update" method:"post" tags:"admin" summary:"更新动作预设"`
	Id         int64  `json:"id" dc:"动作ID"`
	Name       string `json:"name" dc:"动作名"`
	TargetType string `json:"targetType" dc:"目标类型（如 start/end/one 等）"`
}

// DeviceAdminActionUpdateRes 更新成功。
type DeviceAdminActionUpdateRes struct{}

// DeviceAdminActionDeleteReq 删除动作预设。
type DeviceAdminActionDeleteReq struct {
	g.Meta `path:"/device/admin/api/action/delete" method:"post" tags:"admin" summary:"删除动作预设"`
	Id     int64 `json:"id" dc:"动作ID"`
}

// DeviceAdminActionDeleteRes 删除成功。
type DeviceAdminActionDeleteRes struct{}
