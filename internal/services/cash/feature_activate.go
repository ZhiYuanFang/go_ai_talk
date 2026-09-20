package cash

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// ActivateFeatureRequest 功能开通原子入参（支付/邀请/试用共用；广告通道保留兼容但种子已剥离）。
//
// 业务：按主体与通道解析授予效果；支持 Subject=device|user（由 feature_def.activation_subject 驱动）。
// ActorWxID 供审计；权益主体为 SubjectKey（device_no 或 wx_id 字符串）。
type ActivateFeatureRequest struct {
	FeatureID   string
	SubjectType string // device | user
	SubjectKey  string // device_no（device）或 wxId 十进制（user）
	Channel     string // payment | invite_code | ad | trial
	ChannelRef  string
	ActorWxID   int64
	// 以下字段主要由支付通道传入（SKU）；邀请忽略 GrantKind/DurationDays，改读 feature_def 分列天数。
	GrantKind    string
	GrantQty     int
	DurationDays int
	// DurationHours 试用等短时授予；>0 时优先于 DurationDays（秒级 = hours*3600）。
	DurationHours int
}

// ActivateFeature 共用开通原子入口：写入设备/账号权益或预测条数并失效缓存。
//
// 效果解析：
//   - payment：grant_kind/quantity/duration 来自入参（SKU）；
//   - invite_code：读 feature_def.invite_duration_days，须≥1；
//   - ad：读 feature_def.ad_duration_days，须≥1；种子已剥离，旧客户端调用将因 unlock_methods 拒绝；
//   - trial：DurationHours（默认 TrialDurationHours），仅账号维权益。
//
// Args: req 见 ActivateFeatureRequest。
// Returns: 参数/主体错误或写库错误。
// Side Effects: 写 feature_entitlement / feature_user_entitlement 或 feature_allowed_count，Del 相关缓存。
func ActivateFeature(ctx context.Context, req ActivateFeatureRequest) error {
	featureID := strings.TrimSpace(req.FeatureID)
	subjectType := NormalizeActivationSubject(req.SubjectType)
	subjectKey := strings.TrimSpace(req.SubjectKey)
	channel := strings.TrimSpace(req.Channel)
	if featureID == "" || subjectKey == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "featureId/subjectKey 不能为空")
	}
	switch channel {
	case UnlockMethodPayment, UnlockMethodInviteCode, UnlockMethodAd, UnlockMethodTrial, UnlockMethodAdmin:
	default:
		return gerror.NewCode(gcode.CodeInvalidParameter, "未知开通通道")
	}

	grantQty := req.GrantQty
	if grantQty <= 0 {
		grantQty = 1
	}
	grantKind := strings.TrimSpace(req.GrantKind)
	durationDays := req.DurationDays
	durationHours := req.DurationHours

	// 预测条数已下线：拒绝并打日志，不当成普通权益。
	if featureID == FeatureIDPredictionUnlock {
		glog.Warningf(ctx, "[cash] prediction_unlock 已下线，拒绝开通 featureId=%s channel=%s", featureID, channel)
		return gerror.NewCode(gcode.CodeInvalidOperation, "预测事项开通数量已下线")
	}

	// 试用：固定短时账号权益。
	if channel == UnlockMethodTrial {
		if subjectType != ActivationSubjectUser {
			return gerror.NewCode(gcode.CodeInvalidParameter, "试用仅支持账号维开通")
		}
		if durationHours <= 0 {
			durationHours = TrialDurationHours
		}
		wxID, err := strconv.ParseInt(subjectKey, 10, 64)
		if err != nil || wxID <= 0 {
			return gerror.NewCode(gcode.CodeInvalidParameter, "账号维开通须提供有效 wxId")
		}
		return GrantUserEntitlementHours(ctx, wxID, featureID, channel, grantQty, durationHours, req.ChannelRef)
	}

	// 邀请/广告：权益型天数分列读取。
	if channel == UnlockMethodInviteCode || channel == UnlockMethodAd {
		var def struct {
			FeatureId          string `json:"feature_id"`
			InviteDurationDays int    `json:"invite_duration_days"`
			AdDurationDays     int    `json:"ad_duration_days"`
			Status             int    `json:"status"`
		}
		// 功能定义可能不存在：One+IsEmpty，禁止 Scan 空集 ErrNoRows 被吞掉
		one, qErr := g.DB().Model("feature_def").Ctx(ctx).Where("feature_id", featureID).One()
		if qErr != nil {
			return qErr
		}
		if one.IsEmpty() {
			return gerror.NewCode(gcode.CodeInvalidParameter, "功能不存在或已停用")
		}
		if err := one.Struct(&def); err != nil {
			return err
		}
		if def.FeatureId == "" || def.Status != 1 {
			return gerror.NewCode(gcode.CodeInvalidParameter, "功能不存在或已停用")
		}
		if channel == UnlockMethodInviteCode {
			durationDays = def.InviteDurationDays
		} else {
			durationDays = def.AdDurationDays
		}
		grantKind = GrantKindEntitlement
	}
	if grantKind == "" {
		grantKind = GrantKindEntitlement
	}
	if grantKind == GrantKindAllowedCountDelta {
		glog.Warningf(ctx, "[cash] allowed_count_delta 已下线 featureId=%s", featureID)
		return gerror.NewCode(gcode.CodeInvalidOperation, "预测条数开通已下线")
	}
	// 支付/邀请/广告都不支持永久：天数须≥1，拒绝写成 expires_at=0。
	if durationDays < 1 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "授予天数须≥1，不支持永久")
	}

	if subjectType == ActivationSubjectUser {
		wxID, err := strconv.ParseInt(subjectKey, 10, 64)
		if err != nil || wxID <= 0 {
			return gerror.NewCode(gcode.CodeInvalidParameter, "账号维开通须提供有效 wxId")
		}
		return GrantUserEntitlement(ctx, wxID, featureID, channel, grantQty, durationDays, req.ChannelRef)
	}
	return GrantEntitlementOrCount(ctx, subjectKey, featureID, channel, grantKind, grantQty, durationDays, req.ChannelRef)
}

