package aimodel

// zhipuAdapter 智谱 OpenAI 兼容 chat/completions。
type zhipuAdapter struct{}

func (zhipuAdapter) ApplyThinkingOptions(req ChatRequest, payload map[string]interface{}) {
	if !req.ThinkingEnabled {
		return
	}
	payload["thinking"] = map[string]interface{}{"type": "enabled"}
}

func (zhipuAdapter) ParseStreamDelta(data string) (string, string, string, error) {
	return parseOpenAIStreamDelta(data)
}
