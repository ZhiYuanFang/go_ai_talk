package cash

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// GrantEntitlementOrCount 向 device 授予权益或累加 allowedCount，并失效相关缓存。
//
// 业务：支付/邀请码/广告履约统一入口。
// Args: grantKind=entitlement|allowed_count_delta；durationDays=0 表示永久。
func GrantEntitlementOrCount(ctx context.Context, deviceNo, featureID, unlockMethod, grantKind string, grantQty, durationDays int, sourceRef string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	featureID = strings.TrimSpace(featureID)
	grantKind = strings.TrimSpace(grantKind)
	if deviceNo == "" || featureID == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo/featureId 不能为空")
	}
	if grantQty <= 0 {
		grantQty = 1
	}
	now := time.Now().Unix()

	// 预测数量增量：权威写 feature_allowed_count。
	if grantKind == GrantKindAllowedCountDelta || featureID == FeatureIDPredictionUnlock && grantKind == "" {
		if grantKind == "" {
			grantKind = GrantKindAllowedCountDelta
		}
	}
	if grantKind == GrantKindAllowedCountDelta {
		if err := incrAllowedCount(ctx, deviceNo, grantQty); err != nil {
			return err
		}
		invalidateDeviceFeatureCaches(ctx, deviceNo)
		return nil
	}

	if grantKind == "" {
		grantKind = GrantKindEntitlement
	}
	if grantKind != GrantKindEntitlement {
		return gerror.NewCode(gcode.CodeInvalidParameter, "未知 grant_kind")
	}
	if err := upsertFeatureEntitlement(ctx, deviceNo, featureID, unlockMethod, durationDays, grantQty, sourceRef, now); err != nil {
		return err
	}
	invalidateDeviceFeatureCaches(ctx, deviceNo)
	return nil
}

func incrAllowedCount(ctx context.Context, deviceNo string, delta int) error {
	now := time.Now().Unix()
	db := g.DB()
	r, err := db.Model("feature_allowed_count").Ctx(ctx).Where("device_no", deviceNo).One()
	if err != nil {
		return err
	}
	if r.IsEmpty() {
		_, err = db.Model("feature_allowed_count").Ctx(ctx).Data(g.Map{
			"device_no":     deviceNo,
			"allowed_count": delta,
			"updated_at":    now,
		}).Insert()
		return err
	}
	cur := r["allowed_count"].Int()
	_, err = db.Model("feature_allowed_count").Ctx(ctx).Where("device_no", deviceNo).Data(g.Map{
		"allowed_count": cur + delta,
		"updated_at":    now,
	}).Update()
	return err
}

func upsertFeatureEntitlement(ctx context.Context, deviceNo, featureID, unlockMethod string, durationDays, quantity int, sourceRef string, now int64) error {
	db := g.DB()
	exist, err := db.Model("feature_entitlement").Ctx(ctx).
		Where("device_no", deviceNo).Where("feature_id", featureID).One()
	if err != nil {
		return err
	}
	var newExp int64
	if durationDays <= 0 {
		newExp = 0
	} else if exist.IsEmpty() {
		newExp = now + int64(durationDays)*86400
	} else {
		base := now
		if cur := exist["expires_at"].Int64(); cur > base {
			base = cur
		}
		// 已永久则保持永久。
		if exist["expires_at"].Int64() == 0 && exist["updated_at"].Int64() > 0 {
			newExp = 0
		} else {
			newExp = base + int64(durationDays)*86400
		}
	}
	if exist.IsEmpty() {
		_, err = db.Model("feature_entitlement").Ctx(ctx).Data(g.Map{
			"device_no":     deviceNo,
			"feature_id":    featureID,
			"unlock_method": unlockMethod,
			"expires_at":    newExp,
			"quantity":      quantity,
			"source_ref":    sourceRef,
			"created_at":    now,
			"updated_at":    now,
		}).Insert()
		return err
	}
	_, err = db.Model("feature_entitlement").Ctx(ctx).
		Where("device_no", deviceNo).Where("feature_id", featureID).
		Data(g.Map{
			"unlock_method": unlockMethod,
			"expires_at":    newExp,
			"quantity":      quantity,
			"source_ref":    sourceRef,
			"updated_at":    now,
		}).Update()
	return err
}

func invalidateDeviceFeatureCaches(ctx context.Context, deviceNo string) {
	c := cachekit.Default()
	_ = c.Del(ctx, cachekit.CashFeatureAllowedCountKey(deviceNo))
	_ = c.Del(ctx, cachekit.CashFeatureCatalogDeviceKey(deviceNo))
}

func invalidateFeatureDefCache(ctx context.Context) {
	_ = cachekit.Default().Del(ctx, cachekit.CashFeatureDefCatalogKey())
}

// GetAllowedCount 读设备预测开通数量（权威 MySQL，可选 Redis）。
func GetAllowedCount(ctx context.Context, deviceNo string) (int, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return 0, nil
	}
	c := cachekit.Default()
	key := cachekit.CashFeatureAllowedCountKey(deviceNo)
	if raw, ok, err := c.Get(ctx, key); err == nil && ok {
		var n int
		if _, e := fmt.Sscanf(raw, "%d", &n); e == nil {
			return n, nil
		}
	}
	r, err := g.DB().Model("feature_allowed_count").Ctx(ctx).Where("device_no", deviceNo).One()
	if err != nil {
		return 0, err
	}
	n := 0
	if !r.IsEmpty() {
		n = r["allowed_count"].Int()
	}
	_ = c.SetEX(ctx, key, fmt.Sprintf("%d", n), 15*time.Minute)
	return n, nil
}
