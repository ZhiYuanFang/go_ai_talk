// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UcgPostMedia is the golang structure of table ucg_post_media for DAO operations like Where/Data.
type UcgPostMedia struct {
	g.Meta     `orm:"table:ucg_post_media, do:true"`
	Id         interface{} //
	PostId     interface{} //
	SortOrder  interface{} //
	ObjectKey  interface{} //
	MediaKind  interface{} // 1 image 2 video
	DurationMs interface{} //
	SizeBytes  interface{} //
}
