// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UcgNotification is the golang structure of table ucg_notification for DAO operations like Where/Data.
type UcgNotification struct {
	g.Meta        `orm:"table:ucg_notification, do:true"`
	Id            interface{} //
	RecipientWxId interface{} // 接收者 wxId
	Type          interface{} // comment_on_post | mention_in_comment
	PostId        interface{} //
	CommentId     interface{} //
	ActorWxId     interface{} // 评论者
	Preview       interface{} // 评论摘要
	PostThumbUrl  interface{} // 写入时快照的帖子封面 URL
	PostMediaKind interface{} // 0=none,1=image,2=video
	ReadAt        interface{} // NULL=未读
	CreatedAt     interface{} //
}
