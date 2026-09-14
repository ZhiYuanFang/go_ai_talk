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

// —— 客户端功能使用统计（client-usage；仅 Redis，可丢；窗口约 30 天）——

// GatewayFeatUsageDayGlobalKey 全局 feature 日计数 Hash；TTL ≈30 天。
func GatewayFeatUsageDayGlobalKey(day string) string {
	return "gw:featusage:d:" + day + ":g"
}

// GatewayFeatUsageDayWxKey 单用户 feature 日计数 Hash；TTL ≈30 天。
func GatewayFeatUsageDayWxKey(day string, wxID int64) string {
	return fmt.Sprintf("gw:featusage:d:%s:w:%d", day, wxID)
}

// GatewayFeatUsageDayCrossKey feature×wxId 交叉日计数 Hash；TTL ≈30 天。
func GatewayFeatUsageDayCrossKey(day string) string {
	return "gw:featusage:d:" + day + ":x"
}

// GatewayFeatUsageLastGlobalKey 全局 feature 最近上报时间 Hash。
func GatewayFeatUsageLastGlobalKey() string {
	return "gw:featusage:last:g"
}

// GatewayFeatUsageLastWxKey 单用户 feature 最近上报时间 Hash。
func GatewayFeatUsageLastWxKey(wxID int64) string {
	return fmt.Sprintf("gw:featusage:last:w:%d", wxID)
}

// GatewayFeatUsageTimelineKey 单用户时间线 LIST（最新在头）；TTL ≈30 天，条数硬顶 1 万。
func GatewayFeatUsageTimelineKey(wxID int64) string {
	return fmt.Sprintf("gw:featusage:tl:%d", wxID)
}

// GatewayFeatUsageRateLimitKey 同 wx 上报限流键；TTL=3 秒。
func GatewayFeatUsageRateLimitKey(wxID int64) string {
	return fmt.Sprintf("gw:featusage:rl:%d", wxID)
}

// GatewayFeatUsageDescGlobalKey 全局 featureId→最近 description Hash（运维展示；TTL 随写入刷新）。
func GatewayFeatUsageDescGlobalKey() string {
	return "gw:featusage:desc:g"
}
