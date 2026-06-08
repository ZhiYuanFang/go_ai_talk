package v1

import "github.com/gogf/gf/v2/frame/g"

type UcgAdminAiConfigGetReq struct {
	g.Meta `path:"/ucg/admin/api/ai-config" method:"get" tags:"ucg-admin" summary:"读取 UCG AI 配置"`
}

type UcgAdminAiConfigGetRes struct {
	VisionModel         string   `json:"visionModel"`
	MaxImagesPerRequest int      `json:"maxImagesPerRequest"`
	UpdatedAt           int64    `json:"updatedAt"`
	UpdatedBy           string   `json:"updatedBy"`
	AllowedModels       []string `json:"allowedModels"`
}

type UcgAdminAiConfigPutReq struct {
	g.Meta              `path:"/ucg/admin/api/ai-config" method:"put" tags:"ucg-admin" summary:"更新 UCG AI 配置"`
	VisionModel         string `json:"visionModel" v:"required"`
	MaxImagesPerRequest int    `json:"maxImagesPerRequest" v:"required|between:1,9"`
	UpdatedBy           string `json:"updatedBy"`
}

type UcgAdminAiConfigPutRes struct {
	VisionModel         string `json:"visionModel"`
	MaxImagesPerRequest int    `json:"maxImagesPerRequest"`
	UpdatedAt           int64  `json:"updatedAt"`
	UpdatedBy           string `json:"updatedBy"`
}
