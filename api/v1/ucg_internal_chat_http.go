package v1

import "github.com/gogf/gf/v2/frame/g"

// UcgInternalChatSendReq 内部模拟用户发聊天消息。
type UcgInternalChatSendReq struct {
	g.Meta         `path:"/ucg/internal/api/chat/send" method:"post" tags:"ucg" summary:"内部-模拟用户发送聊天消息"`
	SenderWxId     int64  `json:"senderWxId" v:"required|min:1"`
	ConversationId uint64 `json:"conversationId" v:"required|min:1"`
	ClientMsgId    string `json:"clientMsgId"`
	Content        string `json:"content"`
	ImageKey       string `json:"imageKey"`
	VideoKey       string `json:"videoKey"`
}

type UcgInternalChatSendRes struct{}
