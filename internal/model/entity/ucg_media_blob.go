// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UcgMediaBlob is the golang structure for table ucg_media_blob.
type UcgMediaBlob struct {
	Id               int64       `json:"id"               ` //
	ContentHash      string      `json:"contentHash"      ` // SHA-256 hex lowercase of prepared bytes
	TransformVersion string      `json:"transformVersion" ` // client transform pipeline version
	ObjectKey        string      `json:"objectKey"        ` // OSS object key
	MediaKind        int         `json:"mediaKind"        ` // 1=image 2=video
	RefCount         int         `json:"refCount"         ` // ownership registrations referencing this blob
	CreatedAt        *gtime.Time `json:"createdAt"        ` //
}
