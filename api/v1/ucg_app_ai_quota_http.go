package v1

import "github.com/gogf/gf/v2/frame/g"

// UcgAppAIQuotaGetReq App 读取 ucg 域润笔额度。
type UcgAppAIQuotaGetReq struct {
	g.Meta `path:"/ucg/app/api/ai-quota" method:"get" tags:"ucg" summary:"润笔 AI 月度额度"`
}

// UcgAppAIQuotaFeatureStatus 单 feature 用量。
type UcgAppAIQuotaFeatureStatus struct {
	Used  int `json:"used"`
	Limit int `json:"limit"`
}

// UcgAppAIQuotaGetRes App 额度响应（polish only）。
type UcgAppAIQuotaGetRes struct {
	Polish UcgAppAIQuotaFeatureStatus `json:"polish"`
}
