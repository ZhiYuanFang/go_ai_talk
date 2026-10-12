package v1

import "github.com/gogf/gf/v2/frame/g"

// VoiceAdminQtyVectorsListReq 列出 Python feeding_qty_cues 数量整段飞轮。
type VoiceAdminQtyVectorsListReq struct {
	g.Meta `path:"/voice/admin/api/qty-vectors" method:"get" tags:"voice-admin" summary:"列出数量飞轮缓存"`
	Offset int `json:"offset" d:"0" dc:"偏移"`
	Limit  int `json:"limit" d:"100" dc:"每页条数，最大 500"`
}

// VoiceAdminQtyVectorsListRes 列表响应（透传 Python 字段）。
type VoiceAdminQtyVectorsListRes struct {
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Limit  int                      `json:"limit"`
	Items  []map[string]interface{} `json:"items"`
}

// VoiceAdminQtyVectorBulkItem 批量种子（payload: quantity）。
type VoiceAdminQtyVectorBulkItem struct {
	Document string                 `json:"document"`
	Payload  map[string]interface{} `json:"payload"`
}

// VoiceAdminQtyVectorsBulkReq 批量写入数量飞轮。
type VoiceAdminQtyVectorsBulkReq struct {
	g.Meta `path:"/voice/admin/api/qty-vectors/bulk" method:"post" tags:"voice-admin" summary:"批量写入数量飞轮"`
	Items  []VoiceAdminQtyVectorBulkItem `json:"items"`
}

// VoiceAdminQtyVectorsBulkRes 批量写入结果。
type VoiceAdminQtyVectorsBulkRes struct {
	Ok      int                      `json:"ok"`
	Added   int                      `json:"added"`
	Skipped []map[string]interface{} `json:"skipped"`
	Failed  []map[string]interface{} `json:"failed"`
	Ids     []string                 `json:"ids"`
}

// VoiceAdminQtyVectorsDeleteReq 按向量 id 删除。
type VoiceAdminQtyVectorsDeleteReq struct {
	g.Meta `path:"/voice/admin/api/qty-vectors/{id}" method:"delete" tags:"voice-admin" summary:"删除数量飞轮"`
	Id     string `json:"id" in:"path" v:"required"`
}

// VoiceAdminQtyVectorsDeleteRes 删除结果。
type VoiceAdminQtyVectorsDeleteRes struct {
	Ok bool   `json:"ok"`
	Id string `json:"id"`
}
