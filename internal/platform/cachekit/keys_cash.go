package cachekit

import (
	"fmt"
)

// —— cash 商业功能 / UCG 资格缓存键（跨 gateway 与 cash-service 共享语义时须走本 builder）——

// CashUCGEligibilityKey 兼容旧 UCG 资格键（新路径请用 CashFeedingEligibilityKey）。
func CashUCGEligibilityKey(deviceNo, yyyyMMdd string) string {
	return fmt.Sprintf("cash:ucg:eligibility:%s:%s", deviceNo, yyyyMMdd)
}

// CashFeedingEligibilityKey 喂养资格按场景+设备+上海日+配置版本；TTL 建议 36h；不落 MySQL。
// cfgUpdatedAt 变更后旧键自然失效，避免 Admin 改阈值后脏读。
func CashFeedingEligibilityKey(sceneKey, deviceNo, yyyyMMdd string, cfgUpdatedAt int64) string {
	return fmt.Sprintf("cash:feeding:elig:%s:%s:%s:%d", sceneKey, deviceNo, yyyyMMdd, cfgUpdatedAt)
}

// CashFeedingEligibilitySceneKey 场景阈值热读；Admin 写 MUST DEL；TTL 建议 5–15min。
func CashFeedingEligibilitySceneKey(sceneKey string) string {
	return fmt.Sprintf("cash:feeding:elig:scene:%s", sceneKey)
}

// CashFeatureDefCatalogKey 全站启用功能定义字典热读；Admin 写路径 MUST DEL；TTL 建议 5–15min。
func CashFeatureDefCatalogKey() string {
	return "cash:feature:def:catalog"
}

// CashFeatureAllowedCountKey 设备预测开通状态热读（JSON：永久累加+临时全开）；履约/Admin 变更 MUST 失效；权威在 MySQL。
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
