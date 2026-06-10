package entity

// AiQuotaDefault AI 月度额度全局默认（singleton id=1）。
type AiQuotaDefault struct {
	Id                  int   `json:"id"`
	PolishMonthlyLimit  int   `json:"polishMonthlyLimit"`
	VoiceAiMonthlyLimit int   `json:"voiceAiMonthlyLimit"`
	UpdatedAt           int64 `json:"updatedAt"`
}
