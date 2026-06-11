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
	ActiveTime int64 `json:"activeTime" dc:"激活时间，Unix 秒"`
}

// DeviceAdminListReq 设备列表（仅需 Header 口令）。
type DeviceAdminListReq struct {
	g.Meta `path:"/device/admin/api/list" method:"get" tags:"admin" summary:"设备列表"`
}

// DeviceAdminListRes 设备列表响应。
type DeviceAdminListRes struct {
	List []entity.User `json:"list"`
}

// DeviceAdminUserListReq 设备记录分页列表（user 表）。
type DeviceAdminUserListReq struct {
	g.Meta   `path:"/device/admin/api/user/list" method:"get" tags:"admin" summary:"设备记录分页列表"`
	Page     int    `json:"page" p:"page" dc:"页码，从 1 开始"`
	PageSize int    `json:"pageSize" p:"pageSize" dc:"每页条数，默认 5，最大 100"`
	Q        string `json:"q" p:"q" dc:"设备号模糊搜索"`
}

// DeviceAdminUserListRes 设备记录分页响应。
type DeviceAdminUserListRes struct {
	List     []entity.User `json:"list"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

// DeviceAdminWxListReq wx 账号分页列表。
type DeviceAdminWxListReq struct {
	g.Meta   `path:"/device/admin/api/wx/list" method:"get" tags:"admin" summary:"wx 账号分页列表"`
	Page     int    `json:"page" p:"page" dc:"页码，从 1 开始"`
	PageSize int    `json:"pageSize" p:"pageSize" dc:"每页条数，默认 20，最大 100"`
	Q        string `json:"q" p:"q" dc:"id/deviceNo/unionid/account 模糊搜索"`
}

// DeviceAdminWxListItem wx 列表项（不含 password）。
type DeviceAdminWxListItem struct {
	Id       int64  `json:"id"`
	DeviceNo string `json:"deviceNo"`
	Unionid  string `json:"unionid"`
	Platform string `json:"platform"`
	Account  string `json:"account"`
}

// DeviceAdminWxListRes wx 账号分页响应。
type DeviceAdminWxListRes struct {
	List     []DeviceAdminWxListItem `json:"list"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

// DeviceAdminEventListReq 事件字典列表。
type DeviceAdminEventListReq struct {
	g.Meta `path:"/device/admin/api/event/list" method:"get" tags:"admin" summary:"事件列表"`
}

// DeviceAdminEventListRes 事件列表响应。
type DeviceAdminEventListRes struct {
	List []entity.Event `json:"list"`
}

// DeviceAdminEventAddReq 新增事件（实现为 multipart，见 device_admin_event.go，不在此 Bind）。
type DeviceAdminEventAddReq struct {
	Name       string `dc:"事件名称，表单字段 name"`
	EventType  string `dc:"事件类型 number|time|one，表单 eventType"`
	ExtraNames string `dc:"事件扩展，表单 extraNames"`
	Color      string `dc:"色值 #RGB/#RRGGBB，表单 color"`
	Logo       string `dc:"可选，表单文件字段 logo"`
	ParentId   int64  `dc:"父事件 ID，0 为根；表单 parentId"`
}

// DeviceAdminEventAddRes 新增成功。
type DeviceAdminEventAddRes struct{}

// DeviceAdminEventUpdateReq 更新事件（multipart，见 device_admin_event.go）。
type DeviceAdminEventUpdateReq struct {
	Id         int64  `dc:"事件ID，表单 id"`
	Name       string `dc:"表单 name"`
	EventType  string `dc:"表单 eventType"`
	ExtraNames string `dc:"表单 extraNames"`
	Color      string `dc:"表单 color"`
	Logo       string `dc:"可选文件 logo"`
	ParentId   int64  `dc:"父事件 ID，0 为根；编辑时表单 parentId，省略表示不修改"`
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

// DeviceAdminQaListReq 问答库分页列表（qa 表，id 倒序）。
type DeviceAdminQaListReq struct {
	g.Meta   `path:"/device/admin/api/qa/list" method:"get" tags:"admin" summary:"问答库列表"`
	Page     int `json:"page" p:"page" dc:"页码，从 1 开始"`
	PageSize int `json:"pageSize" p:"pageSize" dc:"每页条数，默认 10，最大 100"`
}

// DeviceAdminQaListRes 问答库分页响应。
type DeviceAdminQaListRes struct {
	List     []entity.Qa `json:"list"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// DeviceAdminQaDeleteReq 删除问答库行。
type DeviceAdminQaDeleteReq struct {
	g.Meta `path:"/device/admin/api/qa/delete" method:"post" tags:"admin" summary:"删除问答库行"`
	Id     int64 `json:"id" dc:"问答库主键 id"`
}

// DeviceAdminQaDeleteRes 删除成功。
type DeviceAdminQaDeleteRes struct{}

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
