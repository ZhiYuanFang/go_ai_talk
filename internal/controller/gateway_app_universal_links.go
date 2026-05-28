package controller

import (
	"net/http"

	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

// gatewayAppAppleAppSiteAssociation 返回 Apple Universal Links 所需的 AASA 验证文件。
// Team ID 未配置时返回 503 与显式错误语义，便于运维定位缺失配置。
func gatewayAppAppleAppSiteAssociation(r *ghttp.Request) {
	ctx := r.Context()
	r.Response.Header().Set("Content-Type", "application/json")
	r.Response.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	r.Response.Header().Set("X-Content-Type-Options", "nosniff")

	payload, err := gatewayapp.BuildAppleAppSiteAssociation(ctx)
	if err != nil {
		glog.Warningf(ctx, "[gateway-app-universal-links] 生成 AASA 失败 err=%v", err)
		r.Response.Status = http.StatusServiceUnavailable
		if r.Method == http.MethodHead {
			return
		}
		r.Response.WriteJson(g.Map{
			"code":           http.StatusServiceUnavailable,
			"message":        err.Error(),
			"bundleId":       gatewayapp.IOSBundleID,
			"universalLinks": gatewayapp.UniversalLinksPublicURL(ctx),
		})
		return
	}

	r.Response.Status = http.StatusOK
	if r.Method == http.MethodHead {
		return
	}
	r.Response.Write(payload)
}

// gatewayAppUniversalLinksLanding 用于浏览器直接访问 Universal Links 前缀时的回退壳页。
// 这样在未安装 App 或桌面浏览器环境中也不会看到 404。
func gatewayAppUniversalLinksLanding(r *ghttp.Request) {
	r.Response.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	r.Response.ServeFile("resource/public/pangbao-home.html")
}
