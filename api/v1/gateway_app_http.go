package v1

import "github.com/gogf/gf/v2/frame/g"

// GatewayAppLoginReq App 登录：网关聚合 device 微信登录并签发令牌。
type GatewayAppLoginReq struct {
	g.Meta   `path:"/device/app/api/login" method:"post" tags:"gateway-app" summary:"App 登录"`
	WxCode   string `json:"wxCode"   dc:"微信侧 code"`
	Platform string `json:"platform" dc:"平台"`
}

// GatewayAppLoginRes 登录响应（含 JWT access 与不透明 refresh）。
type GatewayAppLoginRes struct {
	WxId         int64  `json:"wxId"`
	WxCode       string `json:"wxCode"`
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
