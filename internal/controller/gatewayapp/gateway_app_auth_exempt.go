package gatewayappctrl

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
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
		"/device/app/api/username_login",
		"/device/app/api/apple_login",
		"/device/app/api/device_login",
		"/device/app/api/token/refresh",
		"/device/app/api/user/login",
		"/device/app/api/user/apple/login",
		"/device/app/api/user/username/register",
		"/device/app/api/user/username/login",
		"/device/app/api/user/device_login",
		"/device/admin/api/login",
		// 支付宝异步通知：匿名到达，cash-service 内验签。
		"/cash/app/api/vip/alipay/notify",
	}

	// GET 且路径精确匹配（WebSocket Upgrade 等不要求 HTTP 层 Bearer）。
	gatewayAppAuthExemptExactGET = []string{
		"/device/app/ws/history",
		"/voice/chat/ws",
		"/voice/asr/ws",
		"/voice/clinic/ws",
		"/ucg/app/ws/chat",
	}

	// GET 或 HEAD：路径精确匹配。
	gatewayAppAuthExemptExactGETHEAD = []string{
		"/",
		"/apple-app-site-association",
		"/.well-known/apple-app-site-association",
		"/api.json",
		"/device/admin",
		"/device/admin/",
		"/device/admin/qa-records",
		"/device/admin/feedback-records",
		"/device/admin/api-usage-stats",
		"/device/app/api/site/home",
		"/device/app/api/version/check",
		"/device/history/api/event/options",
		"/ucg/app/api/feed/recommend",
		"/ucg/app/api/health",
		// VIP 商品现价/原价：落地页未登录可读；建单/支付仍须登录。
		"/cash/app/api/vip/product",
		"/device/app/version-admin.html",
		"/device/admin/ucg-admin.html",
		"/device/admin/voice-admin.html",
		"/device/admin/ai-model-admin.html",
		"/device/admin/sim-admin.html",
		"/device/admin/cash-vip-admin.html",
		"/device/admin/cash-feature-admin.html",
		"/device/admin/cash-feeding-eligibility-admin.html",
		"/user-agreement.html",
		"/privacy-policy.html",
		"/favicon.ico",
		"/robots.txt",
	}

	// GET 或 HEAD：路径以前缀命中即豁免。
	gatewayAppAuthExemptPrefixesGETHEAD = []string{
		"/wx/ulink/",
		"/resource/",
		"/device/app/apk/",
	}
)

// gatewayAppAuthPrefixExcept 表示「以前缀放行但排除更具体前缀」的壳页类路由。
type gatewayAppAuthPrefixExcept struct {
	Prefix        string
	ExcludePrefix string
}

var gatewayAppAuthExemptPrefixGETHEADExcept = []gatewayAppAuthPrefixExcept{
	{Prefix: "/device/admin/history/", ExcludePrefix: "/device/admin/api/"},
	{Prefix: "/ucg/app/api/profile/", ExcludePrefix: "/ucg/app/api/profile/me"},
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
		if gatewayAppGetPostByIDExempt(path, method) {
			return true
		}
		if gatewayAppGetPostsUserExempt(path, method) {
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

func gatewayAppGetPostByIDExempt(path, method string) bool {
	if method != http.MethodGet && method != http.MethodHead {
		return false
	}
	const prefix = "/ucg/app/api/posts/"
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	rest := strings.TrimPrefix(path, prefix)
	if rest == "" || rest == "mine" {
		return false
	}
	if strings.Contains(rest, "/") {
		return false
	}
	for _, c := range rest {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(rest) > 0
}

func gatewayAppGetPostsUserExempt(path, method string) bool {
	if method != http.MethodGet && method != http.MethodHead {
		return false
	}
	const prefix = "/ucg/app/api/posts/user/"
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	rest := strings.TrimPrefix(path, prefix)
	return rest != "" && !strings.Contains(rest, "/")
}

func stringInSlice(s string, list []string) bool {
	for _, v := range list {
		if s == v {
			return true
		}
	}
	return false
}

func WriteAuthJSON(r *ghttp.Request, status int, message string) {
	r.Response.Status = status
	r.Response.WriteJson(g.Map{"code": status, "message": message})
	r.ExitAll()
}
