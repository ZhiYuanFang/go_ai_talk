package cash

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// DeviceAllowedCountState 设备预测数量权威快照（永久累加 + 临时全开）。
type DeviceAllowedCountState struct {
	PermanentDelta        int   `json:"permanentDelta"`
	FullAccess            int   `json:"fullAccess"`
	FullAccessExpiresAt   int64 `json:"fullAccessExpiresAt"`
}

// HasActiveFullAccess 是否处于临时/永久全开（full_access=1 且未过期）。
func (s DeviceAllowedCountState) HasActiveFullAccess(now int64) bool {
	if s.FullAccess != 1 {
		return false
	}
	if s.FullAccessExpiresAt == 0 {
		return true // 永久全开
	}
	return s.FullAccessExpiresAt > now
}

// GrantEntitlementOrCount 向 device 授予权益或累加永久 allowedCount，并失效相关缓存。
//
// 业务：支付/广告履约统一入口。邀请码预测临时全开请用 GrantPredictionFullAccess。
// Args: grantKind=entitlement|allowed_count_delta；durationDays=0 表示永久（仅权益类）。
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

	// 预测数量增量：权威写 feature_allowed_count.allowed_count（永久累加，不碰临时全开）。
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

// GrantPredictionFullAccess 邀请码等：为设备写入预测临时/永久全开（T1），不增加永久条数。
//
// Args: durationDays>0 有限期；=0 表示永久全开（full_access=1 且 expires_at=0）。
func GrantPredictionFullAccess(ctx context.Context, deviceNo string, durationDays int, sourceRef string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	now := time.Now().Unix()
	db := g.DB()
	r, err := db.Model("feature_allowed_count").Ctx(ctx).Where("device_no", deviceNo).One()
	if err != nil {
		return err
	}

	var newExp int64
	if durationDays <= 0 {
		newExp = 0 // 永久全开
	} else if r.IsEmpty() {
		newExp = now + int64(durationDays)*86400
	} else {
		base := now
		curExp := r["full_access_expires_at"].Int64()
		curFA := r["full_access"].Int()
		// 已永久全开则保持永久。
		if curFA == 1 && curExp == 0 {
			newExp = 0
		} else {
			if curFA == 1 && curExp > base {
				base = curExp
			}
			newExp = base + int64(durationDays)*86400
		}
	}

	if r.IsEmpty() {
		_, err = db.Model("feature_allowed_count").Ctx(ctx).Data(g.Map{
			"device_no":              deviceNo,
			"allowed_count":          0,
			"full_access":            1,
			"full_access_expires_at": newExp,
			"updated_at":             now,
		}).Insert()
	} else {
		_, err = db.Model("feature_allowed_count").Ctx(ctx).Where("device_no", deviceNo).Data(g.Map{
			"full_access":            1,
			"full_access_expires_at": newExp,
			"updated_at":             now,
		}).Update()
	}
	if err != nil {
		return err
	}
	_ = sourceRef // 流水在邀请兑换侧记录
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
			"device_no":              deviceNo,
			"allowed_count":          delta,
			"full_access":            0,
			"full_access_expires_at": 0,
			"updated_at":             now,
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

// GetDeviceAllowedCountState 读设备预测数量状态（权威 MySQL，可选 Redis JSON）。
func GetDeviceAllowedCountState(ctx context.Context, deviceNo string) (DeviceAllowedCountState, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	var empty DeviceAllowedCountState
	if deviceNo == "" {
		return empty, nil
	}
	c := cachekit.Default()
	key := cachekit.CashFeatureAllowedCountKey(deviceNo)
	if raw, ok, err := c.Get(ctx, key); err == nil && ok && raw != "" {
		var st DeviceAllowedCountState
		if json.Unmarshal([]byte(raw), &st) == nil {
			return st, nil
		}
		// 兼容旧缓存：纯整数 = 永久累加。
		var n int
		if _, e := fmt.Sscanf(raw, "%d", &n); e == nil {
			return DeviceAllowedCountState{PermanentDelta: n}, nil
		}
	}
	r, err := g.DB().Model("feature_allowed_count").Ctx(ctx).Where("device_no", deviceNo).One()
	if err != nil {
		return empty, err
	}
	st := DeviceAllowedCountState{}
	if !r.IsEmpty() {
		st.PermanentDelta = r["allowed_count"].Int()
		st.FullAccess = r["full_access"].Int()
		st.FullAccessExpiresAt = r["full_access_expires_at"].Int64()
	}
	if b, mErr := json.Marshal(st); mErr == nil {
		_ = c.SetEX(ctx, key, string(b), 15*time.Minute)
	}
	return st, nil
}

// EffectivePredictionAllowedCount 合成有效条数：临时全开 → -1；否则 defaultFree+permanentDelta。
//
// Returns: allowedCount（含哨兵）、expiresAt（有限期全开时）、unlocked。
func EffectivePredictionAllowedCount(defaultFree int, st DeviceAllowedCountState, now int64) (allowedCount int, expiresAt int64, unlocked bool) {
	if defaultFree < 0 {
		defaultFree = 0
	}
	if st.HasActiveFullAccess(now) {
		exp := st.FullAccessExpiresAt
		return AllowedCountFullAccessSentinel, exp, true
	}
	n := defaultFree + st.PermanentDelta
	if n < 0 {
		n = 0
	}
	return n, 0, n > 0
}

// GetAllowedCount 兼容旧调用：仅返回永久累加（不含默认与全开）。新路径请用 GetDeviceAllowedCountState。
func GetAllowedCount(ctx context.Context, deviceNo string) (int, error) {
	st, err := GetDeviceAllowedCountState(ctx, deviceNo)
	if err != nil {
		return 0, err
	}
	return st.PermanentDelta, nil
}
