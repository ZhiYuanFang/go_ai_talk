package contracts

// AIQuotaFeature 额度维度（润笔 / 喂养 AI）。
type AIQuotaFeature string

const (
	AIQuotaPolish  AIQuotaFeature = "polish"
	AIQuotaVoiceAI AIQuotaFeature = "voice_ai"
)

// AIQuotaSnapshot 某 feature 当月 used/limit 快照。
type AIQuotaSnapshot struct {
	Used    int  `json:"used"`
	Limit   int  `json:"limit"`
	Allowed bool `json:"allowed"`
}
