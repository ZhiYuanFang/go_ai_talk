package v1

import (
	"hello/internal/model/entity"
	sharedtypes "hello/internal/shared/types"

	"github.com/gogf/gf/v2/frame/g"
)

// --- 管理口令校验（内部） ---

type DeviceInternalVerifyAdminPasswordReq struct {
	g.Meta   `path:"/device/internal/api/admin/verify-password" method:"post" tags:"device" summary:"内部-校验管理口令"`
	Password string `json:"password"`
}

type DeviceInternalVerifyAdminPasswordRes struct {
	OK bool `json:"ok"`
}

// --- 设备注册（内部免管理口令） ---

type DeviceInternalRegisterReq struct {
	g.Meta   `path:"/device/internal/api/register" method:"post" tags:"device" summary:"内部-注册设备"`
	DeviceNo string `json:"deviceNo"`
}

type DeviceInternalRegisterRes struct {
	DeviceNo   string `json:"deviceNo"`
	ActiveTime int64 `json:"activeTime" dc:"激活时间，Unix 秒"`
}

// --- 用户 / 会话 ---

type DeviceInternalUserEnsureReq struct {
	g.Meta   `path:"/device/internal/api/user/ensure" method:"post" tags:"device" summary:"内部-确保设备已注册"`
	DeviceNo string `json:"deviceNo"`
}

type DeviceInternalUserEnsureRes struct{}

type DeviceInternalUserLastTalkReq struct {
	g.Meta   `path:"/device/internal/api/user/last-talk" method:"post" tags:"device" summary:"内部-更新最近对话"`
	DeviceNo string `json:"deviceNo"`
	Ask      string `json:"ask"`
	Answer   string `json:"answer"`
}

type DeviceInternalUserLastTalkRes struct{}

type DeviceInternalUserListReq struct {
	g.Meta `path:"/device/internal/api/user/list" method:"get" tags:"device" summary:"内部-设备列表"`
}

type DeviceInternalUserListRes struct {
	List []entity.User `json:"list"`
}

type DeviceInternalUserListPageReq struct {
	g.Meta   `path:"/device/internal/api/user/list-page" method:"get" tags:"device" summary:"内部-设备分页列表"`
	Page     int    `json:"page" p:"page"`
	PageSize int    `json:"pageSize" p:"pageSize"`
	Q        string `json:"q" p:"q"`
}

type DeviceInternalUserListPageRes struct {
	List     []entity.User `json:"list"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

type DeviceInternalUserTouchAPIAccessReq struct {
	g.Meta   `path:"/device/internal/api/user/touch-api-access" method:"post" tags:"device" summary:"内部-记录最近 HTTP 接口访问"`
	DeviceNo string `json:"deviceNo"`
	ApiPath  string `json:"apiPath" dc:"METHOD /path"`
	At       int64  `json:"at" dc:"Unix 秒，0 表示当前时间"`
}

type DeviceInternalUserTouchAPIAccessRes struct{}

// --- 事件 ---

type DeviceInternalEventAddReq struct {
	g.Meta     `path:"/device/internal/api/event/add" method:"post" tags:"device" summary:"内部-新增事件"`
	Name       string `json:"name"`
	EventType  string `json:"eventType"`
	ExtraNames string `json:"extraNames"`
	Unit       string `json:"unit"`
	ParentId   int64  `json:"parentId"`
}

type DeviceInternalEventAddRes struct{}

type DeviceInternalEventUpdateReq struct {
	g.Meta     `path:"/device/internal/api/event/update" method:"post" tags:"device" summary:"内部-更新事件"`
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	EventType  string `json:"eventType"`
	ExtraNames string `json:"extraNames"`
	Unit       string `json:"unit"`
	// ParentId 非空时修改 parent_id；省略或 null 表示不修改父节点。
	ParentId *int64 `json:"parentId"`
}

type DeviceInternalEventUpdateRes struct{}

type DeviceInternalEventDeleteReq struct {
	g.Meta `path:"/device/internal/api/event/delete" method:"post" tags:"device" summary:"内部-删除事件"`
	Id     int64 `json:"id"`
}

type DeviceInternalEventDeleteRes struct{}

// --- QA / Action ---

type DeviceInternalQAListReq struct {
	g.Meta `path:"/device/internal/api/qa/list" method:"get" tags:"device" summary:"内部-QA列表"`
}

type DeviceInternalQAListRes struct {
	List []entity.Qa `json:"list"`
}

type DeviceInternalActionListReq struct {
	g.Meta `path:"/device/internal/api/action/list" method:"get" tags:"device" summary:"内部-动作列表"`
}

type DeviceInternalActionListRes struct {
	List []sharedtypes.AdminActionItem `json:"list"`
}

type DeviceInternalActionUpdateReq struct {
	g.Meta       `path:"/device/internal/api/action/update" method:"post" tags:"device" summary:"内部-更新动作"`
	Id           int64  `json:"id"`
	Name         string `json:"name"`
	TargetType   string `json:"targetType"`
}

type DeviceInternalActionUpdateRes struct{}

type DeviceInternalActionDeleteReq struct {
	g.Meta `path:"/device/internal/api/action/delete" method:"post" tags:"device" summary:"内部-删除动作"`
	Id     int64 `json:"id"`
}

type DeviceInternalActionDeleteRes struct{}

// --- 语音域写路径 ---

type DeviceInternalVoiceActionReq struct {
	g.Meta       `path:"/device/internal/api/voice/action" method:"post" tags:"device" summary:"内部-语音写动作"`
	Name         string `json:"name"`
	TargetType   string `json:"targetType"`
}

type DeviceInternalVoiceActionRes struct{}

type DeviceInternalVoiceEventNeedleReq struct {
	g.Meta    `path:"/device/internal/api/voice/event/needle" method:"post" tags:"device" summary:"内部-语音按名插入事件"`
	Needle    string `json:"needle"`
	EventType string `json:"eventType"`
	Unit      string `json:"unit"`
}

type DeviceInternalVoiceEventNeedleRes struct {
	Item entity.Event `json:"item"`
}

type DeviceInternalVoiceEventDeepSeekReq struct {
	g.Meta     `path:"/device/internal/api/voice/event/deepseek" method:"post" tags:"device" summary:"内部-DeepSeek事件落库"`
	Name       string `json:"name"`
	ExtraNames string `json:"extraNames"`
	EventType  string `json:"eventType"`
	Unit       string `json:"unit"`
}

type DeviceInternalVoiceEventDeepSeekRes struct {
	Item       entity.Event `json:"item"`
	TargetName string       `json:"targetName"`
}
