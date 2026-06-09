// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UcgChatMessageOutbox is the golang structure of table ucg_chat_message_outbox for DAO operations like Where/Data.
type UcgChatMessageOutbox struct {
	g.Meta         `orm:"table:ucg_chat_message_outbox, do:true"`
	Id             interface{} //
	ConversationId interface{} //
	Payload        interface{} // ChatMessage JSON
	Status         interface{} // pending|done|failed
	Attempts       interface{} //
	LastError      interface{} //
	CreatedAt      interface{} //
}
