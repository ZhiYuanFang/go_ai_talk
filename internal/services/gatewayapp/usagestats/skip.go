package usagestats

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
)

// ShouldSkipRecord 判断是否跳过使用统计（与 touch 不同：在响应后结合状态码再调用 Record）。
func ShouldSkipRecord(r *ghttp.Request) bool {
	if r == nil {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(r.Header.Get("Upgrade")), "websocket") {
		return true
	}
	path := r.URL.Path
	if strings.HasPrefix(path, "/device/internal/") {
		return true
	}
	if strings.HasPrefix(path, "/device/admin/api/") {
		return true
	}
	if isStaticOrShellPath(path) {
		return true
	}
	return false
}

// IsSuccessStatus 仅统计 HTTP 2xx。
func IsSuccessStatus(status int) bool {
	return status >= http.StatusOK && status < http.StatusMultipleChoices
}

func isStaticOrShellPath(path string) bool {
	switch path {
	case "/",
		"/api.json",
		"/swagger",
		"/favicon.ico",
		"/robots.txt",
		"/device/admin",
		"/device/admin/",
		"/device/admin/qa-records",
		"/device/admin/feedback-records",
		"/device/admin/ucg-admin.html",
		"/device/admin/api-usage-stats",
		"/device/app/version-admin.html",
		"/user-agreement.html",
		"/privacy-policy.html",
		"/apple-app-site-association",
		"/.well-known/apple-app-site-association":
		return true
	}
	prefixes := []string{
		"/resource/",
		"/vendor/",
		"/device/app/apk/",
		"/wx/ulink/",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	// 运维设备数据 HTML 壳页（非 API）
	if strings.HasPrefix(path, "/device/admin/history/") && !strings.HasPrefix(path, "/device/admin/api/") {
		return true
	}
	// 旧路径壳页（302 至新路径前仍可能命中统计）
	if strings.HasPrefix(path, "/device/history/") && !strings.HasPrefix(path, "/device/history/api/") {
		return true
	}
	if strings.HasPrefix(path, "/ucg/app/api/profile/") && !strings.HasPrefix(path, "/ucg/app/api/profile/me") {
		return true
	}
	return false
}
