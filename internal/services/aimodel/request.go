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
	TimeoutSec int
	// ThinkingEnabled 是否启用上游 thinking/reasoning。零值 false 表示关闭（智谱会显式发送 thinking.type=disabled）。
	// 仅 clinic 等需要 reasoning 流的场景应设为 true。
	ThinkingEnabled bool
	// ReasoningEffort 上游 reasoning 强度（如 DeepSeek high）；仅 ThinkingEnabled 为 true 时生效。
	ReasoningEffort string
	// ExtraTopLevel 合并进请求体顶层（如成长建议 child_info/history）。
	ExtraTopLevel map[string]interface{}
	// MaxTokens 上游 completion token 上限；thinking 开启时 reasoning 与 content 共用该预算，非仅正文长度。
	MaxTokens int
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
