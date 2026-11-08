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

// AppPushLiveActivityRegisterPostReq 注册 Live Activity Push-to-Start / activity update token（分表，不进 push_device）。
// 业务：可无 Bearer（gateway 精确豁免）；落库身份 MUST 为 body.wxId；统一单活动后不再传 attributesKind/activityId。
type AppPushLiveActivityRegisterPostReq struct {
	g.Meta    `path:"/app/api/push/live-activity/register" method:"post" tags:"push" summary:"注册 Live Activity token"`
	WxId      int64  `json:"wxId" v:"required|min:1" dc:"归属用户 wx 主键；免鉴权路径唯一身份来源"`
	DeviceKey string `json:"deviceKey" v:"required|max-length:64"`
	TokenKind string `json:"tokenKind" v:"required|in:push_to_start,activity"`
	Token     string `json:"token" v:"required|max-length:512"`
}

// AppPushLiveActivityRegisterPostRes 注册成功。
type AppPushLiveActivityRegisterPostRes struct{}
