package v1

import "github.com/gogf/gf/v2/frame/g"

// DeviceInternalAIQuotaCheckReq 内部预检 AI 月度额度。
type DeviceInternalAIQuotaCheckReq struct {
	g.Meta  `path:"/device/internal/api/ai-quota/check" method:"post" tags:"device" summary:"内部-AI 额度预检"`
	WxId    int64  `json:"wxId" v:"required|min:1"`
	Feature string `json:"feature" v:"required|in:polish,voice_ai"`
}

// DeviceInternalAIQuotaCheckRes 预检结果。
type DeviceInternalAIQuotaCheckRes struct {
	Allowed bool `json:"allowed"`
	Used    int  `json:"used"`
	Limit   int  `json:"limit"`
}

// DeviceInternalAIQuotaConsumeReq 内部扣减 AI 月度额度。
type DeviceInternalAIQuotaConsumeReq struct {
	g.Meta  `path:"/device/internal/api/ai-quota/consume" method:"post" tags:"device" summary:"内部-AI 额度扣减"`
	WxId    int64  `json:"wxId" v:"required|min:1"`
	Feature string `json:"feature" v:"required|in:polish,voice_ai"`
}

// DeviceInternalAIQuotaConsumeRes 扣减后快照。
type DeviceInternalAIQuotaConsumeRes struct {
	Used  int `json:"used"`
	Limit int `json:"limit"`
}

// DeviceInternalAIQuotaDefaultGetReq 内部读取全局默认额度。
type DeviceInternalAIQuotaDefaultGetReq struct {
	g.Meta `path:"/device/internal/api/ai-quota/default" method:"get" tags:"device" summary:"内部-读取 AI 额度全局默认"`
}

// DeviceInternalAIQuotaDefaultGetRes 全局默认。
type DeviceInternalAIQuotaDefaultGetRes struct {
	PolishMonthlyLimit  int   `json:"polishMonthlyLimit"`
	VoiceAiMonthlyLimit int   `json:"voiceAiMonthlyLimit"`
	UpdatedAt           int64 `json:"updatedAt"`
}

// DeviceInternalAIQuotaDefaultPutReq 内部更新全局默认额度。
type DeviceInternalAIQuotaDefaultPutReq struct {
	g.Meta              `path:"/device/internal/api/ai-quota/default" method:"put" tags:"device" summary:"内部-更新 AI 额度全局默认"`
	PolishMonthlyLimit  int `json:"polishMonthlyLimit" v:"required|min:1"`
	VoiceAiMonthlyLimit int `json:"voiceAiMonthlyLimit" v:"required|min:1"`
}

// DeviceInternalAIQuotaDefaultPutRes 更新后全局默认。
type DeviceInternalAIQuotaDefaultPutRes = DeviceInternalAIQuotaDefaultGetRes

// DeviceInternalAIQuotaUserGetReq 内部读取 wxId override。
type DeviceInternalAIQuotaUserGetReq struct {
	g.Meta `path:"/device/internal/api/ai-quota/user" method:"get" tags:"device" summary:"内部-读取 wxId AI 额度 override"`
	WxId   int64 `json:"wxId" p:"wxId" v:"required|min:1"`
}

// DeviceInternalAIQuotaUserGetRes wxId override。
type DeviceInternalAIQuotaUserGetRes struct {
	WxId                int64 `json:"wxId"`
	PolishMonthlyLimit  *int  `json:"polishMonthlyLimit,omitempty"`
	VoiceAiMonthlyLimit *int  `json:"voiceAiMonthlyLimit,omitempty"`
	UpdatedAt           int64 `json:"updatedAt"`
}

// DeviceInternalAIQuotaUserPutReq 内部更新 wxId override；字段省略表示不修改，显式 null 表示清除。
type DeviceInternalAIQuotaUserPutReq struct {
	g.Meta              `path:"/device/internal/api/ai-quota/user" method:"put" tags:"device" summary:"内部-更新 wxId AI 额度 override"`
	WxId                int64 `json:"wxId" v:"required|min:1"`
	PolishMonthlyLimit  *int  `json:"polishMonthlyLimit"`
	VoiceAiMonthlyLimit *int  `json:"voiceAiMonthlyLimit"`
	ClearPolish         bool  `json:"clearPolish"`
	ClearVoiceAi        bool  `json:"clearVoiceAi"`
}

// DeviceInternalAIQuotaUserPutRes 更新后 override。
type DeviceInternalAIQuotaUserPutRes = DeviceInternalAIQuotaUserGetRes

// DeviceInternalWxIdByDeviceNoReq 内部按 deviceNo 反查 wxId。
type DeviceInternalWxIdByDeviceNoReq struct {
	g.Meta   `path:"/device/internal/api/ai-quota/wx-id-by-device-no" method:"get" tags:"device" summary:"内部-按 deviceNo 查 wxId"`
	DeviceNo string `json:"deviceNo" p:"deviceNo" v:"required"`
}

// DeviceInternalWxIdByDeviceNoRes wxId 查询结果。
type DeviceInternalWxIdByDeviceNoRes struct {
	WxId int64 `json:"wxId"`
}

// DeviceAppAIQuotaGetReq App 读取当前用户 AI 额度。
type DeviceAppAIQuotaGetReq struct {
	g.Meta `path:"/device/app/api/ai-quota" method:"get" tags:"app" summary:"当前用户 AI 月度额度"`
}

// DeviceAppAIQuotaFeatureStatus 单 feature 用量。
type DeviceAppAIQuotaFeatureStatus struct {
	Used  int `json:"used"`
	Limit int `json:"limit"`
}

// DeviceAppAIQuotaGetRes App 额度响应。
type DeviceAppAIQuotaGetRes struct {
	Polish  DeviceAppAIQuotaFeatureStatus `json:"polish"`
	VoiceAi DeviceAppAIQuotaFeatureStatus `json:"voiceAi"`
}
