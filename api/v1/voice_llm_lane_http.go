package v1

import (
	"hello/internal/services/aimodel"

	"github.com/gogf/gf/v2/frame/g"
)

// VoiceAdminLLMLanesGetReq 读取 voice LLM lane 配置。
type VoiceAdminLLMLanesGetReq struct {
	g.Meta `path:"/voice/admin/api/llm-lanes" method:"get" tags:"voice-admin" summary:"读取 LLM lane 配置"`
}

// VoiceAdminLLMLaneItem 单条 lane 配置（clinic/careAlert 可含额度不足 free；voiceUnderstanding 忽略 free 并持久化空）。
type VoiceAdminLLMLaneItem struct {
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	FreeProvider string `json:"freeProvider"`
	FreeModel    string `json:"freeModel"`
	MaxInFlight  int    `json:"maxInFlight"`
	MaxWaiters   int    `json:"maxWaiters"`
	UpdatedAt    int64  `json:"updatedAt"`
	UpdatedBy    string `json:"updatedBy"`
}

// VoiceAdminLLMLanesGetRes GET 响应（含 careAlert）。
type VoiceAdminLLMLanesGetRes struct {
	VoiceUnderstanding VoiceAdminLLMLaneItem `json:"voiceUnderstanding"`
	Clinic             VoiceAdminLLMLaneItem `json:"clinic"`
	CareAlert          VoiceAdminLLMLaneItem `json:"careAlert"`
	Allowlist          map[string][]string   `json:"allowlist"`
}

// VoiceAdminLLMLanesPutReq 更新 voice LLM lane 配置。
type VoiceAdminLLMLanesPutReq struct {
	g.Meta             `path:"/voice/admin/api/llm-lanes" method:"put" tags:"voice-admin" summary:"更新 LLM lane 配置"`
	VoiceUnderstanding VoiceAdminLLMLaneItem `json:"voiceUnderstanding" v:"required"`
	Clinic             VoiceAdminLLMLaneItem `json:"clinic" v:"required"`
	CareAlert          VoiceAdminLLMLaneItem `json:"careAlert" v:"required"`
	UpdatedBy          string                `json:"updatedBy"`
}

// VoiceAdminLLMLanesPutRes PUT 响应。
type VoiceAdminLLMLanesPutRes = VoiceAdminLLMLanesGetRes

// ToLaneDTO 将 API item 转为 aimodel DTO。
func (item VoiceAdminLLMLaneItem) ToLaneDTO() aimodel.LaneProfileDTO {
	return aimodel.LaneProfileDTO{
		Provider:     item.Provider,
		Model:        item.Model,
		FreeProvider: item.FreeProvider,
		FreeModel:    item.FreeModel,
		MaxInFlight:  item.MaxInFlight,
		MaxWaiters:   item.MaxWaiters,
		UpdatedAt:    item.UpdatedAt,
		UpdatedBy:    item.UpdatedBy,
	}
}
