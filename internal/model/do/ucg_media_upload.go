// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UcgMediaUpload is the golang structure of table ucg_media_upload for DAO operations like Where/Data.
type UcgMediaUpload struct {
	g.Meta    `orm:"table:ucg_media_upload, do:true"`
	Id        interface{} //
	WxId      interface{} // uploader wx id
	ObjectKey interface{} // OSS object key
	MediaKind interface{} // 1=image 2=video
	CreatedAt interface{} // unix seconds
}
