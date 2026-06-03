// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UcgPost is the golang structure for table ucg_post.
type UcgPost struct {
	Id           uint64 `json:"id"           ` //
	AuthorWxId   uint64 `json:"authorWxId"   ` //
	Content      string `json:"content"      ` //
	Status       int    `json:"status"       ` // 0 draft 1 pending_audit 2 published 3 rejected
	RejectReason string `json:"rejectReason" ` //
	MediaType    int    `json:"mediaType"    ` // 0 none 1 images 2 video
	LikeCount    uint   `json:"likeCount"    ` //
	CommentCount uint   `json:"commentCount" ` //
	CreatedAt    int64  `json:"createdAt"    ` //
	UpdatedAt    int64  `json:"updatedAt"    ` //
	PublishedAt  int64  `json:"publishedAt"  ` //
}
