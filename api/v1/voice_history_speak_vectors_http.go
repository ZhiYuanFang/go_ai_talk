package v1

import "github.com/gogf/gf/v2/frame/g"

// VoiceAdminHistorySpeakVectorsListReq 列出 Python feeding_history_speak_cues 查史播报策略向量。
type VoiceAdminHistorySpeakVectorsListReq struct {
	g.Meta `path:"/voice/admin/api/history-speak-vectors" method:"get" tags:"voice-admin" summary:"列出查史播报向量缓存"`
	Offset int `json:"offset" d:"0" dc:"偏移"`
	Limit  int `json:"limit" d:"100" dc:"每页条数，最大 500"`
}

// VoiceAdminHistorySpeakVectorsListRes 列表响应（透传 Python 字段）。
type VoiceAdminHistorySpeakVectorsListRes struct {
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Limit  int                      `json:"limit"`
	Items  []map[string]interface{} `json:"items"`
}

// VoiceAdminHistorySpeakVectorBulkItem 批量种子（payload: point_pick/open_speak）。
type VoiceAdminHistorySpeakVectorBulkItem struct {
	Document string                 `json:"document"`
	Payload  map[string]interface{} `json:"payload"`
}

// VoiceAdminHistorySpeakVectorsBulkReq 批量写入查史播报策略关键字。
type VoiceAdminHistorySpeakVectorsBulkReq struct {
	g.Meta `path:"/voice/admin/api/history-speak-vectors/bulk" method:"post" tags:"voice-admin" summary:"批量写入查史播报向量"`
	Items  []VoiceAdminHistorySpeakVectorBulkItem `json:"items"`
}

// VoiceAdminHistorySpeakVectorsBulkRes 批量写入结果（同 document 跳过）。
type VoiceAdminHistorySpeakVectorsBulkRes struct {
	Ok      int                      `json:"ok"`
	Added   int                      `json:"added"`
	Skipped []map[string]interface{} `json:"skipped"`
	Failed  []map[string]interface{} `json:"failed"`
	Ids     []string                 `json:"ids"`
}

// VoiceAdminHistorySpeakVectorsDedupeReq 整理 document 全等重复。
type VoiceAdminHistorySpeakVectorsDedupeReq struct {
	g.Meta `path:"/voice/admin/api/history-speak-vectors/dedupe" method:"post" tags:"voice-admin" summary:"整理查史播报向量重复文案"`
}

// VoiceAdminHistorySpeakVectorsDedupeRes 整理结果（透传 Python）。
type VoiceAdminHistorySpeakVectorsDedupeRes struct {
	Kept    int                      `json:"kept"`
	Deleted []map[string]interface{} `json:"deleted"`
	Groups  int                      `json:"groups"`
}

// VoiceAdminHistorySpeakVectorsDeleteReq 按向量 id 删除。
type VoiceAdminHistorySpeakVectorsDeleteReq struct {
	g.Meta `path:"/voice/admin/api/history-speak-vectors/{id}" method:"delete" tags:"voice-admin" summary:"删除查史播报向量"`
	Id     string `json:"id" in:"path" v:"required"`
}

// VoiceAdminHistorySpeakVectorsDeleteRes 删除结果。
type VoiceAdminHistorySpeakVectorsDeleteRes struct {
	Ok bool   `json:"ok"`
	Id string `json:"id"`
}
