package cash

// 商业功能开通域常量（与 VIP 隔离）。

const (
	// FeatureIDPredictionUnlock 预测事项开通数量功能 ID。
	FeatureIDPredictionUnlock = "prediction_unlock"

	GrantKindEntitlement       = "entitlement"
	GrantKindAllowedCountDelta = "allowed_count_delta"

	UnlockMethodPayment    = "payment"
	UnlockMethodInviteCode = "invite_code"
	UnlockMethodAd         = "ad"
)
