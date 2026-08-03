package v1

import "github.com/gogf/gf/v2/frame/g"

// VoiceInternalAIQuotaCheckReq 内部预检 voice_ai / clinic_ai 额度。
type VoiceInternalAIQuotaCheckReq struct {
	g.Meta  `path:"/voice/internal/api/ai-quota/check" method:"post" tags:"voice" summary:"内部-AI 额度预检"`
	WxId    int64  `json:"wxId" v:"required|min:1"`
	Feature string `json:"feature" v:"required|in:voice_ai,clinic_ai"`
}

// VoiceInternalAIQuotaCheckRes 预检结果（含 degraded，供跨进程对齐降速语义）。
type VoiceInternalAIQuotaCheckRes struct {
	Allowed  bool `json:"allowed"`
	Used     int  `json:"used"`
	Limit    int  `json:"limit"`
	Degraded bool `json:"degraded"`
}

// VoiceInternalAIQuotaConsumeReq 内部扣减 voice_ai / clinic_ai 额度。
type VoiceInternalAIQuotaConsumeReq struct {
	g.Meta  `path:"/voice/internal/api/ai-quota/consume" method:"post" tags:"voice" summary:"内部-AI 额度扣减"`
	WxId    int64  `json:"wxId" v:"required|min:1"`
	Feature string `json:"feature" v:"required|in:voice_ai,clinic_ai"`
}

// VoiceInternalAIQuotaConsumeRes 扣减后快照。
type VoiceInternalAIQuotaConsumeRes struct {
	Used  int `json:"used"`
	Limit int `json:"limit"`
}

// VoiceAppAIQuotaGetReq App 读取 voice 域 AI 额度。
type VoiceAppAIQuotaGetReq struct {
	g.Meta `path:"/voice/app/api/ai-quota" method:"get" tags:"app" summary:"voice 域 AI 月度额度"`
}

// VoiceAppAIQuotaFeatureStatus 单 feature 用量。
type VoiceAppAIQuotaFeatureStatus struct {
	Used     int  `json:"used"`
	Limit    int  `json:"limit"`
	Degraded bool `json:"degraded"`
}

// VoiceAppAIQuotaGetRes App 额度响应（voiceAi + clinicAi）。
type VoiceAppAIQuotaGetRes struct {
	VoiceAi  VoiceAppAIQuotaFeatureStatus `json:"voiceAi"`
	ClinicAi VoiceAppAIQuotaFeatureStatus `json:"clinicAi"`
}

// VoiceAdminAIQuotaDefaultGetReq 读取 voice 域全局默认额度。
type VoiceAdminAIQuotaDefaultGetReq struct {
	g.Meta `path:"/voice/admin/api/ai-quota/default" method:"get" tags:"voice-admin" summary:"读取 voice AI 额度全局默认"`
}

// VoiceAdminAIQuotaDefaultGetRes 全局默认。
type VoiceAdminAIQuotaDefaultGetRes struct {
	VoiceAiMonthlyLimit  int   `json:"voiceAiMonthlyLimit"`
	ClinicAiMonthlyLimit int   `json:"clinicAiMonthlyLimit"`
	UpdatedAt            int64 `json:"updatedAt"`
}

// VoiceAdminAIQuotaDefaultPutReq 更新 voice 域全局默认额度。
type VoiceAdminAIQuotaDefaultPutReq struct {
	g.Meta               `path:"/voice/admin/api/ai-quota/default" method:"put" tags:"voice-admin" summary:"更新 voice AI 额度全局默认"`
	VoiceAiMonthlyLimit  int `json:"voiceAiMonthlyLimit" v:"required|min:1"`
	ClinicAiMonthlyLimit int `json:"clinicAiMonthlyLimit" v:"required|min:1"`
}

// VoiceAdminAIQuotaDefaultPutRes 更新后全局默认。
type VoiceAdminAIQuotaDefaultPutRes = VoiceAdminAIQuotaDefaultGetRes

// VoiceAdminAIQuotaUserGetReq 读取 wxId override。
type VoiceAdminAIQuotaUserGetReq struct {
	g.Meta `path:"/voice/admin/api/ai-quota/user" method:"get" tags:"voice-admin" summary:"读取 wxId voice AI 额度 override"`
	WxId   int64 `json:"wxId" p:"wxId" v:"required|min:1"`
}

// VoiceAdminAIQuotaUserGetRes wxId override。
type VoiceAdminAIQuotaUserGetRes struct {
	WxId                 int64 `json:"wxId"`
	VoiceAiMonthlyLimit  *int  `json:"voiceAiMonthlyLimit,omitempty"`
	ClinicAiMonthlyLimit *int  `json:"clinicAiMonthlyLimit,omitempty"`
	UpdatedAt            int64 `json:"updatedAt"`
}

// VoiceAdminAIQuotaUserPutReq 更新 wxId override。
type VoiceAdminAIQuotaUserPutReq struct {
	g.Meta               `path:"/voice/admin/api/ai-quota/user" method:"put" tags:"voice-admin" summary:"更新 wxId voice AI 额度 override"`
	WxId                 int64 `json:"wxId" v:"required|min:1"`
	VoiceAiMonthlyLimit  *int  `json:"voiceAiMonthlyLimit"`
	ClinicAiMonthlyLimit *int  `json:"clinicAiMonthlyLimit"`
	ClearVoiceAi         bool  `json:"clearVoiceAi"`
	ClearClinicAi        bool  `json:"clearClinicAi"`
}

// VoiceAdminAIQuotaUserPutRes 更新后 override。
type VoiceAdminAIQuotaUserPutRes = VoiceAdminAIQuotaUserGetRes

// VoiceAdminAIQuotaUsersGetReq 分页列出全部真实 wx 的有效额度与身份。
type VoiceAdminAIQuotaUsersGetReq struct {
	g.Meta   `path:"/voice/admin/api/ai-quota/users" method:"get" tags:"voice-admin" summary:"分页列出用户 voice/clinic AI 额度"`
	Page     int    `json:"page" p:"page" dc:"页码，从 1 开始"`
	PageSize int    `json:"pageSize" p:"pageSize" dc:"每页条数，默认 20，最大 100"`
	DeviceNo string `json:"deviceNo" p:"deviceNo" dc:"按设备号过滤（模糊/前缀，经 device wx 列表 q）"`
}

// VoiceAdminAIQuotaUsersFeature 列表行内单 feature 已用/上限。
type VoiceAdminAIQuotaUsersFeature struct {
	Used  int `json:"used"`
	Limit int `json:"limit"`
}

// VoiceAdminAIQuotaUsersItem 用户额度列表行。
type VoiceAdminAIQuotaUsersItem struct {
	DeviceNo string                          `json:"deviceNo"`
	WxId     int64                           `json:"wxId"`
	Account  string                          `json:"account"`
	BabyName string                          `json:"babyName"`
	VoiceAi  VoiceAdminAIQuotaUsersFeature   `json:"voiceAi"`
	ClinicAi VoiceAdminAIQuotaUsersFeature   `json:"clinicAi"`
}

// VoiceAdminAIQuotaUsersGetRes 用户额度分页响应。
type VoiceAdminAIQuotaUsersGetRes struct {
	List     []VoiceAdminAIQuotaUsersItem `json:"list"`
	Total    int                          `json:"total"`
	Page     int                          `json:"page"`
	PageSize int                          `json:"pageSize"`
}
