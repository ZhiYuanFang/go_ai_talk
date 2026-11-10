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

// AppPushLiveActivityActivityClearPostReq 清除本机 Live Activity activity update token。
// 业务：必须 Bearer（gateway 不得豁免）；身份 MUST 为 X-Internal-Wx-Id；仅删该机 token_kind=activity，不删 push_to_start，不同拍 start。
type AppPushLiveActivityActivityClearPostReq struct {
	g.Meta    `path:"/app/api/push/live-activity/activity/clear" method:"post" tags:"push" summary:"清除本机 Live Activity activity token"`
	DeviceKey string `json:"deviceKey" v:"required|max-length:64" dc:"本机 deviceKey；仅清该机 activity 行"`
}

// AppPushLiveActivityActivityClearPostRes 清除受理（无行亦成功）。
type AppPushLiveActivityActivityClearPostRes struct{}
