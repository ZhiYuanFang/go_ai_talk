package controller

import (
	voice "hello/internal/services/voice"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterVoiceServiceHTTP 注册 voice-service 独立进程所需路由。
func RegisterVoiceServiceHTTP(s *ghttp.Server) {
	// 统一响应包装，保持与 gateway 暴露层的返回结构一致。
	s.Use(ghttp.MiddlewareHandlerResponse)
	// 语音会话仍通过 WS 通道承载，避免文本接口和流式接口割裂。
	registerVoiceChatWS(s)
	s.Group("/", func(group *ghttp.RouterGroup) {
		// voice-service 仅绑定语音文本域能力，避免引入非语音职责。
		group.Bind(NewVoiceTextCtrl(voice.Voice(), voice.DeviceAdmin()), Voice)
		group.Bind(NewVoiceSuggestInternalCtrl())
		group.Bind(NewVoiceQaInternalCtrl())
	})
}
