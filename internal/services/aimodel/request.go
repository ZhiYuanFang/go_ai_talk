package aimodel

// Message OpenAI 兼容 chat 消息。
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string 或多模态 []map[string]interface{}
}

// ChatRequest 统一上游请求参数。
type ChatRequest struct {
	Messages        []Message
	Stream          bool
	TimeoutSec      int
	ThinkingEnabled bool
	ReasoningEffort string
	// ExtraTopLevel 合并进请求体顶层（如成长建议 child_info/history）。
	ExtraTopLevel map[string]interface{}
	MaxTokens     int
}

// ChatResponse 非流式响应。
type ChatResponse struct {
	RawBody []byte
	Content string
}

// StreamCallbacks 流式回调；clinic thinking/answer 与 casual chat 共用 content 通道。
type StreamCallbacks struct {
	OnThinkingDelta func(delta string) error
	OnAnswerDelta   func(delta string) error
	OnContentDelta  func(delta string) error
}

// StreamResult 流式聚合结果。
type StreamResult struct {
	Thinking string
	Answer   string
	Content  string
}
