package v1

import "github.com/gogf/gf/v2/frame/g"

// VoiceAdminTimeVectorsListReq 列出 Python feeding_time_cues 时间相对配方飞轮。
type VoiceAdminTimeVectorsListReq struct {
	g.Meta `path:"/voice/admin/api/time-vectors" method:"get" tags:"voice-admin" summary:"列出时间飞轮缓存"`
	Offset int `json:"offset" d:"0" dc:"偏移"`
	Limit  int `json:"limit" d:"100" dc:"每页条数，最大 500"`
}

// VoiceAdminTimeVectorsListRes 列表响应（透传 Python 字段）。
type VoiceAdminTimeVectorsListRes struct {
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Limit  int                      `json:"limit"`
	Items  []map[string]interface{} `json:"items"`
}

// VoiceAdminTimeVectorBulkItem 批量种子（payload: 相对配方）。
type VoiceAdminTimeVectorBulkItem struct {
	Document string                 `json:"document"`
	Payload  map[string]interface{} `json:"payload"`
}

// VoiceAdminTimeVectorsBulkReq 批量写入时间飞轮。
type VoiceAdminTimeVectorsBulkReq struct {
	g.Meta `path:"/voice/admin/api/time-vectors/bulk" method:"post" tags:"voice-admin" summary:"批量写入时间飞轮"`
	Items  []VoiceAdminTimeVectorBulkItem `json:"items"`
}

// VoiceAdminTimeVectorsBulkRes 批量写入结果。
type VoiceAdminTimeVectorsBulkRes struct {
	Ok      int                      `json:"ok"`
	Added   int                      `json:"added"`
	Skipped []map[string]interface{} `json:"skipped"`
	Failed  []map[string]interface{} `json:"failed"`
	Ids     []string                 `json:"ids"`
}

// VoiceAdminTimeVectorsDeleteReq 按向量 id 删除。
type VoiceAdminTimeVectorsDeleteReq struct {
	g.Meta `path:"/voice/admin/api/time-vectors/{id}" method:"delete" tags:"voice-admin" summary:"删除时间飞轮"`
	Id     string `json:"id" in:"path" v:"required"`
}

// VoiceAdminTimeVectorsDeleteRes 删除结果。
type VoiceAdminTimeVectorsDeleteRes struct {
	Ok bool   `json:"ok"`
	Id string `json:"id"`
}
