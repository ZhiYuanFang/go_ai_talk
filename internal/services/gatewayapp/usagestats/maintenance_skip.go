package usagestats

import (
	"strings"
)

// 维护型 / 探测型 App HTTP API：不计入使用统计（登录、注册、绑定与业务 API 仍记录）。
// 新增排除项时在此维护精确 apiKey 或 path 前缀即可。

var maintenanceExactAPI = map[string]struct{}{
	"POST /device/app/api/token/refresh": {},
	"GET /device/app/api/version/check":  {},
	"GET /device/app/api/site/home":      {},
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
	return false
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
