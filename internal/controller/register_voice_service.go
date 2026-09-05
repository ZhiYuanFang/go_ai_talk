package controller

import (
	devicectrl "hello/internal/controller/device"
	voicectrl "hello/internal/controller/voice"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterVoiceServiceHTTP 注册 voice-service 独立进程所需路由。
func RegisterVoiceServiceHTTP(s *ghttp.Server) {
	s.Use(ghttp.MiddlewareHandlerResponse)
	voicectrl.RegisterVoiceChatWS(s)
	voicectrl.RegisterVoiceAsrWS(s)
	voicectrl.RegisterVoiceClinicWS(s)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Bind(voicectrl.NewVoiceSuggestInternalCtrl())
		group.Bind(voicectrl.NewVoiceQaInternalCtrl())
		group.Bind(voicectrl.NewVoiceAppAIQuotaCtrl())
		group.Bind(voicectrl.NewVoiceAdminAIQuotaCtrl())
		group.Bind(voicectrl.NewVoiceAdminLLMLanesCtrl())
		// tip SSE / clinic|tip HTTP 飞轮已下线（remove-tip-and-clinic-feedback）；care-alert 飞轮保留。
		group.Bind(&voicectrl.DeviceCareAlertController{})
	})
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(devicectrl.InternalSecretMiddleware)
		group.Bind(voicectrl.NewVoiceAIQuotaInternalCtrl())
	})
}
