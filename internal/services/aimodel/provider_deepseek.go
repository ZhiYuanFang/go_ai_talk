package aimodel

import "strings"

// deepseekAdapter DeepSeek chat/completions（含 reasoning_effort + extra_body thinking）。
type deepseekAdapter struct{}

func (deepseekAdapter) ApplyThinkingOptions(req ChatRequest, payload map[string]interface{}) {
	// DeepSeek 非默认强制 thinking；false 时不写入 thinking 相关字段即可。
	if !req.ThinkingEnabled {
		return
	}
	if re := strings.TrimSpace(req.ReasoningEffort); re != "" {
		payload["reasoning_effort"] = re
	}
	payload["extra_body"] = map[string]interface{}{
		"thinking": map[string]interface{}{"type": "enabled"},
	}
}

func (deepseekAdapter) ParseStreamDelta(data string) (string, string, string, error) {
	return parseOpenAIStreamDelta(data)
}
