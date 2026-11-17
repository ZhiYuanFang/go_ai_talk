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
		// 版本 APK：独立路由与大小上限，避免与 logo 2MB / 媒体 25MB 纠缠。
		group.POST("/ucg/internal/api/media/upload-apk", ucgctrl.InternalMediaUploadApk)
		group.POST("/ucg/internal/api/media/delete-apk", ucgctrl.InternalMediaDeleteApk)
		group.POST("/ucg/internal/api/profiles/batch", ucgctrl.InternalProfilesBatch)
		group.POST("/ucg/internal/api/force/acquire", ucgctrl.InternalForceAcquire)
	})
	ucgctrl.RegisterUcgChatWS(s)
}
