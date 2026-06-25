package v1

import (
	"hello/internal/services/aimodel"

	"github.com/gogf/gf/v2/frame/g"
)

// SimAdminLLMLanesGetReq 读取 sim LLM lane 配置。
type SimAdminLLMLanesGetReq struct {
	g.Meta `path:"/sim/admin/api/llm-lanes" method:"get" tags:"sim-admin" summary:"读取 sim LLM lane 配置"`
}

// SimAdminLLMLaneItem 单条 sim lane 配置。
type SimAdminLLMLaneItem struct {
	Provider    string `json:"provider"`
	Model       string `json:"model"`
	MaxInFlight int    `json:"maxInFlight"`
	MaxWaiters  int    `json:"maxWaiters"`
	UpdatedAt   int64  `json:"updatedAt"`
	UpdatedBy   string `json:"updatedBy"`
}

// SimAdminLLMLanesGetRes GET 响应。
type SimAdminLLMLanesGetRes struct {
	SimText     SimAdminLLMLaneItem `json:"simText"`
	SimVision   SimAdminLLMLaneItem `json:"simVision"`
	SimImageGen SimAdminLLMLaneItem `json:"simImageGen"`
	SimVideoGen SimAdminLLMLaneItem `json:"simVideoGen"`
	Allowlist   map[string][]string `json:"allowlist"`
}

// SimAdminLLMLanesPutReq 更新 sim LLM lane 配置。
type SimAdminLLMLanesPutReq struct {
	g.Meta      `path:"/sim/admin/api/llm-lanes" method:"put" tags:"sim-admin" summary:"更新 sim LLM lane 配置"`
	SimText     SimAdminLLMLaneItem `json:"simText" v:"required"`
	SimVision   SimAdminLLMLaneItem `json:"simVision" v:"required"`
	SimImageGen SimAdminLLMLaneItem `json:"simImageGen" v:"required"`
	SimVideoGen SimAdminLLMLaneItem `json:"simVideoGen" v:"required"`
	UpdatedBy   string              `json:"updatedBy"`
}

// SimAdminLLMLanesPutRes PUT 响应。
type SimAdminLLMLanesPutRes = SimAdminLLMLanesGetRes

// ToLaneDTO 将 API item 转为 aimodel DTO。
func (item SimAdminLLMLaneItem) ToLaneDTO() aimodel.LaneProfileDTO {
	return aimodel.LaneProfileDTO{
		Provider: item.Provider, Model: item.Model,
		MaxInFlight: item.MaxInFlight, MaxWaiters: item.MaxWaiters,
		UpdatedAt: item.UpdatedAt, UpdatedBy: item.UpdatedBy,
	}
}
