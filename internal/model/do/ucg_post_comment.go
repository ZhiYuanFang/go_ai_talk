// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UcgPostComment is the golang structure of table ucg_post_comment for DAO operations like Where/Data.
type UcgPostComment struct {
	g.Meta     `orm:"table:ucg_post_comment, do:true"`
	Id         interface{} //
	PostId     interface{} //
	AuthorWxId interface{} //
	Content    interface{} //
	CreatedAt  interface{} //
}
