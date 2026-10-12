package v1

import "github.com/gogf/gf/v2/frame/g"

// VoiceAdminReadSourceVectorsListReq 列出 Python feeding_read_source_cues R·数据源飞轮。
type VoiceAdminReadSourceVectorsListReq struct {
	g.Meta `path:"/voice/admin/api/read-source-vectors" method:"get" tags:"voice-admin" summary:"列出R数据源飞轮缓存"`
	Offset int `json:"offset" d:"0" dc:"偏移"`
	Limit  int `json:"limit" d:"100" dc:"每页条数，最大 500"`
}

// VoiceAdminReadSourceVectorsListRes 列表响应（透传 Python 字段）。
type VoiceAdminReadSourceVectorsListRes struct {
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Limit  int                      `json:"limit"`
	Items  []map[string]interface{} `json:"items"`
}

// VoiceAdminReadSourceVectorBulkItem 批量种子（payload: read_source）。
type VoiceAdminReadSourceVectorBulkItem struct {
	Document string                 `json:"document"`
	Payload  map[string]interface{} `json:"payload"`
}

// VoiceAdminReadSourceVectorsBulkReq 批量写入 R·数据源飞轮。
type VoiceAdminReadSourceVectorsBulkReq struct {
	g.Meta `path:"/voice/admin/api/read-source-vectors/bulk" method:"post" tags:"voice-admin" summary:"批量写入R数据源飞轮"`
	Items  []VoiceAdminReadSourceVectorBulkItem `json:"items"`
}

// VoiceAdminReadSourceVectorsBulkRes 批量写入结果。
type VoiceAdminReadSourceVectorsBulkRes struct {
	Ok      int                      `json:"ok"`
	Added   int                      `json:"added"`
	Skipped []map[string]interface{} `json:"skipped"`
	Failed  []map[string]interface{} `json:"failed"`
	Ids     []string                 `json:"ids"`
}

// VoiceAdminReadSourceVectorsDeleteReq 按向量 id 删除。
type VoiceAdminReadSourceVectorsDeleteReq struct {
	g.Meta `path:"/voice/admin/api/read-source-vectors/{id}" method:"delete" tags:"voice-admin" summary:"删除R数据源飞轮"`
	Id     string `json:"id" in:"path" v:"required"`
}

// VoiceAdminReadSourceVectorsDeleteRes 删除结果。
type VoiceAdminReadSourceVectorsDeleteRes struct {
	Ok bool   `json:"ok"`
	Id string `json:"id"`
}
