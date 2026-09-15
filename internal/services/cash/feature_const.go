package cash

// 商业功能开通域常量（与 VIP 隔离）。

import "strings"

const (
	// FeatureIDPredictionUnlock 预测事项开通数量功能 ID（已停用；保留常量供历史/Admin 过滤）。
	FeatureIDPredictionUnlock = "prediction_unlock"
	// FeatureIDCareAlertSmartRemind 值得留意智能提醒功能 ID（账号维权益）。
	FeatureIDCareAlertSmartRemind = "care_alert_smart_remind"
	// FeatureIDGrowthTrajectoryPredict 成长轨迹预测功能 ID（账号维权益）。
	FeatureIDGrowthTrajectoryPredict = "growth_trajectory_predict"

	GrantKindEntitlement       = "entitlement"
	GrantKindAllowedCountDelta = "allowed_count_delta"

	UnlockMethodPayment    = "payment"
	UnlockMethodInviteCode = "invite_code"
	UnlockMethodAd         = "ad" // 已从 unlock_methods 种子移除；履约路径保留拒绝对旧客户端
	// UnlockMethodTrial 免费试用通道（成功落库后 claim，授予 TrialDurationHours）。
	UnlockMethodTrial = "trial"

	// ActivationSubjectDevice 权益落在 device_no（全家共享）。
	ActivationSubjectDevice = "device"
	// ActivationSubjectUser 权益落在 wx_id（一人一份）。
	ActivationSubjectUser = "user"

	// CareAlertSmartRemindProductCode 值得留意付费 30 天 SKU 种子编码（无永久）。
	CareAlertSmartRemindProductCode = "feat_care_alert_30d"
	// CareAlertSmartRemindProductCodeLegacyPerm 旧永久 SKU；EnsureSchema 停用。
	CareAlertSmartRemindProductCodeLegacyPerm = "feat_care_alert_remind_perm"
	// GrowthTrajectoryPredictProductCode 成长轨迹预测付费 30 天 SKU 种子编码（无永久）。
	GrowthTrajectoryPredictProductCode = "feat_growth_traj_30d"

	// TrialStatusUnused 试用未领取。
	TrialStatusUnused = "unused"
	// TrialStatusUsed 试用已领取（首次成功落库后）。
	TrialStatusUsed = "used"
	// TrialDurationHours 试用权益时长（小时）。
	TrialDurationHours = 24

	// AllowedCountFullAccessSentinel catalog 预测项临时/永久全开哨兵（客户端约定：-1=全部可看）。
	AllowedCountFullAccessSentinel = -1
)

// InviteOncePerUser 该功能邀请开通是否按账号仅一次（跨任意邀请码）。
// care + growth：同一人只能邀请开通一次；人×码×功能去重仍保留。
func InviteOncePerUser(featureID string) bool {
	switch featureID {
	case FeatureIDCareAlertSmartRemind, FeatureIDGrowthTrajectoryPredict:
		return true
	default:
		return false
	}
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
