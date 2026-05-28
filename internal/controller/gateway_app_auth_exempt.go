package controller

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
)

// 以下为 gateway-app Bearer 白名单数据，仅在此维护路径集合；匹配规则见 gatewayAppPathAuthExempt。

var (
	// 任意 HTTP 方法：路径以前缀命中即豁免（内部接口、Swagger 等）。
	gatewayAppAuthExemptPrefixesAnyMethod = []string{
		"/device/app/api/user/internal/",
		"/swagger",
	}

	// POST 且路径精确匹配。
	gatewayAppAuthExemptExactPOST = []string{
		"/device/app/api/login",
		"/device/app/api/device_login",
		"/device/app/api/token/refresh",
		"/device/app/api/user/login",
		"/device/app/api/user/device_login",
		"/device/app/api/version/admin/login",
		"/device/app/api/version/admin/upload",
		"/device/app/api/version/admin/update",
		"/device/app/api/version/admin/delete",
	}

	// GET 且路径精确匹配（WebSocket Upgrade 等不要求 HTTP 层 Bearer）。
	gatewayAppAuthExemptExactGET = []string{
		"/device/app/ws/history",
		"/voice/chat/ws",
		"/voice/asr/ws",
	}

	// GET 或 HEAD：路径精确匹配。
	gatewayAppAuthExemptExactGETHEAD = []string{
		"/",
		"/apple-app-site-association",
		"/.well-known/apple-app-site-association",
		"/api.json",
		"/device/admin",
		"/device/admin/",
		"/device/app/api/site/home",
		"/device/app/api/version/check",
		"/device/app/api/version/admin/list",
		"/device/app/api/version/admin/get",
		"/device/app/version-admin.html",
		"/device/app/integration-test.html",
		"/favicon.ico",
		"/robots.txt",
	}

	// GET 或 HEAD：路径以前缀命中即豁免。
	gatewayAppAuthExemptPrefixesGETHEAD = []string{
		"/wx/ulink/",
		"/resource/",
		"/device/app/apk/",
		"/ai_talk_images/",
	}
)

// gatewayAppAuthPrefixExcept 表示「以前缀放行但排除更具体前缀」的壳页类路由。
type gatewayAppAuthPrefixExcept struct {
	Prefix        string // 例如历史 HTML 壳 /device/history/
	ExcludePrefix string // 需鉴权的 API，如 /device/history/api/
}

var gatewayAppAuthExemptPrefixGETHEADExcept = []gatewayAppAuthPrefixExcept{
	{Prefix: "/device/history/", ExcludePrefix: "/device/history/api/"},
}

func gatewayAppPathAuthExempt(r *ghttp.Request) bool {
	path := r.URL.Path
	method := r.Method

	if method == http.MethodOptions {
		return true
	}

	for _, p := range gatewayAppAuthExemptPrefixesAnyMethod {
		if strings.HasPrefix(path, p) {
			return true
		}
	}

	if method == http.MethodGet && stringInSlice(path, gatewayAppAuthExemptExactGET) {
		return true
	}

	if method == http.MethodPost && stringInSlice(path, gatewayAppAuthExemptExactPOST) {
		return true
	}

	if method == http.MethodGet || method == http.MethodHead {
		if stringInSlice(path, gatewayAppAuthExemptExactGETHEAD) {
			return true
		}
		for _, p := range gatewayAppAuthExemptPrefixesGETHEAD {
			if strings.HasPrefix(path, p) {
				return true
			}
		}
		for _, rule := range gatewayAppAuthExemptPrefixGETHEADExcept {
			if strings.HasPrefix(path, rule.Prefix) && !strings.HasPrefix(path, rule.ExcludePrefix) {
				return true
			}
		}
	}

	return false
}

func stringInSlice(s string, list []string) bool {
	for _, v := range list {
		if s == v {
			return true
		}
	}
	return false
}
