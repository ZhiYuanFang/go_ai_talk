package v1

import (
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// VoiceInternalQaListReq 内部分页查询问答库（权威数据在 voice 库 qa 表）。
type VoiceInternalQaListReq struct {
	g.Meta   `path:"/voice/internal/api/qa/list" method:"get" tags:"voice" summary:"内部-问答库列表"`
	Page     int `json:"page" p:"page"`
	PageSize int `json:"pageSize" p:"pageSize"`
}

// VoiceInternalQaListRes 问答库分页响应。
type VoiceInternalQaListRes struct {
	List     []entity.Qa `json:"list"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// VoiceInternalQaDeleteReq 内部删除问答库行。
type VoiceInternalQaDeleteReq struct {
	g.Meta `path:"/voice/internal/api/qa/delete" method:"post" tags:"voice" summary:"内部-删除问答库"`
	Id     int64 `json:"id"`
}

// VoiceInternalQaDeleteRes 删除成功。
type VoiceInternalQaDeleteRes struct{}
