package cash

import (
	"context"
	"strings"
	"time"

	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// GrantEntitlementOrCount 向 device 授予权益并失效相关缓存。
//
// 业务：支付履约写 feature_entitlement。预测条数增量已下线。
// Args: grantKind 须为 entitlement；durationDays 须≥1，不支持永久。
func GrantEntitlementOrCount(ctx context.Context, deviceNo, featureID, unlockMethod, grantKind string, grantQty, durationDays int, sourceRef string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	featureID = strings.TrimSpace(featureID)
	grantKind = strings.TrimSpace(grantKind)
	if deviceNo == "" || featureID == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo/featureId 不能为空")
	}
	if featureID == FeatureIDPredictionUnlock || grantKind == GrantKindAllowedCountDelta {
		glog.Warningf(ctx, "[cash] prediction_unlock/allowed_count_delta 已下线 featureId=%s kind=%s", featureID, grantKind)
		return gerror.NewCode(gcode.CodeInvalidOperation, "预测事项开通数量已下线")
	}
	if grantQty <= 0 {
		grantQty = 1
	}
	now := time.Now().Unix()
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

func upsertFeatureEntitlement(ctx context.Context, deviceNo, featureID, unlockMethod string, durationDays, quantity int, sourceRef string, now int64) error {
	if durationDays < 1 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "授予天数须≥1，不支持永久")
	}
	db := g.DB()
	exist, err := db.Model("feature_entitlement").Ctx(ctx).
		Where("device_no", deviceNo).Where("feature_id", featureID).One()
	if err != nil {
		return err
	}
	var newExp int64
	if exist.IsEmpty() {
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

// GrantUserEntitlement 向 wx 授予/续期功能权益（账号维）。
//
// Args: durationDays 须≥1。
// Side Effects: 写 feature_user_entitlement；失效功能定义缓存。
func GrantUserEntitlement(ctx context.Context, wxID int64, featureID, unlockMethod string, grantQty, durationDays int, sourceRef string) error {
	if durationDays < 1 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "授予天数须≥1，不支持永久")
	}
	addSec := int64(durationDays) * 86400
	return grantUserEntitlementAddSeconds(ctx, wxID, featureID, unlockMethod, grantQty, addSec, sourceRef)
}

// GrantUserEntitlementHours 向 wx 授予/续期功能权益（按小时；试用 24h）。
func GrantUserEntitlementHours(ctx context.Context, wxID int64, featureID, unlockMethod string, grantQty, durationHours int, sourceRef string) error {
	if durationHours <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "durationHours 须为正")
	}
	return grantUserEntitlementAddSeconds(ctx, wxID, featureID, unlockMethod, grantQty, int64(durationHours)*3600, sourceRef)
}

func grantUserEntitlementAddSeconds(ctx context.Context, wxID int64, featureID, unlockMethod string, grantQty int, addSeconds int64, sourceRef string) error {
	featureID = strings.TrimSpace(featureID)
	unlockMethod = strings.TrimSpace(unlockMethod)
	if wxID <= 0 || featureID == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId/featureId 不能为空")
	}
	if grantQty <= 0 {
		grantQty = 1
	}
	now := time.Now().Unix()
	if err := upsertFeatureUserEntitlement(ctx, wxID, featureID, unlockMethod, addSeconds, grantQty, sourceRef, now); err != nil {
		return err
	}
	invalidateFeatureDefCache(ctx)
	return nil
}

// upsertFeatureUserEntitlement addSeconds 须为正；从 now 或未过期 expires_at 起叠加。已有永久行保持永久。
func upsertFeatureUserEntitlement(ctx context.Context, wxID int64, featureID, unlockMethod string, addSeconds int64, quantity int, sourceRef string, now int64) error {
	db := g.DB()
	exist, err := db.Model("feature_user_entitlement").Ctx(ctx).
		Where("wx_id", wxID).Where("feature_id", featureID).One()
	if err != nil {
		return err
	}
	var newExp int64
	if addSeconds <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "授予时长须为正，不支持永久")
	}
	if exist.IsEmpty() {
		newExp = now + addSeconds
	} else {
		base := now
		if cur := exist["expires_at"].Int64(); cur > base {
			base = cur
		}
		if exist["expires_at"].Int64() == 0 && exist["updated_at"].Int64() > 0 {
			newExp = 0 // 已永久不缩短
		} else {
			newExp = base + addSeconds
		}
	}
	if exist.IsEmpty() {
		_, err = db.Model("feature_user_entitlement").Ctx(ctx).Data(g.Map{
			"wx_id": wxID, "feature_id": featureID, "unlock_method": unlockMethod,
			"expires_at": newExp, "quantity": quantity, "source_ref": sourceRef,
			"created_at": now, "updated_at": now,
		}).Insert()
		return err
	}
	_, err = db.Model("feature_user_entitlement").Ctx(ctx).
		Where("wx_id", wxID).Where("feature_id", featureID).
		Data(g.Map{
			"unlock_method": unlockMethod, "expires_at": newExp,
			"quantity": quantity, "source_ref": sourceRef, "updated_at": now,
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

