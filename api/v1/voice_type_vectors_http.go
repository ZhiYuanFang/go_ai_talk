package v1

import "github.com/gogf/gf/v2/frame/g"

// VoiceAdminTypeVectorsListReq 列出 Python feeding_type_cues 类型关键字向量。
type VoiceAdminTypeVectorsListReq struct {
	g.Meta `path:"/voice/admin/api/type-vectors" method:"get" tags:"voice-admin" summary:"列出类型向量缓存"`
	Offset int `json:"offset" d:"0" dc:"偏移"`
	Limit  int `json:"limit" d:"100" dc:"每页条数，最大 500"`
}

// VoiceAdminTypeVectorsListRes 列表响应（透传 Python 字段）。
type VoiceAdminTypeVectorsListRes struct {
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Limit  int                      `json:"limit"`
	Items  []map[string]interface{} `json:"items"`
}

// VoiceAdminTypeVectorBulkItem 批量种子条目（payload: target_type/op/history_mode）。
type VoiceAdminTypeVectorBulkItem struct {
	Document string                 `json:"document"`
	Payload  map[string]interface{} `json:"payload"`
}

// VoiceAdminTypeVectorsBulkReq 批量写入类型关键字。
type VoiceAdminTypeVectorsBulkReq struct {
	g.Meta `path:"/voice/admin/api/type-vectors/bulk" method:"post" tags:"voice-admin" summary:"批量写入类型向量"`
	Items  []VoiceAdminTypeVectorBulkItem `json:"items"`
}

// VoiceAdminTypeVectorsBulkRes 批量写入结果（同 document 跳过）。
type VoiceAdminTypeVectorsBulkRes struct {
	Ok      int                      `json:"ok"`
	Added   int                      `json:"added"`
	Skipped []map[string]interface{} `json:"skipped"`
	Failed  []map[string]interface{} `json:"failed"`
	Ids     []string                 `json:"ids"`
}

// VoiceAdminTypeVectorsDedupeReq 整理 document 全等重复。
type VoiceAdminTypeVectorsDedupeReq struct {
	g.Meta `path:"/voice/admin/api/type-vectors/dedupe" method:"post" tags:"voice-admin" summary:"整理类型向量重复文案"`
}

// VoiceAdminTypeVectorsDedupeRes 整理结果（透传 Python）。
type VoiceAdminTypeVectorsDedupeRes struct {
	Kept    int                      `json:"kept"`
	Deleted []map[string]interface{} `json:"deleted"`
	Groups  int                      `json:"groups"`
}

// VoiceAdminTypeVectorsDeleteReq 按向量 id 删除。
type VoiceAdminTypeVectorsDeleteReq struct {
	g.Meta `path:"/voice/admin/api/type-vectors/{id}" method:"delete" tags:"voice-admin" summary:"删除类型向量"`
	Id     string `json:"id" in:"path" v:"required"`
}

// VoiceAdminTypeVectorsDeleteRes 删除结果。
type VoiceAdminTypeVectorsDeleteRes struct {
	Ok bool   `json:"ok"`
	Id string `json:"id"`
}
