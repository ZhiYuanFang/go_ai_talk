package v1

import "github.com/gogf/gf/v2/frame/g"

// ---------- App 用户域：画像（扁平路径 /device/app/api/user/*） ----------

// DeviceProfileGetReq 查询设备画像（内部服务调用）。
type DeviceProfileGetReq struct {
	g.Meta   `path:"/device/app/api/user/get" method:"get" tags:"device" summary:"查询设备画像"`
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
	g.Meta   `path:"/device/app/api/user/save" method:"post" tags:"device" summary:"保存设备画像"`
	DeviceNo string `json:"deviceNo" dc:"设备号"`
	Birthday string `json:"birthday" dc:"生日"`
	Sex      int    `json:"sex" dc:"性别"`
}

// DeviceProfileSaveRes 保存成功。
type DeviceProfileSaveRes struct{}

// DeviceProfileBindWxReq 绑定设备到当前 wx（wxCode 来自 Header X-Internal-Wx-Code）。
type DeviceProfileBindWxReq struct {
	g.Meta   `path:"/device/app/api/user/bindwx" method:"post" tags:"device" summary:"wx 绑定设备"`
	DeviceNo string `json:"deviceNo" dc:"设备号"`
}

// DeviceProfileBindWxRes 绑定成功。
type DeviceProfileBindWxRes struct{}

// DeviceProfileAutoSaveReq 自动保存画像（wxCode 来自 Header）。
type DeviceProfileAutoSaveReq struct {
	g.Meta   `path:"/device/app/api/user/auto_save" method:"post" tags:"device" summary:"自动保存画像并返回设备号"`
	Birthday string `json:"birthday" dc:"生日"`
	Sex      int    `json:"sex"      dc:"性别"`
}

// DeviceProfileAutoSaveRes 返回设备号。
type DeviceProfileAutoSaveRes struct {
	DeviceNo string `json:"deviceNo"`
}

// ---------- App 用户域：微信（与网关聚合 POST /device/app/api/login 区分） ----------

// DeviceWxLoginReq 微信登录（仅业务，不返回 JWT）。
type DeviceWxLoginReq struct {
	g.Meta   `path:"/device/app/api/user/login" method:"post" tags:"device" summary:"微信登录业务"`
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
	g.Meta `path:"/device/app/api/user/detail" method:"get" tags:"device" summary:"wx 绑定设备号"`
}

// DeviceWxDetailRes wx 详情。
type DeviceWxDetailRes struct {
	DeviceNo string `json:"deviceNo"`
}

// DeviceWxInternalByIDReq 网关内部：按 wx 主键取 wxCode。
type DeviceWxInternalByIDReq struct {
	g.Meta `path:"/device/app/api/user/internal/by-id" method:"get" tags:"device" summary:"内部按 id 取 wxCode"`
	Id     int64 `json:"id" p:"id" dc:"wx 表主键"`
}

// DeviceWxInternalByIDRes 内部查询响应。
type DeviceWxInternalByIDRes struct {
	WxCode string `json:"wxCode"`
}
