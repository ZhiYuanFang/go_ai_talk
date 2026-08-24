package voice

import (
	"os"
	"strings"

	contracts "hello/internal/services/contracts"
)

const (
	defaultDashScopeSTTModel        = "qwen-audio-3.0-asr-flash-streaming"
	defaultDashScopeWorkspaceRegion = "cn-beijing"
	defaultDashScopeSpeechNoise     = -0.2
)

// normalizeVoiceSTTProfiles 听写未单独配置时回退至 legacy stt 块，避免升级后听写行为突变。
func normalizeVoiceSTTProfiles(cfg *VoiceChatConfig) {
	if strings.TrimSpace(cfg.STTDictation.Provider) == "" {
		cfg.STTDictation = cfg.STT
	}
}

// applySTTProfileDefaults 对单路 STT 配置填默认值；forChat 为真时按百炼远场场景兜底。
func applySTTProfileDefaults(cfg *STTProfileConfig, forChat bool) {
	if cfg.TimeoutSeconds == 0 {
		cfg.TimeoutSeconds = 20
	}
	if cfg.Provider == "" {
		if forChat {
			cfg.Provider = "dashscope"
		} else {
			cfg.Provider = "generic"
		}
	}
	if cfg.Format == "" {
		cfg.Format = "pcm"
	}
	if cfg.CUID == "" {
		cfg.CUID = "voice-client"
	}
	if cfg.MaxConcurrency <= 0 {
		cfg.MaxConcurrency = 2
	}
	if forChat {
		if cfg.RealtimeDebounceMs <= 0 {
			cfg.RealtimeDebounceMs = 1200
		}
		if cfg.RealtimeMinRunes <= 0 {
			cfg.RealtimeMinRunes = 4
		}
	}
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "baidu" {
		if cfg.TokenEndpoint == "" {
			cfg.TokenEndpoint = "https://aip.baidubce.com/oauth/2.0/token"
		}
		if cfg.StreamEndpoint == "" {
			cfg.StreamEndpoint = "wss://vop.baidu.com/realtime_asr"
		}
		if cfg.DevPID == 0 {
			cfg.DevPID = 1537
		}
	}
	if provider == "dashscope" {
		if strings.TrimSpace(cfg.Model) == "" {
			cfg.Model = defaultDashScopeSTTModel
		}
		if cfg.SpeechNoiseThreshold == 0 {
			cfg.SpeechNoiseThreshold = defaultDashScopeSpeechNoise
		}
	}
}

// applyAllSTTDefaults 对 legacy stt、对话、听写三路 STT 分别套默认值。
func applyAllSTTDefaults(cfg *VoiceChatConfig) {
	applySTTProfileDefaults(&cfg.STT, false)
	applySTTProfileDefaults(&cfg.STTChat, true)
	applySTTProfileDefaults(&cfg.STTDictation, false)
	normalizeVoiceSTTProfiles(cfg)
}

// sttMaxConcurrency 取三路 STT 并发上限最大值，供限流器容量。
func sttMaxConcurrency(cfg VoiceChatConfig) int {
	max := cfg.STT.MaxConcurrency
	if cfg.STTChat.MaxConcurrency > max {
		max = cfg.STTChat.MaxConcurrency
	}
	if cfg.STTDictation.MaxConcurrency > max {
		max = cfg.STTDictation.MaxConcurrency
	}
	return max
}

// sttConfigForProfile 按 chat/dictation 返回生效的 STT 配置副本。
func (s *VoiceService) sttConfigForProfile(profile contracts.STTProfile) STTProfileConfig {
	switch profile {
	case contracts.STTProfileChat:
		return s.cfg.STTChat
	default:
		return s.cfg.STTDictation
	}
}

// resolveDashScopeAPIKey 解析百炼 API Key：配置 > VOICE_DASHSCOPE_API_KEY > UCG_DASHSCOPE_API_KEY。
func resolveDashScopeAPIKey(cfg STTProfileConfig) string {
	if key := strings.TrimSpace(cfg.APIKey); key != "" {
		return key
	}
	if key := strings.TrimSpace(os.Getenv("VOICE_DASHSCOPE_API_KEY")); key != "" {
		return key
	}
	return strings.TrimSpace(os.Getenv("UCG_DASHSCOPE_API_KEY"))
}

// resolveDashScopeWorkspaceID 解析百炼业务空间 ID，WebSocket 端点必填。
func resolveDashScopeWorkspaceID(cfg STTProfileConfig) string {
	if id := strings.TrimSpace(cfg.WorkspaceID); id != "" {
		return id
	}
	return strings.TrimSpace(os.Getenv("DASHSCOPE_WORKSPACE_ID"))
}

// buildDashScopeStreamWSURL 构造百炼实时 ASR WebSocket 地址。
func buildDashScopeStreamWSURL(cfg STTProfileConfig) (string, error) {
	if endpoint := strings.TrimSpace(cfg.StreamEndpoint); endpoint != "" {
		return endpoint, nil
	}
	workspaceID := resolveDashScopeWorkspaceID(cfg)
	if workspaceID == "" {
		return "", StageError{Stage: "stt", Detail: "DashScope Workspace ID 未配置（sttChat.workspaceId 或 DASHSCOPE_WORKSPACE_ID）"}
	}
	return "wss://" + workspaceID + "." + defaultDashScopeWorkspaceRegion + ".maas.aliyuncs.com/api-ws/v1/inference", nil
}
