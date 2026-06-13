// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UcgPostComment is the golang structure for table ucg_post_comment.
type UcgPostComment struct {
	Id                uint64 `json:"id"                ` //
	PostId            uint64 `json:"postId"            ` //
	AuthorWxId        uint64 `json:"authorWxId"        ` //
	Content           string `json:"content"           ` //
	Status            int    `json:"status"            ` // 0 draft 1 pending 2 published 3 rejected 4 apply_failed 5 moderation_failed
	AuditVersion      int    `json:"auditVersion"      ` // 审核轮次
	ModerationVerdict int    `json:"moderationVerdict" ` // 0=未审 1=pass 2=reject
	ModerationReason  string `json:"moderationReason"  ` //
	ModerationAt      int64  `json:"moderationAt"      ` //
	ApplyAttempts     int    `json:"applyAttempts"     ` //
	ApplyFailedAt     int64  `json:"applyFailedAt"     ` //
	RejectReason      string `json:"rejectReason"      ` //
	CreatedAt         int64  `json:"createdAt"         ` //
}
