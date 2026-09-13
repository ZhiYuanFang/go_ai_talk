package cash

// 商业功能开通域常量（与 VIP 隔离）。

import "strings"

const (
	// FeatureIDPredictionUnlock 预测事项开通数量功能 ID。
	FeatureIDPredictionUnlock = "prediction_unlock"
	// FeatureIDCareAlertSmartRemind 值得留意智能提醒功能 ID（设备维权益；邀请/广告限时）。
	FeatureIDCareAlertSmartRemind = "care_alert_smart_remind"
	// FeatureIDGrowthTrajectoryPredict 成长轨迹预测功能 ID（账号维权益；邀请/广告限时；无喂养门闸）。
	FeatureIDGrowthTrajectoryPredict = "growth_trajectory_predict"

	GrantKindEntitlement       = "entitlement"
	GrantKindAllowedCountDelta = "allowed_count_delta"

	UnlockMethodPayment    = "payment"
	UnlockMethodInviteCode = "invite_code"
	UnlockMethodAd         = "ad"

	// ActivationSubjectDevice 权益落在 device_no（全家共享）。
	ActivationSubjectDevice = "device"
	// ActivationSubjectUser 权益落在 wx_id（一人一份）。
	ActivationSubjectUser = "user"

	// CareAlertSmartRemindProductCode 值得留意付费永久 SKU 种子编码。
	CareAlertSmartRemindProductCode = "feat_care_alert_remind_perm"
	// GrowthTrajectoryPredictProductCode 成长轨迹预测付费 30 天 SKU 种子编码（无永久）。
	GrowthTrajectoryPredictProductCode = "feat_growth_traj_30d"

	// AllowedCountFullAccessSentinel catalog 预测项临时/永久全开哨兵（客户端约定：-1=全部可看）。
	AllowedCountFullAccessSentinel = -1
)

// InviteOncePerDevice 该功能邀请开通是否按设备仅一次（值得留意防刷；与权益主体无关）。
// 成长轨迹为账号维开通，同宝宝多家长须各自可兑不同码，故不得再适用本闸。
func InviteOncePerDevice(featureID string) bool {
	return featureID == FeatureIDCareAlertSmartRemind
}

// InviteOncePerUser 该功能邀请开通是否按账号仅一次（跨任意邀请码）。
// 成长轨迹：权益跟人，同一人只能邀请开通一次；同机可兑多个不同好友码。
func InviteOncePerUser(featureID string) bool {
	return featureID == FeatureIDGrowthTrajectoryPredict
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
