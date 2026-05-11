package v1

import "github.com/gogf/gf/v2/frame/g"

// DeviceWxLoginReq 微信登录（仅业务，不返回 JWT）。
type DeviceWxLoginReq struct {
	g.Meta   `path:"/device/wx/api/login" method:"post" tags:"device" summary:"微信登录业务"`
	WxCode   string `json:"wxCode"   dc:"微信侧 code"`
	Platform string `json:"platform" dc:"平台标识"`
}

// DeviceWxLoginRes 微信登录业务响应。
type DeviceWxLoginRes struct {
	WxId      int64  `json:"wxId"`
	WxCode    string `json:"wxCode"`
	DeviceNo  string `json:"deviceNo"`
	IsNewUser bool   `json:"isNewUser"`
}

// DeviceWxDetailReq 按 wx 查询已绑定设备号（wxCode 来自 Header X-Internal-Wx-Code）。
type DeviceWxDetailReq struct {
	g.Meta `path:"/device/wx/api/detail" method:"get" tags:"device" summary:"wx 绑定设备号"`
}

// DeviceWxDetailRes wx 详情。
type DeviceWxDetailRes struct {
	DeviceNo string `json:"deviceNo"`
}

// DeviceWxInternalByIDReq 网关内部：按 wx 主键取 wxCode。
type DeviceWxInternalByIDReq struct {
	g.Meta `path:"/device/wx/api/internal/by-id" method:"get" tags:"device" summary:"内部按 id 取 wxCode"`
	Id     int64 `json:"id" p:"id" dc:"wx 表主键"`
}

// DeviceWxInternalByIDRes 内部查询响应。
type DeviceWxInternalByIDRes struct {
	WxCode string `json:"wxCode"`
}
