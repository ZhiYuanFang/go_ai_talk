package controller

import (
	ucgctrl "hello/internal/controller/ucg"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterUcgServiceHTTP 注册 ucg-service 独立进程路由（App API 与内部 WS 入口）。
func RegisterUcgServiceHTTP(s *ghttp.Server) {
	s.Use(ghttp.MiddlewareHandlerResponse)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Bind(ucgctrl.NewUcgAppCtrl())
		group.Bind(ucgctrl.NewUcgAdminCtrl())
		group.POST("/ucg/app/api/media/upload", ucgctrl.MediaUpload)
		group.POST("/ucg/internal/api/media/upload", ucgctrl.InternalMediaUpload)
		group.POST("/ucg/internal/api/media/upload-video", ucgctrl.InternalMediaUploadVideo)
		group.POST("/ucg/internal/api/chat/send", ucgctrl.InternalChatSend)
		group.POST("/ucg/internal/api/posts/sample", ucgctrl.InternalPostsSample)
		group.POST("/ucg/internal/api/chat/sim-unread-sample", ucgctrl.InternalChatSimUnreadSample)
		group.POST("/ucg/internal/api/profiles/batch", ucgctrl.InternalProfilesBatch)
		group.POST("/ucg/internal/api/force/acquire", ucgctrl.InternalForceAcquire)
	})
	ucgctrl.RegisterUcgChatWS(s)
}
