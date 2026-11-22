package controller

import (
	devicectrl "hello/internal/controller/device"
	voicectrl "hello/internal/controller/voice"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterVoiceServiceHTTP 注册 voice-service 独立进程所需路由。
// 业务说明：clinic / care-alert / growth-trajectory / AI 额度 / llm-lanes / suggest 已退役；
// 保留聊天 WS、飞轮 Admin、预测与 qa 内部口。
func RegisterVoiceServiceHTTP(s *ghttp.Server) {
	s.Use(ghttp.MiddlewareHandlerResponse)
	voicectrl.RegisterVoiceChatWS(s)
	voicectrl.RegisterVoiceAsrWS(s)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Bind(voicectrl.NewVoiceQaInternalCtrl())
		group.Bind(voicectrl.NewVoiceAdminIntentVectorsCtrl())
		group.Bind(voicectrl.NewVoiceAdminTypeVectorsCtrl())
		group.Bind(voicectrl.NewVoiceAdminOpVectorsCtrl())
		group.Bind(voicectrl.NewVoiceAdminQtyVectorsCtrl())
		group.Bind(voicectrl.NewVoiceAdminReadSourceVectorsCtrl())
		group.Bind(voicectrl.NewVoiceAdminTimeVectorsCtrl())
		group.Bind(voicectrl.NewVoiceAdminRemarkVectorsCtrl())
		group.Bind(voicectrl.NewVoiceAdminPredictVectorsCtrl())
		group.Bind(voicectrl.NewVoiceAdminLocateVectorsCtrl())
		group.Bind(voicectrl.NewVoiceAdminHistorySpeakVectorsCtrl())
		// 按原文一次查九个飞轮；删除仍走各 *-vectors。
		group.Bind(voicectrl.NewVoiceAdminIntentCuesCtrl())
		group.Bind(voicectrl.NewVoiceAdminRuleGapsCtrl())
		group.Bind(voicectrl.NewVoiceAdminMissInboxCtrl())
		// Hub 预测上报只读：imminent + active-eta（Admin 口令；非 InternalSecret）。
		group.Bind(voicectrl.NewVoiceAdminPredictReportCtrl())
		// 预测临近 / 进行中 ETA：宿主 voice（care-alert / growth-trajectory / clinic 已退役）。
		group.Bind(&voicectrl.DevicePredictImminentController{})
		group.Bind(&voicectrl.DevicePredictActiveEtaController{})
	})
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(devicectrl.InternalSecretMiddleware)
		// 进行中 ETA 内部读：供 Python；须 X-Device-Gateway-Internal-Secret。
		group.Bind(voicectrl.NewVoicePredictActiveEtaInternalCtrl())
		// 预测临近待办内部读：供 Python intent next 通道；同组密钥中间件。
		group.Bind(voicectrl.NewVoicePredictImminentInternalCtrl())
	})
}
