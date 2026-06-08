// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import "time"

// UcgMediaBlob is the golang structure for table ucg_media_blob.
type UcgMediaBlob struct {
	Id               uint64    `json:"id"               ` //
	ContentHash      string    `json:"contentHash"      ` // SHA-256 hex lowercase
	TransformVersion string    `json:"transformVersion" ` // client transform pipeline version
	ObjectKey        string    `json:"objectKey"        ` // OSS object key
	MediaKind        int       `json:"mediaKind"        ` // 1=image 2=video
	RefCount         int       `json:"refCount"         ` // ownership registrations
	CreatedAt        time.Time `json:"createdAt"        ` //
}
