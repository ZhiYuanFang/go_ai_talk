package cash

// 功能管理只读开通规则文案（按 featureId 派生，不可经 Admin 编辑改写闸门）。

import "strings"

// 共用邀请前置说明（兑码侧已实现）。
const inviteCommonRulePrefix = "共用：不可使用同一宝宝（同 deviceNo）下其他账号的邀请码。"

// FeatureRuleSummary 按功能编号返回运维只读规则说明；未知功能返回通用提示。
//
// Args: featureID 功能编号。
// Returns: 中文多句说明（含主体、邀请、付费口径）。
func FeatureRuleSummary(featureID string) string {
	switch strings.TrimSpace(featureID) {
	case FeatureIDPredictCapabilityPro:
		return inviteCommonRulePrefix +
			" 本功能：开放主体=对人（user）。开通（试用/付费/邀请）前置：当前绑机喂养资格场景 predict_pro_entry（默认连续 7 有效日）。" +
			"试用点开起算 7 日。本轮仅为开通壳，具体 Pro 能力后续开放。"
	default:
		return inviteCommonRulePrefix + " 请以功能「开放主体」与「允许的开通方式」为准；具体邀请防刷规则以服务端常量为准。"
	}
}
