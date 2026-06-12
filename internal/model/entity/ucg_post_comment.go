// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UcgPostComment is the golang structure for table ucg_post_comment.
type UcgPostComment struct {
	Id         uint64 `json:"id"         ` //
	PostId     uint64 `json:"postId"     ` //
	AuthorWxId uint64 `json:"authorWxId" ` //
	Content      string `json:"content"      ` //
	Status       int    `json:"status"       ` // 0 draft 1 pending_audit 2 published 3 rejected
	AuditVersion int    `json:"auditVersion" ` // 审核轮次
	RejectReason string `json:"rejectReason" ` //
	CreatedAt    int64  `json:"createdAt"    ` //
}
