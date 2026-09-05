package cash

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// ActivateFeatureRequest 功能开通原子入参（支付/邀请/广告共用）。
//
// 业务：按主体与通道解析授予效果；一期仅支持 Subject=device。
// ActorWxID 供审计；权益主体为 SubjectKey（device_no）。
type ActivateFeatureRequest struct {
	FeatureID   string
	SubjectType string // device | user
	SubjectKey  string // device_no（device）
	Channel     string // payment | invite_code | ad
	ChannelRef  string
	ActorWxID   int64
	// 以下字段主要由支付通道传入（SKU）；邀请/广告忽略 GrantKind/DurationDays，改读 feature_def。
	GrantKind    string
	GrantQty     int
	DurationDays int
}

// ActivateFeature 共用开通原子入口：写入权益或预测条数并失效缓存。
//
// 效果解析：
//   - payment：grant_kind/quantity/duration 来自入参（SKU）；
//   - invite_code / ad：同源读 feature_def.duration_days（0=永久）；预测强制 allowed_count_delta +1。
//
// Args: req 见 ActivateFeatureRequest。
// Returns: 参数/主体错误或写库错误。
// Side Effects: 写 feature_entitlement 或 feature_allowed_count，Del 设备功能缓存。
func ActivateFeature(ctx context.Context, req ActivateFeatureRequest) error {
	featureID := strings.TrimSpace(req.FeatureID)
	subjectType := strings.TrimSpace(req.SubjectType)
	subjectKey := strings.TrimSpace(req.SubjectKey)
	channel := strings.TrimSpace(req.Channel)
	if featureID == "" || subjectKey == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "featureId/subjectKey 不能为空")
	}
	if subjectType == "" {
		subjectType = ActivationSubjectDevice
	}
	// 一期仅落地设备维权益；user 主体预留枚举。
	if subjectType != ActivationSubjectDevice {
		return gerror.NewCode(gcode.CodeInvalidParameter, "暂不支持账号维功能开通")
	}
	switch channel {
	case UnlockMethodPayment, UnlockMethodInviteCode, UnlockMethodAd:
	default:
		return gerror.NewCode(gcode.CodeInvalidParameter, "未知开通通道")
	}

	grantQty := req.GrantQty
	if grantQty <= 0 {
		grantQty = 1
	}
	grantKind := strings.TrimSpace(req.GrantKind)
	durationDays := req.DurationDays

	// 预测：三通道均为永久条数增量；忽略定义天数。
	if featureID == FeatureIDPredictionUnlock {
		return GrantEntitlementOrCount(ctx, subjectKey, featureID, channel, GrantKindAllowedCountDelta, grantQty, 0, req.ChannelRef)
	}

	// 邀请/广告：权益型天数与广告同源，取自功能定义。
	if channel == UnlockMethodInviteCode || channel == UnlockMethodAd {
		var def struct {
			FeatureId    string `json:"feature_id"`
			DurationDays int    `json:"duration_days"`
			Status       int    `json:"status"`
		}
		_ = g.DB().Model("feature_def").Ctx(ctx).Where("feature_id", featureID).Scan(&def)
		if def.FeatureId == "" || def.Status != 1 {
			return gerror.NewCode(gcode.CodeInvalidParameter, "功能不存在或已停用")
		}
		durationDays = def.DurationDays
		grantKind = GrantKindEntitlement
	}
	if grantKind == "" {
		grantKind = GrantKindEntitlement
	}
	return GrantEntitlementOrCount(ctx, subjectKey, featureID, channel, grantKind, grantQty, durationDays, req.ChannelRef)
}

// HasActiveFeatureEntitlement 设备某功能权益是否未过期（expires_at=0 为永久）。
func HasActiveFeatureEntitlement(ctx context.Context, deviceNo, featureID string) (active bool, expiresAt int64, err error) {
	deviceNo = strings.TrimSpace(deviceNo)
	featureID = strings.TrimSpace(featureID)
	if deviceNo == "" || featureID == "" {
		return false, 0, nil
	}
	r, err := g.DB().Model("feature_entitlement").Ctx(ctx).
		Fields("expires_at").
		Where("device_no", deviceNo).Where("feature_id", featureID).
		One()
	if err != nil {
		return false, 0, err
	}
	if r.IsEmpty() {
		return false, 0, nil
	}
	exp := r["expires_at"].Int64()
	if exp == 0 {
		return true, 0, nil
	}
	if exp > time.Now().Unix() {
		return true, exp, nil
	}
	return false, exp, nil
}
