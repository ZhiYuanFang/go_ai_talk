package v1

import "github.com/gogf/gf/v2/frame/g"

type UcgAdminAiConfigGetReq struct {
	g.Meta `path:"/ucg/admin/api/ai-config" method:"get" tags:"ucg-admin" summary:"读取 UCG AI 配置"`
}

type UcgAdminAiConfigGetRes struct {
	Provider            string            `json:"provider"`
	VisionModel         string            `json:"visionModel"`
	MaxImagesPerRequest int               `json:"maxImagesPerRequest"`
	MaxInFlight         int               `json:"maxInFlight"`
	MaxWaiters          int               `json:"maxWaiters"`
	UpdatedAt           int64             `json:"updatedAt"`
	UpdatedBy           string            `json:"updatedBy"`
	AllowedModels       []string          `json:"allowedModels"`
	AllowedProviders    []string          `json:"allowedProviders"`
	ProviderModels      map[string][]string `json:"providerModels"`
}

type UcgAdminAiConfigPutReq struct {
	g.Meta              `path:"/ucg/admin/api/ai-config" method:"put" tags:"ucg-admin" summary:"更新 UCG AI 配置"`
	Provider            string `json:"provider" v:"required"`
	VisionModel         string `json:"visionModel" v:"required"`
	MaxImagesPerRequest int    `json:"maxImagesPerRequest" v:"required|between:1,9"`
	MaxInFlight         int    `json:"maxInFlight" v:"required|min:1"`
	MaxWaiters          int    `json:"maxWaiters" v:"required|min:0"`
	UpdatedBy           string `json:"updatedBy"`
}

type UcgAdminAiConfigPutRes struct {
	Provider            string `json:"provider"`
	VisionModel         string `json:"visionModel"`
	MaxImagesPerRequest int    `json:"maxImagesPerRequest"`
	MaxInFlight         int    `json:"maxInFlight"`
	MaxWaiters          int    `json:"maxWaiters"`
	UpdatedAt           int64  `json:"updatedAt"`
	UpdatedBy           string `json:"updatedBy"`
}

// UcgAdminPostsListReq 管理端动态分页列表。
type UcgAdminPostsListReq struct {
	g.Meta   `path:"/ucg/admin/api/posts/list" method:"get" tags:"ucg-admin" summary:"UCG 动态分页列表"`
	Page     int  `json:"page" p:"page" dc:"页码，从 1 开始"`
	PageSize int  `json:"pageSize" p:"pageSize" dc:"每页条数，默认 20，最大 100"`
	Status   *int `json:"status" p:"status" dc:"可选：0 draft 1 pending 2 published 3 rejected"`
}

// UcgAdminPostAuthor 管理端列表作者摘要。
type UcgAdminPostAuthor struct {
	Nickname string `json:"nickname,omitempty"`
}

// UcgAdminPostMediaItem 管理端列表媒体项。
type UcgAdminPostMediaItem struct {
	ObjectKey    string `json:"objectKey"`
	CdnUrl       string `json:"cdnUrl"`
	ThumbnailUrl string `json:"thumbnailUrl,omitempty"`
	MediaKind    int    `json:"mediaKind"`
}

// UcgAdminPostItem 管理端动态列表项。
type UcgAdminPostItem struct {
	Id           uint64                  `json:"id"`
	AuthorWxId   uint64                  `json:"authorWxId"`
	Content      string                  `json:"content"`
	Status       int                     `json:"status"`
	RejectReason string                  `json:"rejectReason,omitempty"`
	CreatedAt    int64                   `json:"createdAt"`
	UpdatedAt    int64                   `json:"updatedAt"`
	PublishedAt  int64                   `json:"publishedAt,omitempty"`
	Media        []UcgAdminPostMediaItem `json:"media,omitempty"`
	Author       *UcgAdminPostAuthor     `json:"author,omitempty"`
}

// UcgAdminPostsListRes 管理端动态分页响应。
type UcgAdminPostsListRes struct {
	List     []UcgAdminPostItem `json:"list"`
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
}

// UcgAdminPostsRejectReq 管理端批量驳回。
type UcgAdminPostsRejectReq struct {
	g.Meta  `path:"/ucg/admin/api/posts/reject" method:"post" tags:"ucg-admin" summary:"UCG 动态批量驳回"`
	PostIds []uint64 `json:"postIds" v:"required"`
	Reason  string   `json:"reason" dc:"可选，空则用默认文案"`
}

// UcgAdminPostsRejectRes 批量驳回结果。
type UcgAdminPostsRejectRes struct {
	Rejected []uint64 `json:"rejected"`
	Skipped  []uint64 `json:"skipped"`
	Failed   []uint64 `json:"failed"`
}

// UcgAdminProfileAuditJobsListReq 资料机审失败 job 列表。
type UcgAdminProfileAuditJobsListReq struct {
	g.Meta   `path:"/ucg/admin/api/profile-audit-jobs/list" method:"get" tags:"ucg-admin" summary:"资料机审失败 job 分页列表"`
	Page     int  `json:"page" p:"page"`
	PageSize int  `json:"pageSize" p:"pageSize"`
	Status   *int `json:"status" p:"status" dc:"默认 5=moderation_failed"`
}

// UcgAdminProfileAuditJobItem 列表项。
type UcgAdminProfileAuditJobItem struct {
	JobId        uint64 `json:"jobId"`
	WxId         int64  `json:"wxId"`
	AuditVersion int    `json:"auditVersion"`
	Status       int    `json:"status"`
	Nickname     string `json:"nickname,omitempty"`
	AvatarKey    string `json:"avatarKey,omitempty"`
	Bio          string `json:"bio,omitempty"`
	RejectReason string `json:"rejectReason,omitempty"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
}

// UcgAdminProfileAuditJobsListRes 列表响应。
type UcgAdminProfileAuditJobsListRes struct {
	List     []UcgAdminProfileAuditJobItem `json:"list"`
	Total    int                           `json:"total"`
	Page     int                           `json:"page"`
	PageSize int                           `json:"pageSize"`
}

// UcgAdminProfileAuditJobResolveReq 人工通过/驳回。
type UcgAdminProfileAuditJobResolveReq struct {
	g.Meta `path:"/ucg/admin/api/profile-audit-jobs/resolve" method:"post" tags:"ucg-admin" summary:"资料机审失败 job 人工处理"`
	JobId  uint64 `json:"jobId" v:"required"`
	Action string `json:"action" v:"required|in:approve,reject"`
	Reason string `json:"reason"`
}

// UcgAdminProfileAuditJobResolveRes 处理结果。
type UcgAdminProfileAuditJobResolveRes struct {
	Ok bool `json:"ok"`
}
