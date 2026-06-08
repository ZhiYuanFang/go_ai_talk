// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UcgPost is the golang structure of table ucg_post for DAO operations like Where/Data.
type UcgPost struct {
	g.Meta       `orm:"table:ucg_post, do:true"`
	Id           interface{} //
	AuthorWxId   interface{} //
	Content      interface{} //
	IpLocation   interface{} // 发帖时 IP属地快照
	Status       interface{} // 0 draft 1 pending_audit 2 published 3 rejected
	RejectReason interface{} //
	MediaType    interface{} // 0 none 1 images 2 video
	LikeCount    interface{} //
	CommentCount interface{} //
	CreatedAt    interface{} //
	UpdatedAt    interface{} //
	PublishedAt  interface{} //
}
