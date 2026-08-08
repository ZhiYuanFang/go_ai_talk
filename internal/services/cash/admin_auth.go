// 管理端鉴权：校验网关注入的 X-Admin-Password。
//
// 业务说明：
// Hub 浏览器仅携带 Admin JWT；gateway-app 在反代 /cash/admin/api/* 时注入口令头。
// cash-service 本文件负责与注入值做常量时间比较，防止时序旁路。
package cash

import (
	"os"
	"strings"

	"hello/internal/services/gatewayapp"
)

// HeaderAdminPassword 与 gateway-app InjectAdminDownstreamPassword 写入的头一致。
const HeaderAdminPassword = "X-Admin-Password"

// CashAdminPassword 读取 cash Admin 口令。
//
// 优先级：
//  1. 环境变量 CASH_ADMIN_PASSWORD（非空则用之，便于与 Hub 口令分离）；
//  2. 否则回退 GATEWAY_APP_ADMIN_PASSWORD（与 Voice/Sim 网关注入回退语义对齐）。
//
// Returns: 期望口令；二者皆空时返回空串（校验必失败）。
func CashAdminPassword() string {
	if v := strings.TrimSpace(os.Getenv("CASH_ADMIN_PASSWORD")); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("GATEWAY_APP_ADMIN_PASSWORD"))
}

// VerifyCashAdminPassword 校验请求携带的管理口令。
//
// Args:
//   - password: 通常来自请求头 X-Admin-Password（由网关注入，浏览器不得自带）。
//
// Returns: true 表示口令匹配且期望口令已配置；未配置或不等则 false。
func VerifyCashAdminPassword(password string) bool {
	expected := CashAdminPassword()
	if expected == "" {
		return false
	}
	return gatewayapp.ConstantTimeEqual(strings.TrimSpace(password), expected)
}
