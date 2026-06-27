package v1

import "github.com/gogf/gf/v2/frame/g"

// UcgInternalPostsSampleReq 内部轻量已发布帖抽样（供 sim 评论等，避免 recommend N+1）。
type UcgInternalPostsSampleReq struct {
	g.Meta `path:"/ucg/internal/api/posts/sample" method:"post" tags:"ucg" summary:"内部-已发布帖轻量抽样"`
	// Mode 抽样模式：缺省或 latest=按 published_at 取最新 N 条；random=全库 ID 探测随机 1 条（略偏新帖）。
	Mode  string `json:"mode"`
	Limit int    `json:"limit"`
}

// UcgInternalPostSampleItem 抽样项最小字段。
type UcgInternalPostSampleItem struct {
	PostId         uint64 `json:"postId"`
	Content        string `json:"content"`
	MediaType      int    `json:"mediaType"`
	CoverObjectKey string `json:"coverObjectKey,omitempty"`
	// CoverCdnUrl 封面 CDN URL（图文全图 / 视频首帧），供 simVision 多模态评论。
	CoverCdnUrl string `json:"coverCdnUrl,omitempty"`
}

// UcgInternalPostsSampleRes 抽样列表。
type UcgInternalPostsSampleRes struct {
	List []UcgInternalPostSampleItem `json:"list"`
}
