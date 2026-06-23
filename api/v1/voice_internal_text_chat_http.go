package v1

import "github.com/gogf/gf/v2/frame/g"

// VoiceInternalTextChatReq history-service 等内部调用方经 secret 中间件触发的文本对话。
type VoiceInternalTextChatReq struct {
	g.Meta     `path:"/voice/internal/api/text/chat" method:"post" tags:"voice" summary:"内部-文本智能对话"`
	DeviceNo   string `json:"deviceNo" v:"required"`
	Transcript string `json:"transcript" v:"required"`
}

// VoiceInternalTextChatRes 文本对话回复。
type VoiceInternalTextChatRes struct {
	Reply string `json:"reply"`
}
