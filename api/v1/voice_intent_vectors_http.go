package v1

import "github.com/gogf/gf/v2/frame/g"

// VoiceAdminIntentVectorsListReq 列出 Python feeding_intents。
type VoiceAdminIntentVectorsListReq struct {
	g.Meta `path:"/voice/admin/api/intent-vectors" method:"get" tags:"voice-admin" summary:"列出意图向量缓存"`
	Offset int `json:"offset" d:"0" dc:"偏移"`
	Limit  int `json:"limit" d:"100" dc:"每页条数，最大 500"`
}

// VoiceAdminIntentVectorsListRes 列表响应（透传 Python 字段）。
type VoiceAdminIntentVectorsListRes struct {
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Limit  int                      `json:"limit"`
	Items  []map[string]interface{} `json:"items"`
}

// VoiceAdminIntentVectorBulkItem 批量种子条目。
type VoiceAdminIntentVectorBulkItem struct {
	Document string                 `json:"document"`
	Payload  map[string]interface{} `json:"payload"`
}

// VoiceAdminIntentVectorsBulkReq 批量写入意图种子。
type VoiceAdminIntentVectorsBulkReq struct {
	g.Meta `path:"/voice/admin/api/intent-vectors/bulk" method:"post" tags:"voice-admin" summary:"批量写入意图向量"`
	Items  []VoiceAdminIntentVectorBulkItem `json:"items"`
}

// VoiceAdminIntentVectorsBulkRes 批量写入结果。
type VoiceAdminIntentVectorsBulkRes struct {
	Ok     int                      `json:"ok"`
	Failed []map[string]interface{} `json:"failed"`
	Ids    []string                 `json:"ids"`
}

// VoiceAdminIntentVectorsDeleteReq 按向量 id 删除。
type VoiceAdminIntentVectorsDeleteReq struct {
	g.Meta   `path:"/voice/admin/api/intent-vectors/{id}" method:"delete" tags:"voice-admin" summary:"删除意图向量"`
	Id       string `json:"id" in:"path" v:"required"`
}

// VoiceAdminIntentVectorsDeleteRes 删除结果。
type VoiceAdminIntentVectorsDeleteRes struct {
	Ok bool   `json:"ok"`
	Id string `json:"id"`
}
