// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UcgAiConfig is the golang structure for table ucg_ai_config.
type UcgAiConfig struct {
	Id                  int    `json:"id"                  ` // singleton row id=1
	VisionModel         string `json:"visionModel"         ` //
	MaxImagesPerRequest int    `json:"maxImagesPerRequest" ` //
	UpdatedAt           int64  `json:"updatedAt"           ` // unix seconds
	UpdatedBy           string `json:"updatedBy"           ` // admin operator label
}
