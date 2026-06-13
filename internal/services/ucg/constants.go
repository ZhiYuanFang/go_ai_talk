package ucg

const (
	PostStatusDraft        = 0
	PostStatusPendingAudit = 1
	PostStatusPublished    = 2
	PostStatusRejected     = 3
	PostStatusApplyFailed  = 4 // 机审已完成但 apply 超限失败

	CommentStatusDraft        = 0
	CommentStatusPendingAudit = 1
	CommentStatusPublished    = 2
	CommentStatusRejected     = 3

	ProfileJobStatusPending          = 1
	ProfileJobStatusApproved         = 2
	ProfileJobStatusRejected         = 3
	ProfileJobStatusApplyFailed      = 4 // 机审已完成但 apply 超限失败
	ProfileJobStatusModerationFailed = 5 // 尝试 Green 失败/写库失败

	// ModerationVerdict Phase1 机审结论（MySQL 权威，MQ 重投时跳过 Green 的依据）。
	ModerationVerdictNone   = 0
	ModerationVerdictPass   = 1
	ModerationVerdictReject = 2

	// applyFailedSystemReason apply 超限后作者可见固定文案（非 Green reason）。
	applyFailedSystemReason = "审核异常，请稍后重试"

	ChatAuditStatusPending  = "pending"
	ChatAuditStatusApproved = "approved"
	ChatAuditStatusRejected = "rejected"

	MediaTypeNone   = 0
	MediaTypeImages = 1
	MediaTypeVideo  = 2
)
