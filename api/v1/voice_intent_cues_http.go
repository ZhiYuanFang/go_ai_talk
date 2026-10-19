package v1

import "github.com/gogf/gf/v2/frame/g"

// VoiceAdminIntentCuesByDocumentReq 按整句原文精确查询九个喂养飞轮。
//
// 业务说明：Hub 意图智能体页一次看完命中；删除仍走各维度已有 DELETE，本接口只读。
type VoiceAdminIntentCuesByDocumentReq struct {
	g.Meta   `path:"/voice/admin/api/intent-cues/by-document" method:"get" tags:"voice-admin" summary:"按原文查询九个飞轮"`
	Document string `json:"document" dc:"trim 后精确匹配的整句原文"`
}

// VoiceAdminIntentCueHit 一条飞轮命中。dimension 与门户删除路径一一对应。
type VoiceAdminIntentCueHit struct {
	Dimension  string                 `json:"dimension"`
	Collection string                 `json:"collection"`
	Id         string                 `json:"id"`
	Document   string                 `json:"document"`
	Payload    map[string]interface{} `json:"payload"`
}

// VoiceAdminIntentCuesByDocumentRes 聚合查询结果，字段与 Python 一致。
type VoiceAdminIntentCuesByDocumentRes struct {
	Items []VoiceAdminIntentCueHit `json:"items"`
}
