package v1

import "github.com/gogf/gf/v2/frame/g"

// UcgInternalPostsSampleReq 内部轻量已发布帖抽样（供 sim 评论等，避免 recommend N+1）。
type UcgInternalPostsSampleReq struct {
	g.Meta `path:"/ucg/internal/api/posts/sample" method:"post" tags:"ucg" summary:"内部-已发布帖轻量抽样"`
	Limit  int `json:"limit"`
}

// UcgInternalPostSampleItem 抽样项最小字段。
type UcgInternalPostSampleItem struct {
	PostId         uint64 `json:"postId"`
	Content        string `json:"content"`
	MediaType      int    `json:"mediaType"`
	CoverObjectKey string `json:"coverObjectKey,omitempty"`
}

// UcgInternalPostsSampleRes 抽样列表。
type UcgInternalPostsSampleRes struct {
	List []UcgInternalPostSampleItem `json:"list"`
}
