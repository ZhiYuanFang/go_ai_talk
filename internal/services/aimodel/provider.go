package aimodel

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

// ProviderAdapter 各上游 provider 的请求/流式差异适配。
type ProviderAdapter interface {
	ApplyThinkingOptions(req ChatRequest, payload map[string]interface{})
	ParseStreamDelta(data string) (thinking, answer, content string, err error)
}

func getProviderAdapter(p Provider) ProviderAdapter {
	switch p {
	case ProviderZhipu:
		return zhipuAdapter{}
	case ProviderDeepSeek:
		return deepseekAdapter{}
	case ProviderDashScope:
		return dashscopeAdapter{}
	default:
		return genericAdapter{}
	}
}

type genericAdapter struct{}

func (genericAdapter) ApplyThinkingOptions(req ChatRequest, _ map[string]interface{}) {}
func (genericAdapter) ParseStreamDelta(data string) (string, string, string, error) {
	return parseOpenAIStreamDelta(data)
}

// resolveAPIKey 从环境变量或 GF 配置读取 provider 密钥。
func resolveAPIKey(ctx context.Context, provider Provider) (string, error) {
	if ctx == nil {
		ctx = gctx.New()
	}
	switch provider {
	case ProviderZhipu:
		if k := strings.TrimSpace(os.Getenv("GLM_API_KEY")); k != "" {
			return k, nil
		}
		return "", fmt.Errorf("%w: %s", ErrProviderKeyMissing, ProviderKeyEnv(provider))
	case ProviderDashScope:
		if k := strings.TrimSpace(os.Getenv("UCG_DASHSCOPE_API_KEY")); k != "" {
			return k, nil
		}
		if k := strings.TrimSpace(os.Getenv("UCG_DEEPSEEK_API_KEY")); k != "" {
			return k, nil
		}
		return "", fmt.Errorf("%w: %s", ErrProviderKeyMissing, ProviderKeyEnv(provider))
	case ProviderDeepSeek:
		if k := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")); k != "" {
			return k, nil
		}
		if k := strings.TrimSpace(g.Cfg().MustGet(ctx, "voiceChat.deepseek.apiKey").String()); k != "" {
			return k, nil
		}
		return "", fmt.Errorf("%w: %s", ErrProviderKeyMissing, ProviderKeyEnv(provider))
	default:
		return "", fmt.Errorf("未知 provider: %s", provider)
	}
}

// parseOpenAIStreamDelta 解析 OpenAI 兼容 SSE 单分片 delta。
// 注意：reasoning 与 content 常在不同分片到达；thinking 开启场景下 invokeStreamHTTP 在见过 reasoning 后将 content 路由到 answer。
func parseOpenAIStreamDelta(data string) (thinking, answer, content string, err error) {
	var obj map[string]interface{}
	if err = json.Unmarshal([]byte(data), &obj); err != nil {
		return "", "", "", err
	}
	choices, ok := obj["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", "", "", nil
	}
	first, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", "", "", nil
	}
	delta, ok := first["delta"].(map[string]interface{})
	if !ok {
		return "", "", "", nil
	}
	if v, ok := delta["reasoning_content"].(string); ok && v != "" {
		thinking = v
	}
	if thinking == "" {
		if v, ok := delta["thinking"].(string); ok && v != "" {
			thinking = v
		}
	}
	if v, ok := delta["content"].(string); ok && v != "" {
		if thinking != "" {
			answer = v
		} else {
			content = v
		}
	}
	return thinking, answer, content, nil
}
