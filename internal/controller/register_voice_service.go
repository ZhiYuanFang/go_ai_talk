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
	// 实时听写专用 WS，与对话 WS 分离（无 TTS/LLM/单设备连接互踢）。
	registerVoiceAsrWS(s)
	registerVoiceClinicWS(s)
	s.Group("/", func(group *ghttp.RouterGroup) {
		// voice-service 仅绑定语音文本域能力，避免引入非语音职责。
		group.Bind(NewVoiceTextCtrl(voice.Voice(), voice.DeviceAdmin()), Voice)
		group.Bind(NewVoiceSuggestInternalCtrl())
		group.Bind(NewVoiceQaInternalCtrl())
		group.Bind(NewVoiceAppAIQuotaCtrl())
		group.Bind(NewVoiceAdminAIQuotaCtrl())
		group.Bind(NewVoiceAdminLLMLanesCtrl())
		// 小贴士流式生成宿主为 voice（非 history）：对外 POST /device/tip/generate（SSE）。
		group.Bind(NewTipCtrl())
		// clinic/tip 反馈飞轮：POST /device/api/clinic|tip/feedback → Python AI（同宿主 voice）。
		group.Bind(&DeviceClinicFeedbackController{})
	})
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(deviceUcgInternalSecretMiddleware)
		group.Bind(NewVoiceAIQuotaInternalCtrl(), NewVoiceInternalTextChatCtrl())
	})
}
