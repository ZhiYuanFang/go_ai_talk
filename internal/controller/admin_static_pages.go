package controller

import (
	"context"
	"net/http"
	"strings"

	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/net/ghttp"
)

type adminStaticPage struct {
	path string
	file string
	// noCache 为 true 时对 HTML 设置 no-store，避免运维页缓存旧脚本。
	noCache bool
}

var adminStaticPages = []adminStaticPage{
	{path: "/device/admin", file: "resource/public/admin.html", noCache: true},
	{path: "/device/admin/qa-records", file: "resource/public/qa-records.html", noCache: true},
	{path: "/device/admin/feedback-records", file: "resource/public/feedback-records.html", noCache: true},
	{path: "/device/admin/api-usage-stats", file: "resource/public/api-usage-stats.html", noCache: true},
	{path: "/device/admin/ucg-admin.html", file: "resource/public/ucg-admin.html", noCache: true},
	{path: "/device/admin/voice-admin.html", file: "resource/public/voice-admin.html", noCache: true},
	{path: "/device/admin/ai-model-admin.html", file: "resource/public/ai-model-admin.html", noCache: true},
	{path: "/device/admin/sim-admin.html", file: "resource/public/sim-admin.html", noCache: true},
	{path: "/device/admin/cash-vip-admin.html", file: "resource/public/cash-vip-admin.html", noCache: true},
	{path: "/device/admin/cash-feature-admin.html", file: "resource/public/cash-feature-admin.html", noCache: true},
	{path: "/device/admin/cash-invite-code-admin.html", file: "resource/public/cash-invite-code-admin.html", noCache: true},
	{path: "/device/admin/history/*deviceNo", file: "resource/public/history.html", noCache: true},
	{path: "/device/app/version-admin.html", file: "resource/public/gateway-app-version-admin.html", noCache: true},
}

// RegisterAdminStaticPages 在 App 网关单点注册全部 Web 管理静态页。
func RegisterAdminStaticPages(s *ghttp.Server) {
	for _, p := range adminStaticPages {
		page := p
		s.BindHandler(page.path, func(r *ghttp.Request) {
			if page.noCache {
				r.Response.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
			}
			r.Response.ServeFile(page.file)
		})
	}
	// 旧壳页路径 302 至 /device/admin/history/{deviceNo}，兼容书签与旧链接。
	s.BindHandler("/device/history/*deviceNo", redirectLegacyHistoryShell)
}

// redirectLegacyHistoryShell 将 /device/history/{deviceNo} 重定向至运维新路径。
func redirectLegacyHistoryShell(r *ghttp.Request) {
	deviceNo := legacyHistoryDeviceNoFromPath(r.URL.Path)
	if deviceNo == "" {
		r.Response.RedirectTo("/device/admin", http.StatusFound)
		return
	}
	r.Response.RedirectTo("/device/admin/history/"+deviceNo, http.StatusFound)
}

func legacyHistoryDeviceNoFromPath(path string) string {
	const prefix = "/device/history/"
	path = strings.TrimSpace(path)
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	rest := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	if rest == "" || strings.Contains(rest, "/") {
		return ""
	}
	return rest
}

// installGatewayAdminRedirects 主网关：admin/history 壳页 302 至 App 网关。
func installGatewayAdminRedirects(s *ghttp.Server) {
	redirect := func(r *ghttp.Request) {
		target := gatewayAdminRedirectURL(r.Context(), r.URL.Path)
		r.Response.RedirectTo(target, http.StatusFound)
	}
	s.BindHandler("/device/admin", redirect)
	s.BindHandler("/device/admin/*", redirect)
	// 旧 history 壳页：主网关 302 至 App 网关新路径。
	s.BindHandler("/device/history/*deviceNo", func(r *ghttp.Request) {
		deviceNo := legacyHistoryDeviceNoFromPath(r.URL.Path)
		if deviceNo == "" {
			target := gatewayAdminRedirectURL(r.Context(), "/device/admin")
			r.Response.RedirectTo(target, http.StatusFound)
			return
		}
		target := gatewayAdminRedirectURL(r.Context(), "/device/admin/history/"+deviceNo)
		r.Response.RedirectTo(target, http.StatusFound)
	})
}

func gatewayAdminRedirectURL(ctx context.Context, path string) string {
	base := strings.TrimRight(strings.TrimSpace(gatewayapp.PublicBaseURL(ctx)), "/")
	if base == "" {
		base = "http://127.0.0.1:9702"
	}
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path
}
