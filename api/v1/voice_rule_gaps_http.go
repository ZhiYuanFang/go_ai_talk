package v1

import "github.com/gogf/gf/v2/frame/g"

// VoiceAdminRuleGapsListReq 列出 Python 规则缺口收件箱。
type VoiceAdminRuleGapsListReq struct {
	g.Meta    `path:"/voice/admin/api/rule-gaps" method:"get" tags:"voice-admin" summary:"列出规则缺口收件箱"`
	Offset    int    `json:"offset" in:"query" d:"0" dc:"偏移"`
	Limit     int    `json:"limit" in:"query" d:"100" dc:"每页条数"`
	Dimension string `json:"dimension" in:"query" dc:"维度筛选；空=全部"`
	Status string `json:"status" in:"query" dc:"状态；未传默认 open；显式空串=不过滤"`
}

// VoiceAdminRuleGapsListRes 列表响应（items 透传 Python snake_case 字段）。
type VoiceAdminRuleGapsListRes struct {
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Limit  int                      `json:"limit"`
	Items  []map[string]interface{} `json:"items"`
}

// VoiceAdminRuleGapsPatchReq 更新缺口状态（如 dismissed）。
type VoiceAdminRuleGapsPatchReq struct {
	g.Meta `path:"/voice/admin/api/rule-gaps/{id}" method:"patch" tags:"voice-admin" summary:"更新规则缺口状态"`
	Id     string `json:"id" in:"path" v:"required"`
	Status string `json:"status" v:"required" dc:"如 dismissed / open"`
}

// VoiceAdminRuleGapsPatchRes PATCH 结果。
type VoiceAdminRuleGapsPatchRes struct {
	Ok     bool   `json:"ok"`
	Id     string `json:"id"`
	Status string `json:"status"`
}
