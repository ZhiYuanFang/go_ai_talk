package aimodel

// zhipuAdapter 智谱 OpenAI 兼容 chat/completions。
type zhipuAdapter struct{}

func (zhipuAdapter) ApplyThinkingOptions(req ChatRequest, payload map[string]interface{}) {
	// GLM-4.7 等模型在未传 thinking 时默认开启；必须显式 disabled/enabled，不可省略字段。
	thinkingType := "disabled"
	if req.ThinkingEnabled {
		thinkingType = "enabled"
	}
	payload["thinking"] = map[string]interface{}{"type": thinkingType}
}

func (zhipuAdapter) ParseStreamDelta(data string) (string, string, string, error) {
	return parseOpenAIStreamDelta(data)
}
