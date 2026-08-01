package voice

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// AIClinicConfig 胖宝诊疗配置（manifest/config/config.voice-service.yaml aiClinic 块）。
// DeepSeek endpoint/apiKey 从 voice-chat.shared.yaml 的 deepseek 段加载，不在此重复。
// 业务说明：已移除 sessionTtl/summaryTtl——Go 不再缓存对话 turns 与喂养摘要。
type AIClinicConfig struct {
	Model                  string
	Endpoint               string
	APIKey                 string
	LLMTimeoutSeconds      int
	ThinkingEnabled        bool
	ReasoningEffort        string
	RateLimitWindowSeconds int
	RateLimitMaxRequests   int
}

// CodeClinicRateLimited WS 胖宝 per-wxId 限流超限。
const CodeClinicRateLimited = 42901

func loadAIClinicConfig(ctx context.Context, voiceCfg VoiceChatConfig) AIClinicConfig {
	cfg := AIClinicConfig{
		Model:                  "deepseek-v4-pro",
		Endpoint:               strings.TrimSpace(voiceCfg.DeepSeek.Endpoint),
		APIKey:                 strings.TrimSpace(voiceCfg.DeepSeek.APIKey),
		LLMTimeoutSeconds:      120,
		ThinkingEnabled:        true,
		ReasoningEffort:        "high",
		RateLimitWindowSeconds: 60,
		RateLimitMaxRequests:   10,
	}
	value, err := g.Cfg().Get(ctx, "aiClinic")
	if err == nil && value != nil && !value.IsNil() {
		_ = value.Scan(&cfg)
	}
	if strings.TrimSpace(cfg.Endpoint) == "" {
		cfg.Endpoint = strings.TrimSpace(voiceCfg.DeepSeek.Endpoint)
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		cfg.APIKey = strings.TrimSpace(voiceCfg.DeepSeek.APIKey)
	}
	if cfg.LLMTimeoutSeconds <= 0 {
		cfg.LLMTimeoutSeconds = 120
	}
	if cfg.RateLimitWindowSeconds <= 0 {
		cfg.RateLimitWindowSeconds = 60
	}
	if cfg.RateLimitMaxRequests <= 0 {
		cfg.RateLimitMaxRequests = 10
	}
	if strings.TrimSpace(cfg.Model) == "" {
		cfg.Model = "deepseek-v4-pro"
	}
	if strings.TrimSpace(cfg.ReasoningEffort) == "" {
		cfg.ReasoningEffort = "high"
	}
	return cfg
}
