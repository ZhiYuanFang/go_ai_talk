// 管理端鉴权：校验网关注入的 X-Admin-Password。
//
// 业务说明：
// Hub 浏览器仅携带 Admin JWT；gateway-app 在反代 /push/admin/api/* 时注入口令头。
// push-service 本文件与注入值做常量时间比较。
package push

import (
	"os"
	"strings"

	"hello/internal/platform/httpmeta"
)

// HeaderAdminPassword 与 gateway-app InjectAdminDownstreamPassword 写入的头一致。
const HeaderAdminPassword = "X-Admin-Password"

// PushAdminPassword 读取 push Admin 口令。
//
// 优先级：
//  1. PUSH_ADMIN_PASSWORD（非空则用之）；
//  2. DEVICE_ADMIN_PASSWORD；
//  3. GATEWAY_APP_ADMIN_PASSWORD（与网关注入回退对齐）。
func PushAdminPassword() string {
	if v := strings.TrimSpace(os.Getenv("PUSH_ADMIN_PASSWORD")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("DEVICE_ADMIN_PASSWORD")); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("GATEWAY_APP_ADMIN_PASSWORD"))
}

// VerifyPushAdminPassword 校验请求携带的管理口令。
func VerifyPushAdminPassword(password string) bool {
	expected := PushAdminPassword()
	if expected == "" {
		return false
	}
	return httpmeta.ConstantTimeEqual(strings.TrimSpace(password), expected)
}
