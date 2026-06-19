package v1

import "github.com/gogf/gf/v2/frame/g"

// DeviceSimUsernameRegisterReq 内部注册模拟用户。
type DeviceSimUsernameRegisterReq struct {
	g.Meta   `path:"/device/internal/api/sim/username/register" method:"post" tags:"device" summary:"内部-注册模拟用户"`
	Account  string `json:"account" v:"required|length:4,32"`
	Password string `json:"password" v:"required|length:6,64"`
}

type DeviceSimUsernameRegisterRes struct {
	WxId    int64  `json:"wxId"`
	Account string `json:"account"`
}

// DeviceSimWxListReq 模拟用户 wxId 分页列表。
type DeviceSimWxListReq struct {
	g.Meta   `path:"/device/internal/api/sim/wx/list" method:"get" tags:"device" summary:"内部-模拟用户 wx 列表"`
	Page     int `json:"page" in:"query" d:"1"`
	PageSize int `json:"pageSize" in:"query" d:"20"`
}

type DeviceSimWxListItem struct {
	WxId    int64  `json:"wxId"`
	Account string `json:"account"`
}

type DeviceSimWxListRes struct {
	List     []DeviceSimWxListItem `json:"list"`
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"pageSize"`
}
