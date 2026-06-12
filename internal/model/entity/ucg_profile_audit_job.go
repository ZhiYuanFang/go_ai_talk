// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UcgProfileAuditJob 资料/头像待审 patch（audit_version 权威源）。
type UcgProfileAuditJob struct {
	Id                uint64 `json:"id"                ` //
	WxId              uint64 `json:"wxId"              ` //
	Nickname          string `json:"nickname"          ` //
	AvatarKey         string `json:"avatarKey"         ` //
	Bio               string `json:"bio"               ` //
	Status            int    `json:"status"            ` // 1 pending 2 approved 3 rejected 4 apply_failed
	AuditVersion      int    `json:"auditVersion"      ` // 审核轮次
	ModerationVerdict int    `json:"moderationVerdict" ` // 0=未审 1=pass 2=reject
	ModerationReason  string `json:"moderationReason"  ` //
	ModerationAt      int64  `json:"moderationAt"      ` //
	ApplyAttempts     int    `json:"applyAttempts"     ` //
	ApplyFailedAt     int64  `json:"applyFailedAt"     ` //
	RejectReason      string `json:"rejectReason"      ` //
	CreatedAt         int64  `json:"createdAt"         ` //
	UpdatedAt         int64  `json:"updatedAt"         ` //
}
