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
	BabyName string `json:"babyName" dc:"宝宝名字"`
	Birthday int64  `json:"birthday" dc:"生日，Unix 秒"`
	Sex      int    `json:"sex"`
}

// DeviceProfileSaveReq 保存设备画像（内部/网关可调）。
type DeviceProfileSaveReq struct {
	g.Meta   `path:"/device/app/api/user/save" method:"post" tags:"device" summary:"保存设备画像"`
	DeviceNo string `json:"deviceNo" dc:"设备号"`
	BabyName string `json:"babyName" dc:"宝宝名字"`
	Birthday int64  `json:"birthday" dc:"生日，Unix 秒"`
	Sex      int    `json:"sex" dc:"性别"`
}

// DeviceProfileSaveRes 保存成功。
type DeviceProfileSaveRes struct{}

// DeviceProfileBindWxReq 绑定设备到当前 wx（wx 主键来自 Header X-Internal-Wx-Id，由 gateway-app 从 JWT sub 注入）。
type DeviceProfileBindWxReq struct {
	g.Meta   `path:"/device/app/api/user/bindwx" method:"post" tags:"device" summary:"wx 绑定设备"`
	DeviceNo string `json:"deviceNo" dc:"设备号"`
}

// DeviceProfileBindWxRes 绑定成功。
type DeviceProfileBindWxRes struct{}

// DeviceProfileAutoSaveReq 自动保存画像（wx 主键来自 Header X-Internal-Wx-Id）。
type DeviceProfileAutoSaveReq struct {
	g.Meta   `path:"/device/app/api/user/auto_save" method:"post" tags:"device" summary:"自动保存画像并返回设备号"`
	BabyName string `json:"babyName" dc:"宝宝名字"`
	Birthday int64  `json:"birthday" dc:"生日，Unix 秒"`
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
	JsCode   string `json:"jsCode"   dc:"微信开放平台授权临时 code（移动应用 SendAuth 或网站 qrconnect），服务端 OAuth 换 unionid"`
	Platform string `json:"platform" dc:"与 wechat.platforms 配置键一致，如 ios/android/web"`
}

// DeviceWxLoginRes 微信登录业务响应。
type DeviceWxLoginRes struct {
	WxId      int64  `json:"wxId"`
	DeviceNo  string `json:"deviceNo"`
	IsNewUser bool   `json:"isNewUser"`
}

// DeviceUsernameRegisterReq 用户名注册（仅业务，不返回 JWT）。
type DeviceUsernameRegisterReq struct {
	g.Meta   `path:"/device/app/api/user/username/register" method:"post" tags:"device" summary:"用户名注册业务"`
	Account  string `json:"account" dc:"账号"`
	Password string `json:"password" dc:"密码明文（服务端哈希入库）"`
}

// DeviceUsernameRegisterRes 用户名注册业务响应。
type DeviceUsernameRegisterRes struct {
	WxId int64 `json:"wxId"`
}

// DeviceUsernameLoginReq 用户名密码登录（仅业务，不返回 JWT）。
type DeviceUsernameLoginReq struct {
	g.Meta   `path:"/device/app/api/user/username/login" method:"post" tags:"device" summary:"用户名登录业务"`
	Account  string `json:"account" dc:"账号"`
	Password string `json:"password" dc:"密码明文"`
}

// DeviceUsernameLoginRes 用户名密码登录业务响应。
type DeviceUsernameLoginRes struct {
	WxId      int64  `json:"wxId"`
	DeviceNo  string `json:"deviceNo"`
	IsNewUser bool   `json:"isNewUser"`
}

// DeviceUsernameBindWxReq 用户名账号绑定微信（当前账号来自 Header X-Internal-Wx-Id）。
type DeviceUsernameBindWxReq struct {
	g.Meta   `path:"/device/app/api/user/username/bindwx" method:"post" tags:"device" summary:"用户名账号绑定微信"`
	JsCode   string `json:"jsCode" dc:"微信开放平台授权临时 code"`
	Platform string `json:"platform" dc:"平台标识（ios/android/web）"`
}

// DeviceUsernameBindWxRes 绑定微信成功。
type DeviceUsernameBindWxRes struct{}

// DeviceUsernameBindDeviceReq 用户名账号绑定设备号（当前账号来自 Header X-Internal-Wx-Id）。
type DeviceUsernameBindDeviceReq struct {
	g.Meta   `path:"/device/app/api/user/username/bind_device" method:"post" tags:"device" summary:"用户名账号绑定设备号"`
	DeviceNo string `json:"deviceNo" dc:"设备号"`
}

