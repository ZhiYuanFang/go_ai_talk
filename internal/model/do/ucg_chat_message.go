// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UcgChatMessage is the golang structure of table ucg_chat_message for DAO operations like Where/Data.
type UcgChatMessage struct {
	g.Meta         `orm:"table:ucg_chat_message, do:true"`
	Id             interface{} // 会话内消息序号，与 Redis seq 一致
	ConversationId interface{} //
	ClientMsgId    interface{} // 客户端幂等 ID，空表示无
	SenderWxId     interface{} //
	Content        interface{} //
	ImageKey       interface{} //
	VideoKey       interface{} //
	MediaCdnUrl    interface{} //
	CreatedAt      interface{} //
	Status         interface{} //
}
