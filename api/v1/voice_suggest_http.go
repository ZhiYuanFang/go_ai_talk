package v1

import (
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// VoiceSuggestListReq 内部查询成长建议列表。
type VoiceSuggestListReq struct {
	g.Meta   `path:"/voice/internal/api/suggest/list" method:"get" tags:"voice" summary:"内部-建议列表"`
	DeviceNo string `json:"deviceNo" p:"deviceNo" dc:"设备号"`
}

// VoiceSuggestListRes 建议列表响应。
type VoiceSuggestListRes struct {
	List []entity.Suggest `json:"list"`
}

// VoiceSuggestDeleteReq 内部删除成长建议。
type VoiceSuggestDeleteReq struct {
	g.Meta   `path:"/voice/internal/api/suggest/delete" method:"post" tags:"voice" summary:"内部-删除建议"`
	Id       int64  `json:"id" dc:"记录ID"`
	DeviceNo string `json:"deviceNo" dc:"设备号"`
}

// VoiceSuggestDeleteRes 删除结果。
type VoiceSuggestDeleteRes struct{}
