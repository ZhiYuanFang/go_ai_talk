package v1

import "github.com/gogf/gf/v2/frame/g"

// UcgInternalProfilesBatchReq internal 批量查询公开 profile 展示字段。
type UcgInternalProfilesBatchReq struct {
	g.Meta `path:"/ucg/internal/api/profiles/batch" method:"post" tags:"ucg" summary:"内部-批量 profile 展示字段"`
	WxIds  []int64 `json:"wxIds"`
}

// UcgInternalProfileBatchItem 单条 profile 展示项（不含敏感字段）。
type UcgInternalProfileBatchItem struct {
	WxId               int64  `json:"wxId"`
	Nickname           string `json:"nickname"`
	AvatarUrl          string `json:"avatarUrl,omitempty"`
	AvatarThumbnailUrl string `json:"avatarThumbnailUrl,omitempty"`
}

// UcgInternalProfilesBatchRes 批量 profile 响应。
type UcgInternalProfilesBatchRes struct {
	List []UcgInternalProfileBatchItem `json:"list"`
}