// DeviceUsernameBindDeviceRes 绑定设备成功。
type DeviceUsernameBindDeviceRes struct{}

// DeviceUsernameChangePasswordReq 修改用户名账号密码（当前账号来自 Header X-Internal-Wx-Id）。
type DeviceUsernameChangePasswordReq struct {
	g.Meta      `path:"/device/app/api/user/username/change_password" method:"post" tags:"device" summary:"用户名账号修改密码"`
	OldPassword string `json:"oldPassword" dc:"旧密码"`
	NewPassword string `json:"newPassword" dc:"新密码"`
}

// DeviceUsernameChangePasswordRes 修改密码成功。
type DeviceUsernameChangePasswordRes struct{}

// DeviceWxCreateUsernameReq 微信账号下创建用户名密码（当前账号来自 Header X-Internal-Wx-Id）。
type DeviceWxCreateUsernameReq struct {
	g.Meta   `path:"/device/app/api/user/wx/create_username" method:"post" tags:"device" summary:"微信账号创建用户名密码"`
	Account  string `json:"account" dc:"账号"`
	Password string `json:"password" dc:"密码明文（服务端哈希入库）"`
}

// DeviceWxCreateUsernameRes 微信账号创建用户名成功。
type DeviceWxCreateUsernameRes struct{}

// DeviceWxDeviceLoginReq 设备号业务登录（仅校验设备域已注册，不签发 JWT）。
type DeviceWxDeviceLoginReq struct {
	g.Meta   `path:"/device/app/api/user/device_login" method:"post" tags:"device" summary:"设备号登录业务"`
	DeviceNo string `json:"deviceNo" dc:"设备号，须已在设备域 user 表注册"`
}

// DeviceWxDeviceLoginRes 设备号登录业务响应（无 token）。wxId 为 0 表示当前无 wx 行绑定该设备号；非 0 为 wx 表主键。
type DeviceWxDeviceLoginRes struct {
	WxId      int64  `json:"wxId" dc:"wx 表主键，无绑定时为 0"`
	DeviceNo  string `json:"deviceNo"`
	IsNewUser bool   `json:"isNewUser" dc:"设备号登录恒为 false"`
}

// DeviceWxDetailReq 按 wx 查询已绑定设备号（wx 主键来自 Header X-Internal-Wx-Id）。
type DeviceWxDetailReq struct {
	g.Meta `path:"/device/app/api/user/detail" method:"get" tags:"device" summary:"wx 绑定设备号"`
}

// DeviceWxDetailRes wx 详情。
type DeviceWxDetailRes struct {
	DeviceNo string `json:"deviceNo"`
}

// DeviceUserDeactivateReq 注销账号（按 Header X-Internal-Wx-Id 删除 wx 记录）。
// 注意：该接口仅删除 wx 表当前主键记录，不联动删除其他域数据。
type DeviceUserDeactivateReq struct {
	g.Meta `path:"/device/app/api/user/deactivate" method:"post" tags:"device" summary:"注销账号（删除 wx 记录）"`
}

// DeviceUserDeactivateRes 注销成功。
type DeviceUserDeactivateRes struct{}

// DeviceWxInternalByIDReq 网关内部：按 wx 主键取 unionid。
type DeviceWxInternalByIDReq struct {
	g.Meta `path:"/device/app/api/user/internal/by-id" method:"get" tags:"device" summary:"内部按 id 取 unionid"`
	Id     int64 `json:"id" p:"id" dc:"wx 表主键"`
}

// DeviceWxInternalByIDRes 内部查询响应。
type DeviceWxInternalByIDRes struct {
	UnionId string `json:"unionId"`
}

// DeviceWxInternalDeviceNoByWxIDReq 网关内部：按 wx 主键取 device_no（刷新 access 写 claim 等）。
type DeviceWxInternalDeviceNoByWxIDReq struct {
	g.Meta `path:"/device/app/api/user/internal/device-no-by-wx-id" method:"get" tags:"device" summary:"内部按 wxId 取 device_no"`
	WxId   int64 `json:"wxId" p:"wxId" dc:"wx 表主键"`
}

// DeviceWxInternalDeviceNoByWxIDRes 内部 device_no 响应。
type DeviceWxInternalDeviceNoByWxIDRes struct {
	DeviceNo string `json:"deviceNo"`
}
