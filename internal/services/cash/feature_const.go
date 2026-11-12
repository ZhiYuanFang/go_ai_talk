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
	// FeatureIDPredictCapabilityBasic 预测能力基础版（账号维；门禁 imminent 写入）。
	FeatureIDPredictCapabilityBasic = "predict_capability_basic"
	// FeatureIDPredictCapabilityPro 预测能力 Pro（账号维；开通需喂养资格；含基础版推送能力）。
	FeatureIDPredictCapabilityPro = "predict_capability_pro"

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

	// CareAlertSmartRemindProductCode 值得留意付费 30 天 SKU 种子编码（无永久）。
	CareAlertSmartRemindProductCode = "feat_care_alert_30d"
	// CareAlertSmartRemindProductCodeLegacyPerm 旧永久 SKU；EnsureSchema 停用。
	CareAlertSmartRemindProductCodeLegacyPerm = "feat_care_alert_remind_perm"
	// GrowthTrajectoryPredictProductCode 成长轨迹预测付费 30 天 SKU 种子编码（无永久）。
	GrowthTrajectoryPredictProductCode = "feat_growth_traj_30d"
	// PredictCapabilityBasicProductCode 预测基础版 30 天 SKU（占位价，运维可改）。
	PredictCapabilityBasicProductCode = "feat_predict_basic_30d"
	// PredictCapabilityProProductCode 预测 Pro 30 天 SKU（占位价，运维可改）。
	PredictCapabilityProProductCode = "feat_predict_pro_30d"

	// TrialStatusUnused 试用未领取。
	TrialStatusUnused = "unused"
	// TrialStatusUsed 试用已领取（首次成功落库后）。
	TrialStatusUsed = "used"
	// TrialDurationHours 试用权益时长（小时）；成长轨迹/值得留意 soft claim。
	TrialDurationHours = 24
	// PredictTrialDurationDays 预测能力点试用授予天数（自点击起算）。
	PredictTrialDurationDays = 7

	// AllowedCountFullAccessSentinel catalog 预测项临时/永久全开哨兵（客户端约定：-1=全部可看）。
	AllowedCountFullAccessSentinel = -1
)

// InviteOncePerUser 该功能邀请开通是否按账号仅一次（跨任意邀请码）。
// care / growth / 预测两档：同一人只能邀请开通一次；人×码×功能去重仍保留。
func InviteOncePerUser(featureID string) bool {
	switch featureID {
	case FeatureIDCareAlertSmartRemind, FeatureIDGrowthTrajectoryPredict,
		FeatureIDPredictCapabilityBasic, FeatureIDPredictCapabilityPro:
		return true
	default:
		return false
	}
}

// IsPredictCapabilityFeature 是否为预测能力基础版或 Pro。
func IsPredictCapabilityFeature(featureID string) bool {
	switch strings.TrimSpace(featureID) {
	case FeatureIDPredictCapabilityBasic, FeatureIDPredictCapabilityPro:
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
