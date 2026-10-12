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
		group.Bind(voicectrl.NewVoiceAdminIntentVectorsCtrl())
		group.Bind(voicectrl.NewVoiceAdminTypeVectorsCtrl())
		group.Bind(voicectrl.NewVoiceAdminOpVectorsCtrl())
		group.Bind(voicectrl.NewVoiceAdminQtyVectorsCtrl())
		group.Bind(voicectrl.NewVoiceAdminReadSourceVectorsCtrl())
		group.Bind(voicectrl.NewVoiceAdminHistorySpeakVectorsCtrl())
		group.Bind(voicectrl.NewVoiceAdminRuleGapsCtrl())
		// Hub 预测上报只读：imminent + active-eta（Admin 口令；非 InternalSecret）。
		group.Bind(voicectrl.NewVoiceAdminPredictReportCtrl())
		// tip SSE / clinic|tip HTTP 飞轮已下线（remove-tip-and-clinic-feedback）；care-alert / growth-trajectory 保留。
		group.Bind(&voicectrl.DeviceCareAlertController{})
		group.Bind(&voicectrl.DeviceGrowthTrajectoryController{})
		group.Bind(&voicectrl.DevicePredictImminentController{})
		group.Bind(&voicectrl.DevicePredictActiveEtaController{})
	})
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(devicectrl.InternalSecretMiddleware)
		group.Bind(voicectrl.NewVoiceAIQuotaInternalCtrl())
		// 进行中 ETA 内部读：供 Python；须 X-Device-Gateway-Internal-Secret。
		group.Bind(voicectrl.NewVoicePredictActiveEtaInternalCtrl())
		// 预测临近待办内部读：供 Python intent next 通道；同组密钥中间件。
		group.Bind(voicectrl.NewVoicePredictImminentInternalCtrl())
	})
}
