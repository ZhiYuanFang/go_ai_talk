package v1

import (
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// VoiceInternalQaListReq 内部查询问答库全表（权威数据在 voice 库 qa 表）。
type VoiceInternalQaListReq struct {
	g.Meta `path:"/voice/internal/api/qa/list" method:"get" tags:"voice" summary:"内部-问答库列表"`
}

// VoiceInternalQaListRes 问答库列表响应。
type VoiceInternalQaListRes struct {
	List []entity.Qa `json:"list"`
}
