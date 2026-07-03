package v2

import (
	v1 "hello/api/v1"

	"github.com/gogf/gf/v2/frame/g"
)

// UcgPostCreateV2Req v2 创建帖（可选 lat/lng）；计入 App API usage 统计。
type UcgPostCreateV2Req struct {
	g.Meta      `path:"/ucg/app/api/v2/posts" method:"post" tags:"ucg" summary:"创建帖子(v2含坐标)"`
	Content     string              `json:"content"`
	Type        string              `json:"type"`
	DebateLeft  string              `json:"debateLeft"`
	DebateRight string              `json:"debateRight"`
	MediaType   int                 `json:"mediaType"`
	Submit      bool                `json:"submit"`
	Media       []v1.UcgPostMediaInput `json:"media"`
	Lat         *float64            `json:"lat"`
	Lng         *float64            `json:"lng"`
}

// UcgPostCreateV2Res 与 v1 创建帖响应一致。
type UcgPostCreateV2Res struct {
	v1.UcgPostItem
}
