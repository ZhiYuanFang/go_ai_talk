package cachekit

import (
	"fmt"
	"strconv"
)

const (
	gatewayRefreshKeyPrefix     = "gw:app:rt:"
	gatewayAppVersionLatestBase = "gw:app:version:latest"
	gatewayUsageSimWxSetKey     = "usage:sim_wx_ids"
)

// GatewayRefreshTokenKey App refresh token；TTL 见 gatewayApp.refreshTtlSeconds。
func GatewayRefreshTokenKey(token string) string {
	return gatewayRefreshKeyPrefix + token
}

// GatewayAppVersionLatestBaseKey 版本检查缓存基键（无环境后缀）。
func GatewayAppVersionLatestBaseKey() string {
	return gatewayAppVersionLatestBase
}

// GatewayAppVersionLatestKey 带环境后缀的版本缓存键。
func GatewayAppVersionLatestKey(suffix string) string {
	if suffix == "" {
		return gatewayAppVersionLatestBase
	}
	return gatewayAppVersionLatestBase + ":" + suffix
}

// GatewayUsageDayGlobalKey 全局 API 日计数 Hash；TTL 90 天。
func GatewayUsageDayGlobalKey(day string) string {
	return "gw:usage:d:" + day + ":g"
}

// GatewayUsageDayWxKey 单用户 API 日计数 Hash；TTL 90 天。
func GatewayUsageDayWxKey(day string, wxID int64) string {
	return fmt.Sprintf("gw:usage:d:%s:w:%d", day, wxID)
}

// GatewayUsageDayCrossKey API×wxId 交叉日计数 Hash；TTL 90 天。
func GatewayUsageDayCrossKey(day string) string {
	return "gw:usage:d:" + day + ":x"
}

// GatewayUsageLastGlobalKey 全局 API 最近调用时间 Hash。
func GatewayUsageLastGlobalKey() string {
	return "gw:usage:last:g"
}

// GatewayUsageLastWxKey 单用户 API 最近调用时间 Hash。
func GatewayUsageLastWxKey(wxID int64) string {
	return fmt.Sprintf("gw:usage:last:w:%d", wxID)
}

// GatewayUsageSimWxSetKey 模拟用户 wxId 集合；device 注册 sim 用户时 SADD。
func GatewayUsageSimWxSetKey() string {
	return gatewayUsageSimWxSetKey
}

// GatewayUsageSimWxMember 模拟用户 SET member 字符串形式。
func GatewayUsageSimWxMember(wxID int64) string {
	return strconv.FormatInt(wxID, 10)
}
