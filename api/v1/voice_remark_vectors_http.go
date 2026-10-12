package v1

import "github.com/gogf/gf/v2/frame/g"

// VoiceAdminRemarkVectorsListReq 列出 Python feeding_remark_cues 备注整段飞轮。
type VoiceAdminRemarkVectorsListReq struct {
	g.Meta `path:"/voice/admin/api/remark-vectors" method:"get" tags:"voice-admin" summary:"列出备注飞轮缓存"`
	Offset int `json:"offset" d:"0" dc:"偏移"`
	Limit  int `json:"limit" d:"100" dc:"每页条数，最大 500"`
}

// VoiceAdminRemarkVectorsListRes 列表响应（透传 Python 字段）。
type VoiceAdminRemarkVectorsListRes struct {
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Limit  int                      `json:"limit"`
	Items  []map[string]interface{} `json:"items"`
}

// VoiceAdminRemarkVectorBulkItem 批量种子（payload: remark，可空）。
type VoiceAdminRemarkVectorBulkItem struct {
	Document string                 `json:"document"`
	Payload  map[string]interface{} `json:"payload"`
}

// VoiceAdminRemarkVectorsBulkReq 批量写入备注飞轮。
type VoiceAdminRemarkVectorsBulkReq struct {
	g.Meta `path:"/voice/admin/api/remark-vectors/bulk" method:"post" tags:"voice-admin" summary:"批量写入备注飞轮"`
	Items  []VoiceAdminRemarkVectorBulkItem `json:"items"`
}

// VoiceAdminRemarkVectorsBulkRes 批量写入结果。
type VoiceAdminRemarkVectorsBulkRes struct {
	Ok      int                      `json:"ok"`
	Added   int                      `json:"added"`
	Skipped []map[string]interface{} `json:"skipped"`
	Failed  []map[string]interface{} `json:"failed"`
	Ids     []string                 `json:"ids"`
}

// VoiceAdminRemarkVectorsDeleteReq 按向量 id 删除。
type VoiceAdminRemarkVectorsDeleteReq struct {
	g.Meta `path:"/voice/admin/api/remark-vectors/{id}" method:"delete" tags:"voice-admin" summary:"删除备注飞轮"`
	Id     string `json:"id" in:"path" v:"required"`
}

// VoiceAdminRemarkVectorsDeleteRes 删除结果。
type VoiceAdminRemarkVectorsDeleteRes struct {
	Ok bool   `json:"ok"`
	Id string `json:"id"`
}
