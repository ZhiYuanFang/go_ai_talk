package entity

// AiQuotaUserOverride per-wxId AI 月度额度 override；NULL 字段表示走全局默认。
type AiQuotaUserOverride struct {
	WxId                int64 `json:"wxId"`
	PolishMonthlyLimit  *int  `json:"polishMonthlyLimit"`
	VoiceAiMonthlyLimit *int  `json:"voiceAiMonthlyLimit"`
	UpdatedAt           int64 `json:"updatedAt"`
}
