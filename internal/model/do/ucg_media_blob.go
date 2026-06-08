// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UcgMediaBlob is the golang structure of table ucg_media_blob for DAO operations like Where/Data.
type UcgMediaBlob struct {
	g.Meta           `orm:"table:ucg_media_blob, do:true"`
	Id               interface{} //
	ContentHash      interface{} // SHA-256 hex lowercase
	TransformVersion interface{} // client transform pipeline version
	ObjectKey        interface{} // OSS object key
	MediaKind        interface{} // 1=image 2=video
	RefCount         interface{} // ownership registrations
	CreatedAt        interface{} //
}
