// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UcgConversationMember is the golang structure of table ucg_conversation_member for DAO operations like Where/Data.
type UcgConversationMember struct {
	g.Meta         `orm:"table:ucg_conversation_member, do:true"`
	Id             interface{} //
	ConversationId interface{} //
	WxId           interface{} //
	Pinned         interface{} //
	DeletedAt      interface{} // user soft delete
	LastReadMsgId  interface{} //
	UnreadCount    interface{} //
	UpdatedAt      interface{} // last activity; drives idx_wx_list sort
}
