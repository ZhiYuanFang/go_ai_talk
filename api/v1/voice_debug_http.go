package v1

import "github.com/gogf/gf/v2/frame/g"

// VoiceTextChatReq 文本对话（仅走内部 chat，不经 STT/TTS）。
// 需在 Header 携带 X-Admin-Password。
type VoiceTextChatReq struct {
	g.Meta     `path:"/voice/text/chat" method:"post" tags:"voice" summary:"文本对话 chat"`
	DeviceNo   string `json:"deviceNo" dc:"设备号"`
	Transcript string `json:"transcript" dc:"模拟用户转写文本"`
}

// VoiceTextChatRes 文本对话返回。
type VoiceTextChatRes struct {
	Reply string `json:"reply"`
}
