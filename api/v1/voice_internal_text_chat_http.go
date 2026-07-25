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

// VoiceInternalTextChatStreamReq history-service 等内部调用方触发的流式文本对话。
// 以 SSE 方式返回 thinking/answer 事件。
type VoiceInternalTextChatStreamReq struct {
	g.Meta     `path:"/voice/internal/api/text/chat/stream" method:"post" tags:"voice" summary:"内部-文本智能对话（流式SSE）"`
	DeviceNo   string `json:"deviceNo" v:"required"`
	Transcript string `json:"transcript" v:"required"`
}

// VoiceInternalTextChatStreamRes 流式文本对话响应（SSE，无固定 JSON 结构）。
type VoiceInternalTextChatStreamRes struct{}
