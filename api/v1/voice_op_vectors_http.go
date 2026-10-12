package v1

import "github.com/gogf/gf/v2/frame/g"

// VoiceAdminOpVectorsListReq 列出 Python feeding_op_cues OP 轨整段飞轮。
type VoiceAdminOpVectorsListReq struct {
	g.Meta `path:"/voice/admin/api/op-vectors" method:"get" tags:"voice-admin" summary:"列出OP轨飞轮缓存"`
	Offset int `json:"offset" d:"0" dc:"偏移"`
	Limit  int `json:"limit" d:"100" dc:"每页条数，最大 500"`
}

// VoiceAdminOpVectorsListRes 列表响应（透传 Python 字段）。
type VoiceAdminOpVectorsListRes struct {
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Limit  int                      `json:"limit"`
	Items  []map[string]interface{} `json:"items"`
}

// VoiceAdminOpVectorBulkItem 批量种子（payload: op_rail）。
type VoiceAdminOpVectorBulkItem struct {
	Document string                 `json:"document"`
	Payload  map[string]interface{} `json:"payload"`
}

// VoiceAdminOpVectorsBulkReq 批量写入 OP 轨飞轮。
type VoiceAdminOpVectorsBulkReq struct {
	g.Meta `path:"/voice/admin/api/op-vectors/bulk" method:"post" tags:"voice-admin" summary:"批量写入OP轨飞轮"`
	Items  []VoiceAdminOpVectorBulkItem `json:"items"`
}

// VoiceAdminOpVectorsBulkRes 批量写入结果。
type VoiceAdminOpVectorsBulkRes struct {
	Ok      int                      `json:"ok"`
	Added   int                      `json:"added"`
	Skipped []map[string]interface{} `json:"skipped"`
	Failed  []map[string]interface{} `json:"failed"`
	Ids     []string                 `json:"ids"`
}

// VoiceAdminOpVectorsDeleteReq 按向量 id 删除。
type VoiceAdminOpVectorsDeleteReq struct {
	g.Meta `path:"/voice/admin/api/op-vectors/{id}" method:"delete" tags:"voice-admin" summary:"删除OP轨飞轮"`
	Id     string `json:"id" in:"path" v:"required"`
}

// VoiceAdminOpVectorsDeleteRes 删除结果。
type VoiceAdminOpVectorsDeleteRes struct {
	Ok bool   `json:"ok"`
	Id string `json:"id"`
}
