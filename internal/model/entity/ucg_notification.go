// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UcgNotification is the golang structure for table ucg_notification.
type UcgNotification struct {
	Id            uint64 `json:"id"            ` //
	RecipientWxId uint64 `json:"recipientWxId" ` //
	Type          string `json:"type"          ` //
	PostId        uint64 `json:"postId"        ` //
	CommentId     uint64 `json:"commentId"     ` //
	ActorWxId     uint64 `json:"actorWxId"     ` //
	Preview       string `json:"preview"       ` //
	ReadAt        int64  `json:"readAt"        ` //
	CreatedAt     int64  `json:"createdAt"     ` //
}
