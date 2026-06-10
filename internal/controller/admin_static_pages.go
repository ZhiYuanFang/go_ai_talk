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
	{path: "/device/history/*deviceNo", file: "resource/public/history.html", noCache: false},
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
}

// installGatewayAdminRedirects 主网关：admin/history 壳页 302 至 App 网关。
func installGatewayAdminRedirects(s *ghttp.Server) {
	redirect := func(r *ghttp.Request) {
		target := gatewayAdminRedirectURL(r.Context(), r.URL.Path)
		r.Response.RedirectTo(target, http.StatusFound)
	}
	s.BindHandler("/device/admin", redirect)
	s.BindHandler("/device/admin/*", redirect)
	s.BindHandler("/device/history/*deviceNo", redirect)
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
