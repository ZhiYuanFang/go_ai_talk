// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UcgNotification is the golang structure for table ucg_notification.
type UcgNotification struct {
	Id            uint64 `json:"id"            ` //
	RecipientWxId uint64 `json:"recipientWxId" ` // 接收者 wxId
	Type          string `json:"type"          ` // comment_on_post | mention_in_comment
	PostId        uint64 `json:"postId"        ` //
	CommentId     uint64 `json:"commentId"     ` //
	ActorWxId     uint64 `json:"actorWxId"     ` // 评论者
	Preview       string `json:"preview"       ` // 评论摘要
	PostThumbUrl  string `json:"postThumbUrl"  ` // 写入时快照的帖子封面 URL
	PostMediaKind int    `json:"postMediaKind" ` // 0=none,1=image,2=video
	ReadAt        int64  `json:"readAt"        ` // NULL=未读
	CreatedAt     int64  `json:"createdAt"     ` //
}
