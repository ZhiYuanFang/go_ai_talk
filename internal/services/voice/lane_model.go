package voice

// LaneModelCfg Go 选模结果（注模进 Gateway 的 x-openclaw-model 用 provider/name）。
// 历史名 PythonModelCfg 保留别名，避免大范围改签名。
type LaneModelCfg struct {
	Provider    string `json:"provider"`
	Name        string `json:"name"`
	MaxInFlight int    `json:"max_in_flight"`
}

// PythonModelCfg 兼容旧名（选模结构，非「再调 Python 编排」）。
type PythonModelCfg = LaneModelCfg

// IntentEvent 遗留结构（流式合成 Result 时可选）；产品权威已不在此信封。
type IntentEvent struct {
	Op            string `json:"op,omitempty"`
	Action        string `json:"action,omitempty"`
	EventName     string `json:"event_name"`
	EventId       string `json:"event_id"`
	Quantity      *int   `json:"quantity,omitempty"`
	RemarkKeyword string `json:"remark_keyword,omitempty"`
	StartTime     int64  `json:"start_time,omitempty"`
	EndTime       int64  `json:"end_time,omitempty"`
}

// AnalyzeIntentResponse 流式路径合成用的轻量结构（仅 Content 等播报字段）。
type AnalyzeIntentResponse struct {
	TargetType     string        `json:"target_type"`
	Action         string        `json:"action,omitempty"`
	EventName      string        `json:"event_name"`
	EventId        string        `json:"event_id"`
	Quantity       *int          `json:"quantity,omitempty"`
	EventType      string        `json:"event_type,omitempty"`
	EventUnit      string        `json:"event_unit,omitempty"`
	IsNewEvent     bool          `json:"is_new_event,omitempty"`
	Keywords       []string      `json:"keywords"`
	Content        string        `json:"content"`
	NeedConfirm    bool          `json:"need_confirm"`
	ConfirmMessage string        `json:"confirm_message"`
	ConversationID string        `json:"conversation_id,omitempty"`
	Events         []IntentEvent `json:"events"`
}

// AnalyzeIntentStreamResponse 流式意图累计结果（Gateway 路径只填 Answer/Result.Content）。
type AnalyzeIntentStreamResponse struct {
	Thinking string
	Answer   string
	Result   *AnalyzeIntentResponse
}
