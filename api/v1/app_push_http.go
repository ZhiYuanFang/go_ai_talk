package v1

import "github.com/gogf/gf/v2/frame/g"

// AppPushRegisterPostReq 注册推送设备 token（全局能力，经 gateway /app/api/push/*）。
type AppPushRegisterPostReq struct {
	g.Meta    `path:"/app/api/push/register" method:"post" tags:"push" summary:"注册推送设备 token"`
	Channel   string `json:"channel" v:"required|in:apns,hms,mipush"`
	Token     string `json:"token" v:"required|max-length:512"`
	DeviceKey string `json:"deviceKey" v:"required|max-length:64"`
}

// AppPushRegisterPostRes 注册成功。
type AppPushRegisterPostRes struct{}

// AppPushUnregisterPostReq 注销推送设备 token。
type AppPushUnregisterPostReq struct {
	g.Meta    `path:"/app/api/push/unregister" method:"post" tags:"push" summary:"注销推送设备 token"`
	DeviceKey string `json:"deviceKey" v:"required|max-length:64"`
	Channel   string `json:"channel" v:"in:apns,hms,mipush"`
}

// AppPushUnregisterPostRes 注销成功。
type AppPushUnregisterPostRes struct{}
