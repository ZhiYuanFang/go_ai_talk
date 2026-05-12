package v1

import "github.com/gogf/gf/v2/frame/g"

// GatewayAppLoginReq App 登录：网关聚合 device 微信登录并签发令牌。
type GatewayAppLoginReq struct {
	g.Meta   `path:"/device/app/api/login" method:"post" tags:"gateway-app" summary:"App 登录"`
	JsCode   string `json:"jsCode"   dc:"微信小程序临时登录凭证（wx.login），服务端换票，禁止持久化"`
	Platform string `json:"platform" dc:"与 device 配置 wechatMp.platforms 下键一致，用于选择 appId/secret"`
}

// GatewayAppLoginRes 登录响应（含 JWT access 与不透明 refresh）。
type GatewayAppLoginRes struct {
	WxId         int64  `json:"wxId"`
	DeviceNo     string `json:"deviceNo"`
	IsNewUser    bool   `json:"isNewUser"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// GatewayAppDeviceLoginReq 设备号聚合登录：转发 device 业务校验后签发与微信登录同形态的令牌。
type GatewayAppDeviceLoginReq struct {
	g.Meta   `path:"/device/app/api/device_login" method:"post" tags:"gateway-app" summary:"设备号聚合登录"`
	DeviceNo string `json:"deviceNo" dc:"须在 device user 表已注册；无 wx 绑定时网关签发 sub=0 的 access"`
}

// GatewayAppDeviceLoginRes 与 GatewayAppLoginRes 字段对齐；wxId 可为 0（纯设备会话）。
type GatewayAppDeviceLoginRes struct {
	WxId         int64  `json:"wxId" dc:"wx 主键，无 wx 绑定时为 0"`
	DeviceNo     string `json:"deviceNo"`
	IsNewUser    bool   `json:"isNewUser"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// GatewayAppTokenRefreshReq 刷新 access。
type GatewayAppTokenRefreshReq struct {
	g.Meta       `path:"/device/app/api/token/refresh" method:"post" tags:"gateway-app" summary:"刷新 access_token"`
	RefreshToken string `json:"refreshToken" dc:"刷新令牌"`
}

// GatewayAppTokenRefreshRes 刷新结果（旋转 refresh 时同时返回新 refresh）。
type GatewayAppTokenRefreshRes struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// GatewayAppVersionCheckReq 版本检查。
type GatewayAppVersionCheckReq struct {
	g.Meta         `path:"/device/app/api/version/check" method:"get" tags:"gateway-app" summary:"App 版本检查"`
	CurrentVersion string `json:"currentVersion" p:"currentVersion" dc:"当前客户端版本号"`
}

// GatewayAppVersionCheckRes 版本检查结果。
type GatewayAppVersionCheckRes struct {
	NeedUpdate    bool   `json:"needUpdate"`
	LatestVersion string `json:"latestVersion"`
	ReleaseDate   int64  `json:"releaseDate" dc:"当前最新版本上线时间，Unix 时间戳（秒）"`
	ReleaseNotes  string `json:"releaseNotes"`
	DownloadUrl   string `json:"downloadUrl"`
	ForceUpdate   bool   `json:"forceUpdate"`
}
