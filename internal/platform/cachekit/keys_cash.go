package cachekit

import (
	"fmt"
)

// —— cash 商业功能 / UCG 资格缓存键（跨 gateway 与 cash-service 共享语义时须走本 builder）——

// CashUCGEligibilityKey UCG 入场资格按日结果；TTL 建议至当日结束+缓冲或固定 36h；不落 MySQL。
// 同 device+上海日重复请求 MUST 命中缓存；跨日换日期段自动隔离。
func CashUCGEligibilityKey(deviceNo, yyyyMMdd string) string {
	return fmt.Sprintf("cash:ucg:eligibility:%s:%s", deviceNo, yyyyMMdd)
}

// CashFeatureDefCatalogKey 全站启用功能定义字典热读；Admin 写路径 MUST DEL；TTL 建议 5–15min。
func CashFeatureDefCatalogKey() string {
	return "cash:feature:def:catalog"
}

// CashFeatureAllowedCountKey 设备预测开通数量热读；履约/Admin 变更 MUST 失效或写穿；权威在 MySQL。
func CashFeatureAllowedCountKey(deviceNo string) string {
	return fmt.Sprintf("cash:feature:allowed:%s", deviceNo)
}

// CashFeatureCatalogDeviceKey 可选：合成目录 per-device 短缓存；履约后 MUST DEL；可不使用（请求路径 JOIN）。
func CashFeatureCatalogDeviceKey(deviceNo string) string {
	return fmt.Sprintf("cash:feature:catalog:%s", deviceNo)
}

// CashFeatureAdIdempotencyKey 广告完成短窗幂等；TTL 建议数分钟级。
func CashFeatureAdIdempotencyKey(deviceNo, featureID, clientKey string) string {
	return fmt.Sprintf("cash:feature:ad:idem:%s:%s:%s", deviceNo, featureID, clientKey)
}

// CashFeatureAdDailyLimitKey 广告开通设备日限额计数；TTL 至上海日结束或 36h。
func CashFeatureAdDailyLimitKey(deviceNo, yyyyMMdd string) string {
	return fmt.Sprintf("cash:feature:ad:daily:%s:%s", deviceNo, yyyyMMdd)
}
