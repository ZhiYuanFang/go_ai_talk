package cash

// 商业功能开通域常量（与 VIP 隔离）。

const (
	// FeatureIDPredictionUnlock 预测事项开通数量功能 ID。
	FeatureIDPredictionUnlock = "prediction_unlock"
	// FeatureIDCareAlertSmartRemind 值得留意智能提醒功能 ID（设备维权益；邀请/广告限时）。
	FeatureIDCareAlertSmartRemind = "care_alert_smart_remind"

	GrantKindEntitlement       = "entitlement"
	GrantKindAllowedCountDelta = "allowed_count_delta"

	UnlockMethodPayment    = "payment"
	UnlockMethodInviteCode = "invite_code"
	UnlockMethodAd         = "ad"

	// ActivationSubjectDevice 权益落在 device_no（全家共享）。
	ActivationSubjectDevice = "device"
	// ActivationSubjectUser 账号维权益（一期未落表，ActivateFeature 拒绝）。
	ActivationSubjectUser = "user"

	// CareAlertSmartRemindProductCode 值得留意付费永久 SKU 种子编码。
	CareAlertSmartRemindProductCode = "feat_care_alert_remind_perm"

	// AllowedCountFullAccessSentinel catalog 预测项临时/永久全开哨兵（客户端约定：-1=全部可看）。
	AllowedCountFullAccessSentinel = -1
)

// InviteOncePerDevice 该功能邀请开通是否按设备仅一次（值得留意全家共享防刷）。
func InviteOncePerDevice(featureID string) bool {
	return featureID == FeatureIDCareAlertSmartRemind
}
