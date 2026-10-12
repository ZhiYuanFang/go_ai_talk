package v1

import "github.com/gogf/gf/v2/frame/g"

// VoiceAdminPredictVectorsListReq 列出 Python feeding_predict_cues R·predict 两维飞轮。
type VoiceAdminPredictVectorsListReq struct {
	g.Meta `path:"/voice/admin/api/predict-vectors" method:"get" tags:"voice-admin" summary:"列出预测飞轮缓存"`
	Offset int `json:"offset" d:"0" dc:"偏移"`
	Limit  int `json:"limit" d:"100" dc:"每页条数，最大 500"`
}

// VoiceAdminPredictVectorsListRes 列表响应（透传 Python 字段）。
type VoiceAdminPredictVectorsListRes struct {
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Limit  int                      `json:"limit"`
	Items  []map[string]interface{} `json:"items"`
}

// VoiceAdminPredictVectorBulkItem 批量种子（payload: predict_channel + time_answer）。
type VoiceAdminPredictVectorBulkItem struct {
	Document string                 `json:"document"`
	Payload  map[string]interface{} `json:"payload"`
}

// VoiceAdminPredictVectorsBulkReq 批量写入预测飞轮。
type VoiceAdminPredictVectorsBulkReq struct {
	g.Meta `path:"/voice/admin/api/predict-vectors/bulk" method:"post" tags:"voice-admin" summary:"批量写入预测飞轮"`
	Items  []VoiceAdminPredictVectorBulkItem `json:"items"`
}

// VoiceAdminPredictVectorsBulkRes 批量写入结果。
type VoiceAdminPredictVectorsBulkRes struct {
	Ok      int                      `json:"ok"`
	Added   int                      `json:"added"`
	Skipped []map[string]interface{} `json:"skipped"`
	Failed  []map[string]interface{} `json:"failed"`
	Ids     []string                 `json:"ids"`
}

// VoiceAdminPredictVectorsDeleteReq 按向量 id 删除。
type VoiceAdminPredictVectorsDeleteReq struct {
	g.Meta `path:"/voice/admin/api/predict-vectors/{id}" method:"delete" tags:"voice-admin" summary:"删除预测飞轮"`
	Id     string `json:"id" in:"path" v:"required"`
}

// VoiceAdminPredictVectorsDeleteRes 删除结果。
type VoiceAdminPredictVectorsDeleteRes struct {
	Ok bool   `json:"ok"`
	Id string `json:"id"`
}
