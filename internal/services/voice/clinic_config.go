package voice

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// AIClinicConfig 胖宝诊疗配置（manifest/config/config.voice-service.yaml aiClinic 块）。
// DeepSeek endpoint/apiKey 从 voice-chat.shared.yaml 的 deepseek 段加载，不在此重复。
type AIClinicConfig struct {
	Model                  string
	Endpoint               string
	APIKey                 string
	LLMTimeoutSeconds      int
	ThinkingEnabled        bool
	ReasoningEffort        string
	RateLimitWindowSeconds int
	RateLimitMaxRequests   int
	SessionTTLSeconds      int
	SummaryTTLSeconds      int
	SystemPrompt           string
}

const (
	// Redis 键前缀（负责人已确认引入 voice:clinic:* 读缓存/会话/限流）：
	// - voice:clinic:session:{wxId}  12h 固定 TTL，自首问起算，非 sliding
	// - voice:clinic:rate:{wxId}     per-wxId 限流计数
	// - voice:clinic:summary:{wxId}:{deviceNo}  7 天摘要懒刷新缓存
	clinicSessionKeyPrefix = "voice:clinic:session:"
	clinicRateKeyPrefix    = "voice:clinic:rate:"
	clinicSummaryKeyPrefix = "voice:clinic:summary:"
)

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
		SessionTTLSeconds:      12 * 3600,
		SummaryTTLSeconds:      24 * 3600,
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
	if cfg.SessionTTLSeconds <= 0 {
		cfg.SessionTTLSeconds = 12 * 3600
	}
	if cfg.SummaryTTLSeconds <= 0 {
		cfg.SummaryTTLSeconds = 24 * 3600
	}
	if strings.TrimSpace(cfg.Model) == "" {
		cfg.Model = "deepseek-v4-pro"
	}
	if strings.TrimSpace(cfg.ReasoningEffort) == "" {
		cfg.ReasoningEffort = "high"
	}
	if strings.TrimSpace(cfg.SystemPrompt) == "" {
		cfg.SystemPrompt = "你是「胖宝」AI 育儿助手，结合用户近 7 天喂养事件聚合摘要回答问题。"
	}
	return cfg
}
