package voice

import "strings"

// ActionTargetType 表示动作目标类型枚举（wire 为字符串）。
type ActionTargetType string

// 动作目标枚举。
const (
	ActionTargetTypeStart        ActionTargetType = "start"        // 开始记录计时动作
	ActionTargetTypeEnd          ActionTargetType = "end"          // 结束记录计时动作
	ActionTargetTypeOne          ActionTargetType = "one"          // 记录一次性动作
	ActionTargetTypeExit         ActionTargetType = "exit"         // 退出动作
	ActionTargetTypeSuggest      ActionTargetType = "suggest"      // 成长建议动作
	ActionTargetTypeSearch       ActionTargetType = "search"       // 搜索动作
	ActionTargetTypeConversation ActionTargetType = "conversation" // 对话动作
)

func (t ActionTargetType) String() string {
	return string(t)
}

func ParseActionTargetType(raw string) ActionTargetType {
	switch ActionTargetType(strings.TrimSpace(strings.ToLower(raw))) {
	case ActionTargetTypeStart, ActionTargetTypeEnd, ActionTargetTypeOne, ActionTargetTypeExit, ActionTargetTypeSuggest, ActionTargetTypeSearch, ActionTargetTypeConversation:
		return ActionTargetType(strings.TrimSpace(strings.ToLower(raw)))
	default:
		return ActionTargetTypeConversation
	}
}

// ActionTargetTypeChinese 将动作目标类型映射为简短中文说明（管理端展示用）。
func ActionTargetTypeChinese(t ActionTargetType) string {
	switch ParseActionTargetType(t.String()) {
	case ActionTargetTypeStart:
		return "开始记录计时"
	case ActionTargetTypeEnd:
		return "结束记录计时"
	case ActionTargetTypeOne:
		return "一次性记录"
	case ActionTargetTypeExit:
		return "退出"
	case ActionTargetTypeSuggest:
		return "成长建议"
	case ActionTargetTypeSearch:
		return "搜索"
	default:
		return "对话"
	}
}

// STTProfileConfig 单路 STT 配置（对话 chat / 听写 dictation 可独立选型）。
type STTProfileConfig struct {
	Provider               string  `json:"provider"`
	Endpoint               string  `json:"endpoint"`
	StreamEnabled          bool    `json:"streamEnabled"`
	StreamEndpoint         string  `json:"streamEndpoint"`
	RealtimeDebounceMs     int     `json:"realtimeDebounceMs"`
	RealtimeMinRunes       int     `json:"realtimeMinRunes"`
	SN                     string  `json:"sn"`
	APIKey                 string  `json:"apiKey"`
	APISecret              string  `json:"apiSecret"`
	TokenEndpoint          string  `json:"tokenEndpoint"`
	Model                  string  `json:"model"`
	TimeoutSeconds         int     `json:"timeoutSeconds"`
	CUID                   string  `json:"cuid"`
	DevPID                 int     `json:"devPid"`
	Format                 string  `json:"format"`
	MaxConcurrency         int     `json:"maxConcurrency"`
	WorkspaceID            string  `json:"workspaceId"`
	SpeechNoiseThreshold   float64 `json:"speechNoiseThreshold"`
	FallbackProvider       string  `json:"fallbackProvider"`
}

// VoiceChatConfig 语音对话相关配置（ASR/LLM/TTS/会话缓存）。
type VoiceChatConfig struct {
	DebugLog bool `json:"debugLog"`
	Audio    struct {
		SampleRate     int `json:"sampleRate"`
		Bits           int `json:"bits"`
		Channels       int `json:"channels"`
		MaxDurationSec int `json:"maxDurationSec"`
		MaxSizeBytes   int `json:"maxSizeBytes"`
	} `json:"audio"`
	Session struct {
		MaxRounds              int `json:"maxRounds"`
		TTLSeconds             int `json:"ttlSeconds"`
		MaxDeviceSessions      int `json:"maxDeviceSessions"`
		CleanupIntervalSeconds int `json:"cleanupIntervalSeconds"`
	} `json:"session"`
	// STT 为历史兼容块；听写 profile 未显式配置 sttDictation 时回退至此。
	STT STTProfileConfig `json:"stt"`
	// STTChat 对话 WebSocket（/voice/chat/ws）专用 STT，默认百炼远场模型。
	STTChat STTProfileConfig `json:"sttChat"`
	// STTDictation 听写 WebSocket（/voice/asr/ws）专用 STT，默认百度近场。
	STTDictation STTProfileConfig `json:"sttDictation"`
	DeepSeek struct {
		Endpoint       string `json:"endpoint"`
		APIKey         string `json:"apiKey"`
		Model          string `json:"model"`
		SystemPrompt   string `json:"systemPrompt"`
		Stream         bool   `json:"stream"`
		MinTextLength  int    `json:"minTextLength"`
		TimeoutSeconds int    `json:"timeoutSeconds"`
		MaxConcurrency int    `json:"maxConcurrency"`
	} `json:"deepseek"`
	TTS struct {
		Provider                   string `json:"provider"`
		StreamEnabled              bool   `json:"streamEnabled"`
		StreamEndpoint             string `json:"streamEndpoint"`
		APIKey                     string `json:"apiKey"`
		APISecret                  string `json:"apiSecret"`
		TokenEndpoint              string `json:"tokenEndpoint"`
		Model                      string `json:"model"`
		Voice                      string `json:"voice"`
		TimeoutSeconds             int    `json:"timeoutSeconds"`
		CUID                       string `json:"cuid"`
		Language                   string `json:"language"`
		Speed                      string `json:"speed"`
		Pitch                      string `json:"pitch"`
		Volume                     string `json:"volume"`
		AUE                        string `json:"aue"`
		StreamIdleTimeoutSeconds   int    `json:"streamIdleTimeoutSeconds"`
		StreamFinishTimeoutSeconds int    `json:"streamFinishTimeoutSeconds"`
	} `json:"tts"`
}
