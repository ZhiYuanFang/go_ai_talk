// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UcgPostMedia is the golang structure for table ucg_post_media.
type UcgPostMedia struct {
	Id         uint64 `json:"id"         ` //
	PostId     uint64 `json:"postId"     ` //
	SortOrder  int    `json:"sortOrder"  ` //
	ObjectKey  string `json:"objectKey"  ` //
	MediaKind  int    `json:"mediaKind"  ` // 1 image 2 video
	DurationMs int    `json:"durationMs" ` //
	SizeBytes  int64  `json:"sizeBytes"  ` //
}
