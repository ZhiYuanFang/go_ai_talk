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

// AppPushLiveActivityRegisterPostReq 注册 Live Activity Push-to-Start / per-activity update token（分表，不进 push_device）。
type AppPushLiveActivityRegisterPostReq struct {
	g.Meta         `path:"/app/api/push/live-activity/register" method:"post" tags:"push" summary:"注册 Live Activity token"`
	DeviceKey      string `json:"deviceKey" v:"required|max-length:64"`
	TokenKind      string `json:"tokenKind" v:"required|in:push_to_start,activity"`
	AttributesKind string `json:"attributesKind" v:"required|in:imminent,timing"`
	ActivityId     string `json:"activityId" v:"max-length:128" dc:"tokenKind=activity 时必填；push_to_start 可空"`
	Token          string `json:"token" v:"required|max-length:512"`
}

// AppPushLiveActivityRegisterPostRes 注册成功。
type AppPushLiveActivityRegisterPostRes struct{}
