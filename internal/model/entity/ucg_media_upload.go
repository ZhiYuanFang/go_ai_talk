// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UcgMediaUpload is the golang structure for table ucg_media_upload.
type UcgMediaUpload struct {
	Id        uint64 `json:"id"        ` //
	WxId      uint64 `json:"wxId"      ` // uploader wx id
	ObjectKey string `json:"objectKey" ` // OSS object key
	MediaKind int    `json:"mediaKind" ` // 1=image 2=video
	CreatedAt int64  `json:"createdAt" ` // unix seconds
}
