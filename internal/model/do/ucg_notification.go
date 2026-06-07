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
	RecipientWxId interface{} //
	Type          interface{} //
	PostId        interface{} //
	CommentId     interface{} //
	ActorWxId     interface{} //
	Preview       interface{} //
	ReadAt        interface{} //
	CreatedAt     interface{} //
}
