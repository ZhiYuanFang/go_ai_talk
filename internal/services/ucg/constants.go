package ucg

const (
	PostStatusDraft        = 0
	PostStatusPendingAudit = 1
	PostStatusPublished    = 2
	PostStatusRejected     = 3

	CommentStatusDraft        = 0
	CommentStatusPendingAudit = 1
	CommentStatusPublished    = 2
	CommentStatusRejected     = 3

	ProfileJobStatusPending  = 1
	ProfileJobStatusApproved = 2
	ProfileJobStatusRejected = 3

	ChatAuditStatusPending  = "pending"
	ChatAuditStatusApproved = "approved"
	ChatAuditStatusRejected = "rejected"

	MediaTypeNone   = 0
	MediaTypeImages = 1
	MediaTypeVideo  = 2
)
