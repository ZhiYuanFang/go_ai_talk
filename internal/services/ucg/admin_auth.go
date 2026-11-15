package ucg

import (
	"context"
	"os"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// UcgAdminPassword 读取 Admin 口令：优先 env UCG_ADMIN_PASSWORD，否则 yaml ucg.admin.password。
// 业务说明：UCG Admin HTTP（审帖/资料机审等）依赖 X-Admin-Password；gateway-app 在校验 Admin JWT 后注入该头。
func UcgAdminPassword(ctx context.Context) string {
	if v := strings.TrimSpace(os.Getenv("UCG_ADMIN_PASSWORD")); v != "" {
		return v
	}
	return strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.admin.password").String())
}

// VerifyUcgAdminPassword 校验请求携带的 Admin 口令是否与配置一致。
// Args: password — 通常来自 Header X-Admin-Password。
// Returns: true 表示口令非空且匹配。
func VerifyUcgAdminPassword(ctx context.Context, password string) bool {
	expected := UcgAdminPassword(ctx)
	if expected == "" {
		return false
	}
	return strings.TrimSpace(password) == expected
}
