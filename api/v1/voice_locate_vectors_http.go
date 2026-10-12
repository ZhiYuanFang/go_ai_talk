package v1

import "github.com/gogf/gf/v2/frame/g"

// VoiceAdminLocateVectorsListReq 列出 Python feeding_locate_cues U/D 定位策略飞轮。
type VoiceAdminLocateVectorsListReq struct {
	g.Meta `path:"/voice/admin/api/locate-vectors" method:"get" tags:"voice-admin" summary:"列定位飞轮缓存"`
	Offset int `json:"offset" d:"0" dc:"偏移"`
	Limit  int `json:"limit" d:"100" dc:"每页条数，最大 500"`
}

// VoiceAdminLocateVectorsListRes 列表响应（透传 Python 字段）。
type VoiceAdminLocateVectorsListRes struct {
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Limit  int                      `json:"limit"`
	Items  []map[string]interface{} `json:"items"`
}

// VoiceAdminLocateVectorBulkItem 批量种子（payload: strategy；禁 historyId）。
type VoiceAdminLocateVectorBulkItem struct {
	Document string                 `json:"document"`
	Payload  map[string]interface{} `json:"payload"`
}

// VoiceAdminLocateVectorsBulkReq 批量写入定位飞轮。
type VoiceAdminLocateVectorsBulkReq struct {
	g.Meta `path:"/voice/admin/api/locate-vectors/bulk" method:"post" tags:"voice-admin" summary:"批量写入定位飞轮"`
	Items  []VoiceAdminLocateVectorBulkItem `json:"items"`
}

// VoiceAdminLocateVectorsBulkRes 批量写入结果。
type VoiceAdminLocateVectorsBulkRes struct {
	Ok      int                      `json:"ok"`
	Added   int                      `json:"added"`
	Skipped []map[string]interface{} `json:"skipped"`
	Failed  []map[string]interface{} `json:"failed"`
	Ids     []string                 `json:"ids"`
}

// VoiceAdminLocateVectorsDeleteReq 按向量 id 删除。
type VoiceAdminLocateVectorsDeleteReq struct {
	g.Meta `path:"/voice/admin/api/locate-vectors/{id}" method:"delete" tags:"voice-admin" summary:"删除定位飞轮"`
	Id     string `json:"id" in:"path" v:"required"`
}

// VoiceAdminLocateVectorsDeleteRes 删除结果。
type VoiceAdminLocateVectorsDeleteRes struct {
	Ok bool   `json:"ok"`
	Id string `json:"id"`
}
