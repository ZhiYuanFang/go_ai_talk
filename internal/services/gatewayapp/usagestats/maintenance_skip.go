package usagestats

import (
	"strings"
	"unicode"
)

// 维护型 / 探测型 App HTTP API：不计入使用统计（登录、注册、绑定与业务 API 仍记录）。
// 新增排除项时在此维护精确 apiKey 或 path 前缀即可。
// 流程约束：新增 App HTTP 接口 MUST 先向负责人确认是否统计，见 openspec/project.md「App API 使用统计约定」。

var maintenanceExactAPI = map[string]struct{}{
	"POST /device/app/api/token/refresh":    {},
	"GET /device/app/api/version/check":     {},
	"GET /device/app/api/site/home":         {},
	"GET /voice/app/api/ai-quota":           {},
	"GET /ucg/app/api/ai-quota":             {},
	"POST /ucg/app/api/push/register":       {},
	"POST /ucg/app/api/push/unregister":     {},
	"GET /device/history/api/event/options": {},
	"GET /ucg/app/api/conversations":        {},
	"GET /device/app/api/user/get":          {},
	"GET /device/history/api/list":          {},
	// VIP 商品现价/原价：负责人确认匿名读价不计入 usage（vip-price-db）。
	"GET /cash/app/api/vip/product": {},
	// 商业功能：负责人确认查询链路不计入 usage；开通意图 POST 仍统计（commercial-feature-entitlement）。
	"GET /cash/app/api/ucg/eligibility":  {},
	"GET /cash/app/api/feature/catalog": {},
	// clinic/tip 点赞反馈：负责人确认不计入 usage（close-clinic-tip-feedback）；
	// tip generate（POST /device/tip/generate）统计策略属包 B，不得在此排除。
	"POST /device/api/clinic/feedback": {},
	"POST /device/api/tip/feedback":    {},
}

var maintenancePathPrefixes = []string{
	"/device/app/api/version/admin/",
}

// isMaintenanceAPI 判断是否为不应计入统计的维护型接口。
func isMaintenanceAPI(method, path string) bool {
	method = strings.ToUpper(strings.TrimSpace(method))
	path = normalizeUsagePath(path)
	if path == "" {
		return false
	}
	if _, ok := maintenanceExactAPI[method+" "+path]; ok {
		return true
	}
	for _, prefix := range maintenancePathPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	// GET 评论列表：负责人确认进入帖子详情时高频读，不计入 usage 统计。
	// 归一化 apiKey：GET /ucg/app/api/posts/{id}/comments（见 api/v1/ucg_app_http.go UcgPostCommentsGetReq）。
	// POST 同路径（发表评论）仍走上方精确表或默认统计逻辑，不在此排除。
	if isUcgPostCommentsListGET(method, path) {
		return true
	}
	return false
}

// isUcgPostCommentsListGET 匹配 GET /ucg/app/api/posts/<numericPostId>/comments（gateway 原始 path）。
// 不误匹配 mine、单帖 GET /posts/{id}、likes 等同前缀子路径。
func isUcgPostCommentsListGET(method, path string) bool {
	if method != "GET" {
		return false
	}
	const prefix = "/ucg/app/api/posts/"
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	rest := strings.TrimPrefix(path, prefix)
	slash := strings.Index(rest, "/")
	if slash <= 0 {
		return false
	}
	postID := rest[:slash]
	suffix := rest[slash+1:]
	if suffix != "comments" {
		return false
	}
	if len(postID) == 0 {
		return false
	}
	for _, c := range postID {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func normalizeUsagePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if len(p) > 1 && strings.HasSuffix(p, "/") {
		p = strings.TrimRight(p, "/")
	}
	return p
}
