package cash

// 商业功能开通域常量。

import "strings"

const (
	// FeatureIDPredictionUnlock 预测事项开通数量功能 ID（已停用；保留常量供历史/Admin 过滤）。
	FeatureIDPredictionUnlock = "prediction_unlock"
	// FeatureIDPredictCapabilityBasic 已停用基础版 ID（Ensure 置 status=0；拒绝新开通）。
	FeatureIDPredictCapabilityBasic = "predict_capability_basic"
	// FeatureIDPredictCapabilityPro 预测能力 Pro（账号维开通壳；业务能力后续提案）。
	FeatureIDPredictCapabilityPro = "predict_capability_pro"
	// FeatureIDCareAlertSmartRemindLegacy / FeatureIDGrowthTrajectoryPredictLegacy 已退役 ID（Ensure 置 status=0）。
	FeatureIDCareAlertSmartRemindLegacy     = "care_alert_smart_remind"
	FeatureIDGrowthTrajectoryPredictLegacy  = "growth_trajectory_predict"

	GrantKindEntitlement       = "entitlement"
	GrantKindAllowedCountDelta = "allowed_count_delta"

	UnlockMethodPayment    = "payment"
	UnlockMethodInviteCode = "invite_code"
	UnlockMethodAd         = "ad" // 已下线；Activate / Admin / App complete 均拒绝
	// UnlockMethodTrial 免费试用通道（成功落库后 claim，授予 TrialDurationHours）。
	UnlockMethodTrial = "trial"
	// UnlockMethodAdmin Hub 手工授（补单/赠送）；写入权益 unlock_method。
	UnlockMethodAdmin = "admin"

	// ActivationSubjectDevice 权益落在 device_no（全家共享）。
	ActivationSubjectDevice = "device"
	// ActivationSubjectUser 权益落在 wx_id（一人一份）。
	ActivationSubjectUser = "user"

	// PredictCapabilityBasicProductCode 已停用基础版 SKU 编码（Ensure 置 status=0）。
	PredictCapabilityBasicProductCode = "feat_predict_basic_30d"
	// PredictCapabilityProProductCode 预测 Pro 30 天 SKU（占位价，运维可改）。
	PredictCapabilityProProductCode = "feat_predict_pro_30d"
	// CareAlert / Growth 旧 SKU：Ensure 停用，禁止新种子。
	CareAlertSmartRemindProductCodeLegacy       = "feat_care_alert_30d"
	CareAlertSmartRemindProductCodeLegacyPerm   = "feat_care_alert_remind_perm"
	GrowthTrajectoryPredictProductCodeLegacy    = "feat_growth_traj_30d"

	// TrialStatusUnused 试用未领取。
	TrialStatusUnused = "unused"
	// TrialStatusUsed 试用已领取（首次成功落库后）。
	TrialStatusUsed = "used"
	// TrialDurationHours 试用权益时长（小时）；soft claim 遗留路径仍可用。
	TrialDurationHours = 24
	// PredictTrialDurationDays 预测能力点试用授予天数（自点击起算）。
	PredictTrialDurationDays = 7

	// AllowedCountFullAccessSentinel catalog 预测项临时/永久全开哨兵（客户端约定：-1=全部可看）。
	AllowedCountFullAccessSentinel = -1
)

// InviteOncePerUser 该功能邀请开通是否按账号仅一次（跨任意邀请码）。
// 预测 Pro：同一人只能邀请开通一次；人×码×功能去重仍保留。
func InviteOncePerUser(featureID string) bool {
	switch featureID {
	case FeatureIDPredictCapabilityPro:
		return true
	default:
		return false
	}
}

// IsPredictCapabilityProFeature 是否为预测能力 Pro（点试用仅允许此功能）。
func IsPredictCapabilityProFeature(featureID string) bool {
	return strings.TrimSpace(featureID) == FeatureIDPredictCapabilityPro
}

// NormalizeActivationSubject 规范化开通主体；空或未知回落 device。
func NormalizeActivationSubject(s string) string {
	switch strings.TrimSpace(s) {
	case ActivationSubjectUser:
		return ActivationSubjectUser
	default:
		return ActivationSubjectDevice
	}
}
