// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UcgPost is the golang structure for table ucg_post.
type UcgPost struct {
	Id                uint64 `json:"id"                ` //
	AuthorWxId        uint64 `json:"authorWxId"        ` //
	Type              string `json:"type"              ` // moment|debate
	Content           string `json:"content"           ` //
	DebateLeftLabel   string `json:"debateLeftLabel"   ` //
	DebateRightLabel  string `json:"debateRightLabel"  ` //
	IpLocation        string   `json:"ipLocation"        ` // 发帖时 IP属地快照
	Lat               *float64 `json:"lat"               ` // 可选发帖纬度 WGS84
	Lng               *float64 `json:"lng"               ` // 可选发帖经度 WGS84
	Status            int      `json:"status"            ` // 0 draft 1 pending_audit 2 published 3 rejected 4 apply_failed 5 moderation_failed
	AuditVersion      int    `json:"auditVersion"      ` // 审核轮次，仅 submit/再提审时递增
	ModerationVerdict int    `json:"moderationVerdict" ` // 0=未审 1=pass 2=reject
	ModerationReason  string `json:"moderationReason"  ` //
	ModerationAt      int64  `json:"moderationAt"      ` //
	ApplyAttempts     int    `json:"applyAttempts"     ` //
	ApplyFailedAt     int64  `json:"applyFailedAt"     ` //
	RejectReason      string `json:"rejectReason"      ` //
	MediaType         int    `json:"mediaType"         ` // 0 none 1 images 2 video
	LikeCount         uint   `json:"likeCount"         ` //
	CommentCount      uint   `json:"commentCount"      ` //
	CreatedAt         int64  `json:"createdAt"         ` //
	UpdatedAt         int64  `json:"updatedAt"         ` //
	PublishedAt       int64  `json:"publishedAt"       ` //
}
