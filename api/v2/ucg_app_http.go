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

// UcgPostItemV2 Feed 帖子项（v1 字段 + 评论预览）。
type UcgPostItemV2 struct {
	v1.UcgPostItem
	Comments []UcgCommentItemV2 `json:"comments,omitempty"`
}

// UcgFeedRecommendV2Req v2 推荐 Feed（含评论预览，最多 commentsPreviewMax 条/帖）。
type UcgFeedRecommendV2Req struct {
	g.Meta   `path:"/ucg/app/api/v2/feed/recommend" method:"get" tags:"ucg" summary:"推荐 Feed(v2含评论预览)"`
	Page     int      `json:"page" in:"query" d:"1"`
	PageSize int      `json:"pageSize" in:"query" d:"20"`
	Type     string   `json:"type" in:"query"`
	Lat      *float64 `json:"lat" in:"query"`
	Lng      *float64 `json:"lng" in:"query"`
	Cursor   string   `json:"cursor" in:"query"`
}

type UcgFeedRecommendV2Res struct {
	List       []UcgPostItemV2 `json:"list"`
	HasMore    bool            `json:"hasMore"`
	NextCursor string          `json:"nextCursor,omitempty"`
}

// UcgFeedFollowingV2Req v2 关注 Feed（含评论预览）。
type UcgFeedFollowingV2Req struct {
	g.Meta   `path:"/ucg/app/api/v2/feed/following" method:"get" tags:"ucg" summary:"关注 Feed(v2含评论预览)"`
	Page     int      `json:"page" in:"query" d:"1"`
	PageSize int      `json:"pageSize" in:"query" d:"20"`
	Type     string   `json:"type" in:"query"`
	Lat      *float64 `json:"lat" in:"query"`
	Lng      *float64 `json:"lng" in:"query"`
}

type UcgFeedFollowingV2Res struct {
	List     []UcgPostItemV2 `json:"list"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

// UcgPostCommentsGetV2Req v2 全量评论列表（Redis 读模型）。
type UcgPostCommentsGetV2Req struct {
	g.Meta `path:"/ucg/app/api/v2/posts/{id}/comments" method:"get" tags:"ucg" summary:"评论列表(v2 Redis)"`
	Id     uint64 `json:"id" in:"path" v:"required|min:1"`
}

type UcgCommentsListV2Res struct {
	List      []UcgCommentItemV2 `json:"list"`
	Total     int                `json:"total"`
	Truncated bool               `json:"truncated"`
}

// UcgCommentItemV2 v2 评论项（v1 字段 + 辩论立场快照）。
type UcgCommentItemV2 struct {
	v1.UcgCommentItem
	VoteSide      string `json:"voteSide,omitempty"`
	VoteSideLabel string `json:"voteSideLabel,omitempty"`
}

// UcgPostCommentPostV2Req v2 发表评论（辩论帖须已投票）。
type UcgPostCommentPostV2Req struct {
	g.Meta  `path:"/ucg/app/api/v2/posts/{id}/comments" method:"post" tags:"ucg" summary:"发表评论(v2含立场)"`
	Id      uint64 `json:"id" in:"path" v:"required|min:1"`
	Content string `json:"content" v:"required"`
}

type UcgPostCommentPostV2Res struct {
	UcgCommentItemV2
}
