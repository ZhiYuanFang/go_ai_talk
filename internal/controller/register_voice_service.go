package controller

import (
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
		// voice-service：额度/车道/护理留意等；自然语言喂养仅经 /voice/chat/ws。
		group.Bind(NewVoiceSuggestInternalCtrl())
		group.Bind(NewVoiceQaInternalCtrl())
		group.Bind(NewVoiceAppAIQuotaCtrl())
		group.Bind(NewVoiceAdminAIQuotaCtrl())
		group.Bind(NewVoiceAdminLLMLanesCtrl())
		// clinic 显式 feedback 已删除（仅 Gateway 内隐式采纳飞轮）。
		// 护理留意日缓存：GET/DELETE/POST /device/api/care-alert/*（feedback 仅为 UI，无飞轮）。
		group.Bind(&DeviceCareAlertController{})
	})
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(deviceUcgInternalSecretMiddleware)
		group.Bind(NewVoiceAIQuotaInternalCtrl())
	})
}
