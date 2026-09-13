package cash

// 功能管理只读开通规则文案（按 featureId 派生，不可经 Admin 编辑改写闸门）。

import "strings"

// 共用邀请前置说明（兑码侧已实现）。
const inviteCommonRulePrefix = "共用：不可使用自己的邀请码；不可使用同一宝宝（同 deviceNo）下其他账号的邀请码。"

// FeatureRuleSummary 按功能编号返回运维只读规则说明；未知功能返回通用提示。
//
// Args: featureID 功能编号。
// Returns: 中文多句说明（含主体、邀请、付费口径）。
func FeatureRuleSummary(featureID string) string {
	switch strings.TrimSpace(featureID) {
	case FeatureIDPredictionUnlock:
		return inviteCommonRulePrefix +
			" 本功能：开放主体=对机（device）。邀请：同宝宝可兑多个不同好友码，人×码×功能仅一次；每次永久预测可看条数 +1（非限时开关）。" +
			"付费/广告：同样对机累加条数，不跟人走。"
	case FeatureIDCareAlertSmartRemind:
		return inviteCommonRulePrefix +
			" 本功能：开放主体=对机（device），全家共享权益。邀请：同一宝宝对本功能仅能成功邀请开通一次；授予天数看功能定义「邀请授予天数」（0=永久）。" +
			"付费：对机写入权益（种子 SKU 多为永久）。VIP 可覆盖使用权，但开通快照页不含 VIP 旁路。"
	case FeatureIDGrowthTrajectoryPredict:
		return inviteCommonRulePrefix +
			" 本功能：开放主体=对人（user），一人一份。邀请：同宝宝可兑多个不同好友码；同一人对本功能任意邀请码仅能成功一次；授予天数看「邀请授予天数」。" +
			"付费：对人开通（种子多为 30 天限时）。VIP 可覆盖 turn 使用权，快照页不含 VIP。"
	default:
		return inviteCommonRulePrefix + " 请以功能「开放主体」与「允许的开通方式」为准；具体邀请防刷规则以服务端常量为准。"
	}
}
