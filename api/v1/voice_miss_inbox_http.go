package v1

import "github.com/gogf/gf/v2/frame/g"

// VoiceAdminMissInboxListReq 列出 Python 覆盖缺口（答不出）收件箱。
type VoiceAdminMissInboxListReq struct {
	g.Meta   `path:"/voice/admin/api/miss-inbox" method:"get" tags:"voice-admin" summary:"列出覆盖缺口收件箱"`
	Offset   int    `json:"offset" in:"query" d:"0" dc:"偏移"`
	Limit    int    `json:"limit" in:"query" d:"100" dc:"每页条数"`
	Status   string `json:"status" in:"query" dc:"状态；未传默认 open；显式空串=不过滤"`
	FailKind string `json:"fail_kind" in:"query" dc:"失败类别；空=不过滤"`
}

// VoiceAdminMissInboxListRes 列表响应（items 透传 Python 字段）。
type VoiceAdminMissInboxListRes struct {
	Total  int                      `json:"total"`
	Offset int                      `json:"offset"`
	Limit  int                      `json:"limit"`
	Items  []map[string]interface{} `json:"items"`
}

// VoiceAdminMissInboxDeleteReq 按 id 物理删除一条覆盖缺口。
type VoiceAdminMissInboxDeleteReq struct {
	g.Meta `path:"/voice/admin/api/miss-inbox/{id}" method:"delete" tags:"voice-admin" summary:"删除覆盖缺口"`
	Id     string `json:"id" in:"path" v:"required"`
}

// VoiceAdminMissInboxDeleteRes 删除结果。
type VoiceAdminMissInboxDeleteRes struct {
	Ok bool   `json:"ok"`
	Id string `json:"id"`
}
