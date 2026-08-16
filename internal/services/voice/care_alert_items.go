package voice

// CareAlertAnalyzeReason 卡片原因片段（Gateway emit_care_cards items 反序列化用）。
type CareAlertAnalyzeReason struct {
	Type            string   `json:"type"`
	Score           float64  `json:"score"`
	ExpectationUsed bool     `json:"expectationUsed"`
	AgeMonths       int      `json:"ageMonths"`
	MedianGapMs     int64    `json:"medianGapMs"`
	LastGapMs       int64    `json:"lastGapMs"`
	ExpectGapMaxMs  int64    `json:"expectGapMaxMs"`
	P75DurMs        int64    `json:"p75DurMs"`
	ElapsedMs       int64    `json:"elapsedMs"`
	ExpectDurMaxMs  int64    `json:"expectDurMaxMs"`
	DailyAvg        float64  `json:"dailyAvg"`
	Recent48hCount  int      `json:"recent48hCount"`
	StillExpected   bool     `json:"stillExpected"`
	DetailLines     []string `json:"detailLines"`
}

// CareAlertAnalyzeItem 单条留意卡片（suggestionId 可空，由 Go 补齐）。
type CareAlertAnalyzeItem struct {
	SuggestionID   string                   `json:"suggestionId"`
	EventID        string                   `json:"eventId"`
	EventName      string                   `json:"eventName"`
	SummaryLine    string                   `json:"summaryLine"`
	FollowUpPrompt string                   `json:"followUpPrompt"`
	Reasons        []CareAlertAnalyzeReason `json:"reasons"`
}