// GetFeatureActivationSubject 读取功能开通主体（缺省 device）。
func GetFeatureActivationSubject(ctx context.Context, featureID string) (string, error) {
	featureID = strings.TrimSpace(featureID)
	if featureID == "" {
		return ActivationSubjectDevice, nil
	}
	r, err := g.DB().Model("feature_def").Ctx(ctx).Fields("activation_subject").Where("feature_id", featureID).One()
	if err != nil {
		return "", err
	}
	if r.IsEmpty() {
		return ActivationSubjectDevice, nil
	}
	return NormalizeActivationSubject(r["activation_subject"].String()), nil
}

// ResolveActivateSubject 按功能定义解析支付/邀请/试用应使用的 SubjectType 与 SubjectKey。
//
// Args: featureID；deviceNo 与 wxID 由通道提供（user 主体必须 wxID>0）。
// Returns: subjectType、subjectKey、错误。
func ResolveActivateSubject(ctx context.Context, featureID, deviceNo string, wxID int64) (subjectType, subjectKey string, err error) {
	subj, err := GetFeatureActivationSubject(ctx, featureID)
	if err != nil {
		return "", "", err
	}
	if subj == ActivationSubjectUser {
		if wxID <= 0 {
			return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "该功能按账号开通，须登录")
		}
		return ActivationSubjectUser, strconv.FormatInt(wxID, 10), nil
	}
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "该功能按设备开通，deviceNo 不能为空")
	}
	return ActivationSubjectDevice, deviceNo, nil
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

// HasActiveUserFeatureEntitlement 账号某功能权益是否未过期（expires_at=0 为永久）。
func HasActiveUserFeatureEntitlement(ctx context.Context, wxID int64, featureID string) (active bool, expiresAt int64, err error) {
	featureID = strings.TrimSpace(featureID)
	if wxID <= 0 || featureID == "" {
		return false, 0, nil
	}
	r, err := g.DB().Model("feature_user_entitlement").Ctx(ctx).
		Fields("expires_at").
		Where("wx_id", wxID).Where("feature_id", featureID).
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
