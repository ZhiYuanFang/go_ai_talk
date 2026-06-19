package aimodel

// dashscopeAdapter 阿里云 DashScope OpenAI 兼容模式（润笔多模态消息由业务层构造）。
type dashscopeAdapter struct{}

func (dashscopeAdapter) ApplyThinkingOptions(_ ChatRequest, _ map[string]interface{}) {}

func (dashscopeAdapter) ParseStreamDelta(data string) (string, string, string, error) {
	return parseOpenAIStreamDelta(data)
}
