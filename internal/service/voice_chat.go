package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
)

// 动作目标枚举
const (
	ActionTargetTypeStart        = "start"        //开始记录计时动作
	ActionTargetTypeEnd          = "end"          //结束记录计时动作
	ActionTargetTypeOne          = "one"          //记录一次性动作
	ActionTargetTypeExit         = "exit"         //退出动作
	ActionTargetTypeSuggest      = "suggest"      //成长建议动作
	ActionTargetTypeSearch       = "search"       //搜索动作
	ActionTargetTypeConversation = "conversation" //对话动作

)

// ActionTargetTypeChinese 将动作目标类型映射为简短中文说明（管理端展示用）。
func ActionTargetTypeChinese(t string) string {
	switch strings.TrimSpace(t) {
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
	STT struct {
		Provider           string `json:"provider"`
		Endpoint           string `json:"endpoint"`
		StreamEnabled      bool   `json:"streamEnabled"`
		StreamEndpoint     string `json:"streamEndpoint"`
		RealtimeDebounceMs int    `json:"realtimeDebounceMs"`
		RealtimeMinRunes   int    `json:"realtimeMinRunes"`
		SN                 string `json:"sn"`
		APIKey             string `json:"apiKey"`
		APISecret          string `json:"apiSecret"`
		TokenEndpoint      string `json:"tokenEndpoint"`
		Model              string `json:"model"`
		TimeoutSeconds     int    `json:"timeoutSeconds"`
		CUID               string `json:"cuid"`
		DevPID             int    `json:"devPid"`
		Format             string `json:"format"`
		MaxConcurrency     int    `json:"maxConcurrency"`
	} `json:"stt"`
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
		Provider       string `json:"provider"`
		Endpoint       string `json:"endpoint"`
		APIKey         string `json:"apiKey"`
		APISecret      string `json:"apiSecret"`
		TokenEndpoint  string `json:"tokenEndpoint"`
		Model          string `json:"model"`
		Voice          string `json:"voice"`
		TimeoutSeconds int    `json:"timeoutSeconds"`
		CUID           string `json:"cuid"`
		Language       string `json:"language"`
		Speed          string `json:"speed"`
		Pitch          string `json:"pitch"`
		Volume         string `json:"volume"`
		AUE            string `json:"aue"`
	} `json:"tts"`
}

const baiduTTSMaxTextBytes = 1024
const (
	voiceStageSlowThreshold = 6 * time.Second
	voiceTotalSlowThreshold = 10 * time.Second
	quantityKeyword         = "多少?"
)

type baiduTokenCache struct {
	// 令牌内容。
	token string
	// 令牌过期时间（内部会预留安全窗口）。
	expiresAt time.Time
	// 保护 token/expiresAt 的并发锁。
	mu sync.Mutex
}

// chatHistoryMessage 单条对话历史消息。
type chatHistoryMessage struct {
	Role    string
	Content string
}

// deviceChatSession 设备级会话缓存。
type deviceChatSession struct {
	Messages   []chatHistoryMessage
	LastActive time.Time
}

// pendingQuantityState 记录“待补量词”的上下文。
type pendingQuantityState struct {
	HistoryID int64
	EventName string
}

// eventIntentResult 为大模型返回的结构化事件意图。
type eventIntentResult struct {
	EventName     string `json:"event_name"`
	WantDo        string `json:"want_do"`
	ActionKeyWord string `json:"action_key_word"` //动作关键词
	Action        string `json:"action"`
	Quantity      int    `json:"quantity"`
	Remark        string `json:"remark"`
	Reply         string `json:"reply"`
	Reason        string `json:"reason"`
	bool          `json:"need_time"`
	NeedQuantity  bool `json:"need_quantity"`
	HasExtraNotes bool `json:"has_extra_notes"`
	Exit          bool `json:"exit"`
}

type generalChatResult struct {
	Reply         string `json:"reply"`
	NeedUserReply bool   `json:"need_user_reply"`
}

type eventInfo struct {
	Id           int64  `json:"id"`
	Name         string `json:"name"`
	NeedQuantity int    `json:"needQuantity"`
}

// VoiceService 语音服务核心实现：
// 1) 语音转写（STT）
// 2) 对话推理（DeepSeek）
// 3) 语音合成（TTS）
// 4) 设备级会话与事件记录
type VoiceService struct {
	cfg                    VoiceChatConfig
	httpClient             *http.Client
	sttToken               baiduTokenCache
	ttsToken               baiduTokenCache
	sttLimiter             chan struct{}
	chatLimiter            chan struct{}
	sessionMu              sync.Mutex
	sessions               map[string]*deviceChatSession
	pendingQuantityMu      sync.Mutex
	pendingQuantity        map[string]pendingQuantityState
	deviceLocks            sync.Map
	janitorOnce            sync.Once
	janitorStop            chan struct{}
	ensureDeviceRegistered func(ctx context.Context, deviceNo string) error
	persistTalkRecord      func(ctx context.Context, deviceNo, ask, answer string) error
}

var (
	voiceSvc  *VoiceService
	voiceOnce sync.Once
)

func newLimiter(limit int) chan struct{} {
	if limit <= 0 {
		return nil
	}
	return make(chan struct{}, limit)
}

// Voice 返回全局单例服务，并注入设备注册校验与对话记录落库能力。
func Voice() VoiceContract {
	voiceOnce.Do(func() {
		cfg := loadVoiceConfig(gctx.New())
		voiceSvc = NewVoiceService(cfg)
		voiceSvc.ensureDeviceRegistered = DeviceAdmin().EnsureRegistered
		voiceSvc.persistTalkRecord = DeviceAdmin().UpdateLastTalk
		voiceSvc.startSessionJanitor()
	})
	return voiceSvc
}

var _ VoiceContract = (*VoiceService)(nil)

// startSessionJanitor 启动后台清理协程：定期清理过期会话并执行会话淘汰。
func (s *VoiceService) startSessionJanitor() {
	intervalSeconds := s.cfg.Session.CleanupIntervalSeconds
	if intervalSeconds <= 0 {
		return
	}
	if s.cfg.Session.TTLSeconds <= 0 && s.cfg.Session.MaxDeviceSessions <= 0 {
		return
	}

	s.janitorOnce.Do(func() {
		s.janitorStop = make(chan struct{})
		interval := time.Duration(intervalSeconds) * time.Second
		ticker := time.NewTicker(interval)
		go func() {
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					now := time.Now()
					s.sessionMu.Lock()
					s.pruneSessionsLocked(now)
					s.evictExcessSessionsLocked()
					s.sessionMu.Unlock()
				case <-s.janitorStop:
					return
				}
			}
		}()
	})
}

// stopSessionJanitor 停止后台清理协程。
func (s *VoiceService) stopSessionJanitor() {
	if s.janitorStop == nil {
		return
	}
	close(s.janitorStop)
	s.janitorStop = nil
}

// NewVoiceService 创建服务实例并应用默认配置。
func NewVoiceService(cfg VoiceChatConfig) *VoiceService {
	applyConfigDefaults(&cfg)
	return &VoiceService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		sttLimiter:             newLimiter(cfg.STT.MaxConcurrency),
		chatLimiter:            newLimiter(cfg.DeepSeek.MaxConcurrency),
		sessions:               make(map[string]*deviceChatSession),
		pendingQuantity:        make(map[string]pendingQuantityState),
		ensureDeviceRegistered: func(ctx context.Context, deviceNo string) error { return nil },
		persistTalkRecord:      func(ctx context.Context, deviceNo, ask, answer string) error { return nil },
	}
}

// applyConfigDefaults 对配置进行兜底，避免缺省导致运行错误。
func applyConfigDefaults(cfg *VoiceChatConfig) {
	if cfg.Audio.SampleRate == 0 {
		cfg.Audio.SampleRate = 16000
	}
	if cfg.Audio.Bits == 0 {
		cfg.Audio.Bits = 16
	}
	if cfg.Audio.Channels == 0 {
		cfg.Audio.Channels = 1
	}
	if cfg.Audio.MaxDurationSec == 0 {
		cfg.Audio.MaxDurationSec = 30
	}
	if cfg.Audio.MaxSizeBytes == 0 {
		cfg.Audio.MaxSizeBytes = cfg.Audio.SampleRate * cfg.Audio.Channels * (cfg.Audio.Bits / 8) * cfg.Audio.MaxDurationSec
	}
	if cfg.Session.MaxRounds <= 0 {
		cfg.Session.MaxRounds = 10
	}
	if cfg.Session.TTLSeconds <= 0 {
		cfg.Session.TTLSeconds = 30 * 60
	}
	if cfg.Session.MaxDeviceSessions <= 0 {
		cfg.Session.MaxDeviceSessions = 1000
	}
	if cfg.Session.CleanupIntervalSeconds <= 0 {
		cfg.Session.CleanupIntervalSeconds = 60
	}
	if cfg.STT.TimeoutSeconds == 0 {
		cfg.STT.TimeoutSeconds = 20
	}
	if cfg.STT.Provider == "" {
		cfg.STT.Provider = "generic"
	}
	if cfg.STT.Format == "" {
		cfg.STT.Format = "pcm"
	}
	if cfg.STT.CUID == "" {
		cfg.STT.CUID = "voice-client"
	}
	if cfg.STT.MaxConcurrency <= 0 {
		cfg.STT.MaxConcurrency = 2
	}
	if cfg.STT.RealtimeDebounceMs <= 0 {
		cfg.STT.RealtimeDebounceMs = 1200
	}
	if cfg.STT.RealtimeMinRunes <= 0 {
		cfg.STT.RealtimeMinRunes = 4
	}
	if cfg.STT.Provider == "baidu" {
		if cfg.STT.TokenEndpoint == "" {
			cfg.STT.TokenEndpoint = "https://aip.baidubce.com/oauth/2.0/token"
		}
		if cfg.STT.StreamEndpoint == "" {
			cfg.STT.StreamEndpoint = "wss://vop.baidu.com/realtime_asr"
		}
		if cfg.STT.DevPID == 0 {
			cfg.STT.DevPID = 1537
		}
	}
	if cfg.DeepSeek.TimeoutSeconds == 0 {
		cfg.DeepSeek.TimeoutSeconds = 20
	}
	if cfg.DeepSeek.MinTextLength <= 0 {
		cfg.DeepSeek.MinTextLength = 2
	}
	if cfg.DeepSeek.MaxConcurrency <= 0 {
		cfg.DeepSeek.MaxConcurrency = 2
	}
	if cfg.TTS.TimeoutSeconds == 0 {
		cfg.TTS.TimeoutSeconds = 20
	}
	if cfg.TTS.Provider == "" {
		cfg.TTS.Provider = "generic"
	}
	if cfg.TTS.CUID == "" {
		cfg.TTS.CUID = cfg.STT.CUID
	}
	if cfg.TTS.Language == "" {
		cfg.TTS.Language = "zh"
	}
	if cfg.TTS.Speed == "" {
		cfg.TTS.Speed = "5"
	}
	if cfg.TTS.Pitch == "" {
		cfg.TTS.Pitch = "5"
	}
	if cfg.TTS.Volume == "" {
		cfg.TTS.Volume = "5"
	}
	if cfg.TTS.AUE == "" {
		cfg.TTS.AUE = "6"
	}
	if cfg.TTS.Provider == "baidu" && cfg.TTS.TokenEndpoint == "" {
		cfg.TTS.TokenEndpoint = "https://aip.baidubce.com/oauth/2.0/token"
	}
}

// loadVoiceConfig 从配置中心加载 voiceChat 配置并应用默认值。
func loadVoiceConfig(ctx context.Context) VoiceChatConfig {
	var cfg VoiceChatConfig
	value := g.Cfg().MustGet(ctx, "voiceChat")
	if !value.IsNil() {
		_ = value.Scan(&cfg)
	}
	applyConfigDefaults(&cfg)
	return cfg
}

// Handle 是语音对话主入口（带自动落 last talk 记录）。
func (s *VoiceService) Handle(ctx context.Context, deviceNo string, meta AudioMeta, audioBase64 string) ([]byte, AudioMeta, bool, error) {
	audio, outMeta, ask, answer, exit, _, err := s.HandleWithDialogue(ctx, deviceNo, meta, audioBase64)
	if err != nil {
		return nil, meta, false, err
	}
	if !exit {
		if err := s.persistTalkRecord(ctx, deviceNo, ask, answer); err != nil {
			if s.cfg.DebugLog {
				glog.Warningf(ctx, "[对话落库] 写入失败。deviceNo=%s ask=%q answer=%q err=%v", deviceNo, ask, answer, err)
			}
			return nil, meta, false, StageError{Stage: "device", Detail: err.Error()}
		}
		if s.cfg.DebugLog {
			glog.Infof(ctx, "[对话落库] 写入成功。deviceNo=%s ask=%q answer=%q", deviceNo, ask, answer)
		}
	}
	return audio, outMeta, exit, nil
}

// HandleWithDialogue 执行完整语音链路：设备校验 -> 参数校验 -> STT -> Chat -> TTS。
func (s *VoiceService) HandleWithDialogue(ctx context.Context, deviceNo string, meta AudioMeta, audioBase64 string) ([]byte, AudioMeta, string, string, bool, bool, error) {
	requestStart := time.Now()

	if err := s.ensureDeviceRegistered(ctx, deviceNo); err != nil {
		glog.Warningf(ctx, "[语音链路] 设备校验失败。deviceNo=%s err=%v", deviceNo, err)
		return nil, meta, "", "", false, false, StageError{Stage: "device", Detail: err.Error()}
	}

	if err := s.validateAudio(meta); err != nil {
		glog.Warningf(ctx, "[语音链路] 音频参数校验失败。deviceNo=%s err=%v", deviceNo, err)
		return nil, meta, "", "", false, false, StageError{Stage: "validate", Detail: err.Error()}
	}

	sttStart := time.Now()
	transcript, err := s.transcribe(ctx, meta, audioBase64)
	if err != nil {
		if isTimeoutErr(err) {
			glog.Warningf(ctx, "[语音链路] STT 阶段超时。deviceNo=%s 耗时=%s err=%v", deviceNo, time.Since(sttStart), err)
		} else {
			glog.Warningf(ctx, "[语音链路] STT 阶段失败。deviceNo=%s 耗时=%s err=%v", deviceNo, time.Since(sttStart), err)
		}
		return nil, meta, "", "", false, false, err
	}
	sttCost := time.Since(sttStart)
	if sttCost >= voiceStageSlowThreshold {
		glog.Warningf(ctx, "[语音链路] STT 阶段耗时偏长。deviceNo=%s 耗时=%s transcriptLen=%d", deviceNo, sttCost, utf8.RuneCountInString(strings.TrimSpace(transcript)))
	}

	chatStart := time.Now()
	reply, ask, exit, finishTalk, err := s.chatWithResult(ctx, deviceNo, transcript)
	if err != nil {
		if isTimeoutErr(err) {
			glog.Warningf(ctx, "[语音链路] 对话阶段超时。deviceNo=%s 耗时=%s err=%v", deviceNo, time.Since(chatStart), err)
		} else {
			glog.Warningf(ctx, "[语音链路] 对话阶段失败。deviceNo=%s 耗时=%s err=%v", deviceNo, time.Since(chatStart), err)
		}
		return nil, meta, "", "", false, false, err
	}
	chatCost := time.Since(chatStart)
	if chatCost >= voiceStageSlowThreshold {
		glog.Warningf(ctx, "[语音链路] 对话阶段耗时偏长。deviceNo=%s 耗时=%s exit=%v askLen=%d replyLen=%d", deviceNo, chatCost, exit, utf8.RuneCountInString(strings.TrimSpace(ask)), utf8.RuneCountInString(strings.TrimSpace(reply)))
	}

	if exit {
		// 当检测到退出意图时，不合成音频，直接返回 exit 标志
		totalCost := time.Since(requestStart)
		if totalCost >= voiceTotalSlowThreshold {
			glog.Warningf(ctx, "[语音链路] 退出请求总耗时偏长。deviceNo=%s 总耗时=%s", deviceNo, totalCost)
		}
		return nil, meta, ask, "", true, false, nil
	}

	ttsStart := time.Now()
	audio, err := s.synthesize(ctx, meta, reply)
	if err != nil {
		if isTimeoutErr(err) {
			glog.Warningf(ctx, "[语音链路] TTS 阶段超时。deviceNo=%s 耗时=%s err=%v", deviceNo, time.Since(ttsStart), err)
		} else {
			glog.Warningf(ctx, "[语音链路] TTS 阶段失败。deviceNo=%s 耗时=%s err=%v", deviceNo, time.Since(ttsStart), err)
		}
		return nil, meta, "", "", false, false, err
	}
	ttsCost := time.Since(ttsStart)
	if ttsCost >= voiceStageSlowThreshold {
		glog.Warningf(ctx, "[语音链路] TTS 阶段耗时偏长。deviceNo=%s 耗时=%s audioBytes=%d", deviceNo, ttsCost, len(audio))
	}
	totalCost := time.Since(requestStart)
	if totalCost >= voiceTotalSlowThreshold {
		glog.Warningf(ctx, "[语音链路] 请求总耗时偏长。deviceNo=%s 总耗时=%s", deviceNo, totalCost)
	}

	return audio, meta, ask, reply, false, finishTalk, nil
}

// HandleWithTranscript 直接基于已识别文本继续执行对话与TTS。
// 用于 WebSocket 流式模式下“ASR前置、后续逻辑复用文本”的链路。
func (s *VoiceService) HandleWithTranscript(ctx context.Context, deviceNo string, meta AudioMeta, transcript string) ([]byte, AudioMeta, string, string, bool, bool, error) {
	requestStart := time.Now()

	if err := s.ensureDeviceRegistered(ctx, deviceNo); err != nil {
		glog.Warningf(ctx, "[语音链路] 设备校验失败。deviceNo=%s err=%v", deviceNo, err)
		return nil, meta, "", "", false, false, StageError{Stage: "device", Detail: err.Error()}
	}

	chatStart := time.Now()
	reply, ask, exit, finishTalk, err := s.chatWithResult(ctx, deviceNo, transcript)
	if err != nil {
		if isTimeoutErr(err) {
			glog.Warningf(ctx, "[语音链路] 对话阶段超时。deviceNo=%s 耗时=%s err=%v", deviceNo, time.Since(chatStart), err)
		} else {
			glog.Warningf(ctx, "[语音链路] 对话阶段失败。deviceNo=%s 耗时=%s err=%v", deviceNo, time.Since(chatStart), err)
		}
		return nil, meta, "", "", false, false, err
	}
	chatCost := time.Since(chatStart)
	if chatCost >= voiceStageSlowThreshold {
		glog.Warningf(ctx, "[语音链路] 对话阶段耗时偏长。deviceNo=%s 耗时=%s exit=%v askLen=%d replyLen=%d", deviceNo, chatCost, exit, utf8.RuneCountInString(strings.TrimSpace(ask)), utf8.RuneCountInString(strings.TrimSpace(reply)))
	}

	if exit {
		totalCost := time.Since(requestStart)
		if totalCost >= voiceTotalSlowThreshold {
			glog.Warningf(ctx, "[语音链路] 退出请求总耗时偏长。deviceNo=%s 总耗时=%s", deviceNo, totalCost)
		}
		return nil, meta, ask, "", true, false, nil
	}

	ttsStart := time.Now()
	audio, err := s.synthesize(ctx, meta, reply)
	if err != nil {
		if isTimeoutErr(err) {
			glog.Warningf(ctx, "[语音链路] TTS 阶段超时。deviceNo=%s 耗时=%s err=%v", deviceNo, time.Since(ttsStart), err)
		} else {
			glog.Warningf(ctx, "[语音链路] TTS 阶段失败。deviceNo=%s 耗时=%s err=%v", deviceNo, time.Since(ttsStart), err)
		}
		return nil, meta, "", "", false, false, err
	}
	ttsCost := time.Since(ttsStart)
	if ttsCost >= voiceStageSlowThreshold {
		glog.Warningf(ctx, "[语音链路] TTS 阶段耗时偏长。deviceNo=%s 耗时=%s audioBytes=%d", deviceNo, ttsCost, len(audio))
	}
	totalCost := time.Since(requestStart)
	if totalCost >= voiceTotalSlowThreshold {
		glog.Warningf(ctx, "[语音链路] 请求总耗时偏长。deviceNo=%s 总耗时=%s", deviceNo, totalCost)
	}

	return audio, meta, ask, reply, false, finishTalk, nil
}

// TranscribeAudioRaw 仅执行音频参数校验与转写，不进入对话/TTS。
// 用于 WebSocket 流式模式下的“边收边识别（中间字幕）”。
func (s *VoiceService) TranscribeAudioRaw(ctx context.Context, meta AudioMeta, audioRaw []byte) (string, error) {
	if len(audioRaw) == 0 {
		return "", StageError{Stage: "stt", Detail: "音频为空"}
	}
	meta.Length = len(audioRaw)
	if err := s.validateAudio(meta); err != nil {
		return "", StageError{Stage: "validate", Detail: err.Error()}
	}
	audioBase64 := base64.StdEncoding.EncodeToString(audioRaw)
	text, err := s.transcribe(ctx, meta, audioBase64)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

// CreateStreamASRSession 创建“真流式 ASR”会话。
// 当前仅在 provider=baidu 且 streamEnabled=true 时可用。
func (s *VoiceService) CreateStreamASRSession(ctx context.Context, meta AudioMeta, onPartial func(text string), onFinal func(text string)) (StreamASRSession, error) {
	provider := strings.ToLower(strings.TrimSpace(s.cfg.STT.Provider))
	if provider != "baidu" {
		glog.Warningf(ctx, "[流式ASR] 创建会话前置检查失败：provider不支持。provider=%s streamEnabled=%v streamEndpoint=%q tokenEndpoint=%q sampleRate=%d bits=%d channels=%d", provider, s.cfg.STT.StreamEnabled, strings.TrimSpace(s.cfg.STT.StreamEndpoint), strings.TrimSpace(s.cfg.STT.TokenEndpoint), meta.SampleRate, meta.Bits, meta.Channels)
		return nil, StageError{Stage: "stt", Detail: "当前 provider 不支持流式 ASR"}
	}
	if !s.cfg.STT.StreamEnabled {
		glog.Warningf(ctx, "[流式ASR] 创建会话前置检查失败：stream未启用。provider=%s streamEnabled=%v streamEndpoint=%q tokenEndpoint=%q sampleRate=%d bits=%d channels=%d", provider, s.cfg.STT.StreamEnabled, strings.TrimSpace(s.cfg.STT.StreamEndpoint), strings.TrimSpace(s.cfg.STT.TokenEndpoint), meta.SampleRate, meta.Bits, meta.Channels)
		return nil, StageError{Stage: "stt", Detail: "未启用流式 ASR（stt.streamEnabled=false）"}
	}
	glog.Infof(ctx, "[流式ASR] 开始创建会话。provider=%s streamEndpoint=%q tokenEndpoint=%q cuid=%q devPid=%d format=%q timeoutSec=%d apiKeySet=%v apiSecretSet=%v sampleRate=%d bits=%d channels=%d", provider, strings.TrimSpace(s.cfg.STT.StreamEndpoint), strings.TrimSpace(s.cfg.STT.TokenEndpoint), strings.TrimSpace(s.cfg.STT.CUID), s.cfg.STT.DevPID, strings.TrimSpace(s.cfg.STT.Format), s.cfg.STT.TimeoutSeconds, strings.TrimSpace(s.cfg.STT.APIKey) != "", strings.TrimSpace(s.cfg.STT.APISecret) != "", meta.SampleRate, meta.Bits, meta.Channels)
	return newBaiduStreamASRSession(ctx, s, meta, onPartial, onFinal)
}

func EstimateBase64DecodedLenForDebug(data string) int {
	return estimateBase64DecodedLen(data)
}

// StreamRealtimeOptions 返回流式实时翻译的去抖与最小文本长度配置。
func (s *VoiceService) StreamRealtimeOptions() (time.Duration, int) {
	debounce := s.cfg.STT.RealtimeDebounceMs
	if debounce <= 0 {
		debounce = 1200
	}
	minRunes := s.cfg.STT.RealtimeMinRunes
	if minRunes <= 0 {
		minRunes = 4
	}
	return time.Duration(debounce) * time.Millisecond, minRunes
}

func DetectWaveHeaderForDebug(data string) (bool, string) {
	normalized, err := normalizeBase64Audio(data)
	if err != nil {
		return false, ""
	}
	headBytes, err := base64.StdEncoding.DecodeString(normalized)
	if err != nil || len(headBytes) < 12 {
		return false, ""
	}
	signature := string(headBytes[:4])
	format := string(headBytes[8:12])
	if signature == "RIFF" && format == "WAVE" {
		return true, "RIFF/WAVE"
	}
	return false, signature + "/" + format
}

// normalizeAndValidateChatText 统一清理并校验转写文本长度。
func (s *VoiceService) normalizeAndValidateChatText(text string) (string, error) {
	normalized := strings.TrimSpace(text)
	if normalized == "" {
		return "", StageError{Stage: "chat", Detail: "文本为空，无法进行聊天"}
	}

	if utf8.RuneCountInString(normalized) < s.cfg.DeepSeek.MinTextLength {
		return "", StageError{Stage: "chat", Detail: fmt.Sprintf("文本长度不能小于%d", s.cfg.DeepSeek.MinTextLength)}
	}

	return normalized, nil
}

// validateAudio 校验音频参数与大小/时长限制。
func (s *VoiceService) validateAudio(meta AudioMeta) error {
	if meta.SampleRate <= 0 {
		return errors.New("缺少有效的采样率")
	}
	if meta.Bits != 16 {
		return errors.New("仅支持 16-bit PCM")
	}
	if meta.Channels <= 0 {
		return errors.New("缺少有效的声道数")
	}
	if meta.Length <= 0 {
		return errors.New("缺少有效的音频长度")
	}

	if s.cfg.Audio.MaxSizeBytes > 0 && meta.Length > s.cfg.Audio.MaxSizeBytes {
		return fmt.Errorf("音频大小超过限制 %d bytes", s.cfg.Audio.MaxSizeBytes)
	}

	bytesPerSecond := meta.SampleRate * meta.Channels * (meta.Bits / 8)
	if bytesPerSecond > 0 && s.cfg.Audio.MaxDurationSec > 0 {
		maxBytes := bytesPerSecond * s.cfg.Audio.MaxDurationSec
		if meta.Length > maxBytes {
			return fmt.Errorf("音频时长超出限制 %d 秒", s.cfg.Audio.MaxDurationSec)
		}
	}

	return nil
}

// transcribe 根据配置分发到不同 STT 实现。
func (s *VoiceService) transcribe(ctx context.Context, meta AudioMeta, audioBase64 string) (string, error) {
	release := s.acquireLimiter(s.sttLimiter)
	defer release()

	switch strings.ToLower(s.cfg.STT.Provider) {
	case "baidu":
		return s.transcribeBaidu(ctx, meta, audioBase64)
	default:
		return s.transcribeGeneric(ctx, meta, audioBase64)
	}
}

// transcribeGeneric 通用 STT 调用（按二进制音频请求）。
func (s *VoiceService) transcribeGeneric(ctx context.Context, meta AudioMeta, audioBase64 string) (string, error) {
	pcm, err := decodeBase64Audio(audioBase64)
	if err != nil {
		return "", StageError{Stage: "stt", Detail: err.Error()}
	}
	if s.cfg.STT.Endpoint == "" {
		return "", StageError{Stage: "stt", Detail: "STT endpoint 未配置"}
	}

	cctx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.STT.TimeoutSeconds)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(cctx, http.MethodPost, s.cfg.STT.Endpoint, bytes.NewReader(pcm))
	if err != nil {
		return "", StageError{Stage: "stt", Detail: err.Error()}
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Audio-Sample-Rate", strconv.Itoa(meta.SampleRate))
	req.Header.Set("X-Audio-Bits", strconv.Itoa(meta.Bits))
	req.Header.Set("X-Audio-Channels", strconv.Itoa(meta.Channels))
	if s.cfg.STT.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.STT.APIKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", StageError{Stage: "stt", Detail: err.Error()}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", StageError{Stage: "stt", Detail: err.Error()}
	}
	if resp.StatusCode >= 300 {
		return "", StageError{Stage: "stt", Detail: fmt.Sprintf("status %d: %s", resp.StatusCode, string(body))}
	}

	text, err := extractText(body)
	if err != nil {
		return "", StageError{Stage: "stt", Detail: err.Error()}
	}
	if text == "" {
		return "", StageError{Stage: "stt", Detail: "转写结果为空"}
	}

	return text, nil
}

// transcribeBaidu 百度 STT 调用，支持 WAV data chunk 兼容处理。
func (s *VoiceService) transcribeBaidu(ctx context.Context, meta AudioMeta, audioBase64 string) (string, error) {
	if s.cfg.STT.Endpoint == "" {
		return "", StageError{Stage: "stt", Detail: "Baidu STT endpoint 未配置"}
	}
	token, err := s.getBaiduAccessToken(ctx, &s.sttToken, s.cfg.STT.APIKey, s.cfg.STT.APISecret, s.cfg.STT.TokenEndpoint, s.cfg.STT.TimeoutSeconds)
	if err != nil {
		return "", StageError{Stage: "stt", Detail: err.Error()}
	}

	normalizedB64, err := normalizeBase64Audio(audioBase64)
	if err != nil {
		return "", StageError{Stage: "stt", Detail: err.Error()}
	}
	raw, err := base64.StdEncoding.DecodeString(normalizedB64)
	if err != nil {
		return "", StageError{Stage: "stt", Detail: "音频 base64 解码失败: " + err.Error()}
	}
	pcmBytes, strippedWav := stripWavDataChunkIfPresent(raw)
	providedLen := meta.Length
	meta.Length = len(pcmBytes)
	if strippedWav {
		glog.Warningf(ctx, "stt baidu: stripped WAV data chunk, pcm bytes=%d (decoded input was %d)", len(pcmBytes), len(raw))
		audioBase64 = base64.StdEncoding.EncodeToString(pcmBytes)
	} else if providedLen > 0 && providedLen != len(raw) {
		glog.Warningf(ctx, "stt baidu len mismatch: provided=%d decoded=%d", providedLen, len(raw))
	}

	cctx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.STT.TimeoutSeconds)*time.Second)
	defer cancel()

	payload := map[string]interface{}{
		"format":  s.cfg.STT.Format,
		"rate":    meta.SampleRate,
		"channel": meta.Channels,
		"token":   token,
		"cuid":    s.cfg.STT.CUID,
		"len":     meta.Length,
		"speech":  audioBase64,
	}
	if s.cfg.STT.DevPID != 0 {
		payload["dev_pid"] = s.cfg.STT.DevPID
	}

	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(cctx, http.MethodPost, s.cfg.STT.Endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", StageError{Stage: "stt", Detail: err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", StageError{Stage: "stt", Detail: err.Error()}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", StageError{Stage: "stt", Detail: err.Error()}
	}
	if resp.StatusCode >= 300 {
		return "", StageError{Stage: "stt", Detail: fmt.Sprintf("status %d: %s", resp.StatusCode, string(respBody))}
	}

	var result struct {
		ErrNo  int      `json:"err_no"`
		ErrMsg string   `json:"err_msg"`
		Res    []string `json:"result"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", StageError{Stage: "stt", Detail: err.Error()}
	}
	if result.ErrNo != 0 {
		return "", StageError{Stage: "stt", Detail: fmt.Sprintf("baidu err %d: %s", result.ErrNo, result.ErrMsg)}
	}
	if len(result.Res) == 0 {
		return "", StageError{Stage: "stt", Detail: "转写结果为空"}
	}

	text := strings.TrimSpace(result.Res[0])
	if isBaiduSTTPlaceholderTranscript(text) {
		snippet := string(respBody)
		if len(snippet) > 512 {
			snippet = snippet[:512] + "..."
		}
		glog.Warningf(ctx, "stt baidu: result text is placeholder (%q), raw=%s", text, snippet)
		return "", StageError{Stage: "stt", Detail: fmt.Sprintf("百度返回异常转写文本 %q（非语音内容），请检查音频格式/长度/采样率或 dev_pid", text)}
	}

	return text, nil
}

func (s *VoiceService) chat(ctx context.Context, deviceNo, transcript string) (string, error) {
	reply, _, _, _, err := s.chatWithResult(ctx, deviceNo, transcript)
	return reply, err
}

// TextChat 文本对话：在校验设备已注册后调用 chat，与 HandleWithTranscript 中对话阶段一致（不经 STT/TTS）。
func (s *VoiceService) TextChat(ctx context.Context, deviceNo, transcript string) (string, error) {
	if err := s.ensureDeviceRegistered(ctx, deviceNo); err != nil {
		return "", err
	}
	return s.chat(ctx, deviceNo, transcript)
}

// chatWithResult 对话核心流程：
// - 处理退出意图
// - 处理待补量词
// - 处理“成长建议”特殊意图
// - 调用 DeepSeek 进行结构化事件识别并落库
func (s *VoiceService) chatWithResult(ctx context.Context, deviceNo, transcript string) (string, string, bool, bool, error) {
	normalizedTranscript, err := s.normalizeAndValidateChatText(transcript)
	if err != nil {
		return "", "", false, false, err
	}
	events := []entity.Event{}
	dao.Event.Ctx(ctx).Scan(&events)
	// 取数据库中的预设关键词的集合（dao.Action）
	actions := []entity.Action{}
	dao.Action.Ctx(ctx).Scan(&actions)
	exit := false

	// 获取上一次的对话缓存中，我回答的最后一条记录
	now := time.Now()
	lastUserMessage := ""
	if strings.TrimSpace(deviceNo) != "" {
		s.sessionMu.Lock()
		s.pruneSessionsLocked(now)
		if sess, ok := s.sessions[deviceNo]; ok {
			if !s.isExpired(sess.LastActive, now) {
				historyMessages := sess.Messages
				if len(historyMessages) > 1 {
					lastUserMessage = historyMessages[len(historyMessages)-1].Content
				}
			} else {
				delete(s.sessions, deviceNo)
				s.deviceLocks.Delete(deviceNo)
			}
		}
		s.sessionMu.Unlock()
	}

	// 上一次的对话缓存中，我回答的最后一条记录，是否包含"多少"关键词
	mayReplayQuantity := false
	if strings.Contains(lastUserMessage, quantityKeyword) {
		mayReplayQuantity = true
	}

	// 获取这一次对话中的数量
	number, ok := extractNumberFromText(normalizedTranscript)
	if ok {
		// 上一次对话中如果包含"多少"关键词，则需要判断是要将上一次对话中的"多少"改为这一次的会话内容，然后走下面的逻辑
		if mayReplayQuantity {
			normalizedTranscript = strings.Replace(lastUserMessage, quantityKeyword, "? "+strconv.Itoa(number)+"。", 1)
			// 日志打印
			glog.Infof(ctx, "上一次对话中包含\"多少\"关键词，将\"多少\"改为这一次的会话内容。lastUserMessage=%q normalizedTranscript=%q", lastUserMessage, number)
		}
	}

	// 打印normalizedTranscript
	glog.Infof(ctx, "问题=%q", normalizedTranscript)

	// 先将动作按名称长度从长到短排序
	sort.Slice(actions, func(i, j int) bool {
		return len(actions[i].Name) > len(actions[j].Name)
	})
	// 判断文本是否包含预设动作关键词
	for _, action := range actions {
		if strings.Contains(normalizedTranscript, action.Name) {
			// 打印日志命中动作
			glog.Infof(ctx, "命中动作: %s", action.Name)
			finalReply, exit, finishTalk, err := s.handleActionRecord(ctx, deviceNo, normalizedTranscript, action, events)
			if err != nil {
				// 处理动作失败,可能动作解析错误,尝试解析出新的动作,再走命中事件流程
				continue
			}
			// 往QA里录入问题和答案
			s.insertQa(ctx, normalizedTranscript, finalReply)
			return finalReply, normalizedTranscript, exit, finishTalk, err
		}
	}
	// 没有命中预设动作, 请求deepSeek分析文案中的预设动作,并落库后，再走命中事件流程
	action, err := s.callDeepSeekActionExtract(ctx, deviceNo, normalizedTranscript)
	if err != nil {
		// 认为是对话动作
		finalReply, finishTalk, err := s.handleIntentGeneral(ctx, deviceNo, normalizedTranscript)
		if err == nil {
			// 往QA里录入问题和答案
			s.insertQa(ctx, normalizedTranscript, finalReply)
		}
		return finalReply, normalizedTranscript, exit, finishTalk, err
	}
	// 打印日志命中动作
	glog.Infof(ctx, "没有命中预设动作, 请求deepSeek分析文案中的预设动作,并落库后，再走命中事件流程,命中动作: %s", action.Name)
	finalReply, exit, finishTalk, err := s.handleActionRecord(ctx, deviceNo, normalizedTranscript, action, events)
	if err == nil {
		// 往QA里录入问题和答案
		s.insertQa(ctx, normalizedTranscript, finalReply)
	}
	return finalReply, normalizedTranscript, exit, finishTalk, err
}

// 往qa里录入问题和答案
func (s *VoiceService) insertQa(ctx context.Context, question, answer string) error {
	// 如果已经存在相同的问题,则更新对应问题的命中次数+1
	existingQa := entity.Qa{}
	dao.Qa.Ctx(ctx).Where("question", question).Scan(&existingQa)
	if existingQa.Id > 0 {
		_, err := dao.Qa.Ctx(ctx).Where("question", question).Update(g.Map{"attack": existingQa.Attack + 1})
		return err
	}
	_, err := dao.Qa.Ctx(ctx).Insert(entity.Qa{
		Question: question,
		Replay:   answer,
	})
	return err
}

// 根据文本,请求deepSeek分析文案中的动作是什么,并判断该动作的目标类型(ActionTargetTypeStart,ActionTargetTypeEnd,ActionTargetTypeOne,ActionTargetTypeExit,ActionTargetTypeSuggest,ActionTargetTypeSearch),输出JSON:{"action":"动作名称","target_type":"目标类型"}
func (s *VoiceService) callDeepSeekActionExtract(ctx context.Context, deviceNo, transcript string) (entity.Action, error) {
	prompt := fmt.Sprintf("输入：%s", transcript)
	systemMessage := fmt.Sprintf(
		`你是一个主要记录母婴喂养的助手且具备精准的动作提取能力，严格按指定JSON格式输出，不添加任何解释。

动作名称提取：从输入文本中提取代表性的连续文案,至少两个字。

目标类型选择：
- %s(%s)：开始计时
- %s(%s)：结束计时
- %s(%s)：一次性记录
- %s(%s)：退出
- %s(%s)：成长建议
- %s(%s)：搜索
- %s(%s)：对话

特别规则：
1. 睡眠事件：睡着→%s，睡醒→%s
2. 只有动作无事件→%s或%s
3. 关于孩子的问题→%s
4. 关于历史数据的问题→%s

输出格式：{"name":"动作名称","targetType":"目标类型"}
无法判断时：{"name":"","targetType":""}`,
		ActionTargetTypeStart, ActionTargetTypeChinese(ActionTargetTypeStart),
		ActionTargetTypeEnd, ActionTargetTypeChinese(ActionTargetTypeEnd),
		ActionTargetTypeOne, ActionTargetTypeChinese(ActionTargetTypeOne),
		ActionTargetTypeExit, ActionTargetTypeChinese(ActionTargetTypeExit),
		ActionTargetTypeSuggest, ActionTargetTypeChinese(ActionTargetTypeSuggest),
		ActionTargetTypeSearch, ActionTargetTypeChinese(ActionTargetTypeSearch),
		ActionTargetTypeConversation, ActionTargetTypeChinese(ActionTargetTypeConversation),
		ActionTargetTypeStart, ActionTargetTypeEnd,
		ActionTargetTypeConversation, ActionTargetTypeExit,
		ActionTargetTypeSuggest, ActionTargetTypeSearch)
	raw, _, _, err := s.callDeepSeekRaw(ctx, deviceNo, prompt, 0, systemMessage)
	if err != nil {
		return entity.Action{}, err
	}
	parsed := entity.Action{}
	err = json.Unmarshal([]byte(raw), &parsed)
	if err != nil {
		return entity.Action{}, err
	}
	if parsed.TargetType == ActionTargetTypeConversation || parsed.TargetType == "" {
		return entity.Action{}, errors.New("对话动作不需要落库")
	} else {
		// 如果动作名为空,则为输入的文本值
		if parsed.Name == "" {
			parsed.Name = transcript
		}

		// 将动作落库,保证动作名称唯一
		existingAction := entity.Action{}
		dao.Action.Ctx(ctx).Where("name", parsed.Name).Scan(&existingAction)
		if existingAction.Id > 0 {
			return entity.Action{}, errors.New("动作名称已存在")
		}
		dao.Action.Ctx(ctx).Insert(parsed)
		return parsed, nil
	}
}

// 根据动作，判断后续逻辑
func (s *VoiceService) handleActionRecord(ctx context.Context, deviceNo string, normalizedTranscript string, action entity.Action, events []entity.Event) (finalReply string, exit bool, finishTalk bool, err error) {
	// 根据不同的action做出不同的处理
	nowTime := time.Now().Format("2006-01-02 15:04:05")
	switch action.TargetType {
	case "start": //开始记录计时动作
		event, targetName, ok := s.extractEventFromText(ctx, normalizedTranscript, events)
		if !ok { // 没有命中事件名，交给deepseek分析文案中的事件名
			// 交给deepseek分析文案中的事件名,并落库后，再走命中事件流程
			event, targetName, err = s.callDeepSeekEntityExtract(ctx, deviceNo, normalizedTranscript)
			if err != nil {
				glog.Warningf(ctx, "调用 DeepSeek 进行实体抽取失败，deviceNo=%s transcript=%q err=%v", deviceNo, normalizedTranscript, err)
				return "我听不懂你说的事件,请用具体的名称告诉我", false, false, err
			}
			// 打印日志命中事件
			glog.Infof(ctx, "没有命中事件名, 请求deepSeek分析文案中的事件名,并落库后，再走命中事件流程,命中事件: %s", targetName)
		}
		_, err = dao.History.Ctx(ctx).Insert(entity.History{
			DeviceNo:  deviceNo,
			EventId:   event.Id,
			EventName: targetName,
			StartTime: nowTime,
			Remark:    normalizedTranscript,
		})
		if err != nil {
			return "记录失败,请重试", false, true, err
		}
		finalReply = fmt.Sprintf("好的，已记录%s开始", targetName)
		finishTalk = true
		return finalReply, false, finishTalk, nil
	case "end": //结束记录计时动作，自动补结束时间
		event, targetName, ok := s.extractEventFromText(ctx, normalizedTranscript, events)
		if !ok { // 没有命中事件名，交给deepseek分析文案中的事件名
			// 交给deepseek分析文案中的事件名,并落库后，再走命中事件流程
			event, targetName, err = s.callDeepSeekEntityExtract(ctx, deviceNo, normalizedTranscript)
			if err != nil {
				glog.Warningf(ctx, "调用 DeepSeek 进行实体抽取失败，deviceNo=%s transcript=%q err=%v", deviceNo, normalizedTranscript, err)
				return "我听不懂你说的事件,请用具体的名称告诉我", false, false, err
			}
			// 打印日志命中事件
			glog.Infof(ctx, "没有命中事件名, 请求deepSeek分析文案中的事件名,并落库后，再走命中事件流程,命中事件: %s", targetName)
		}

		// 判断最近一次事件是否是同一事件
		lastEvent := entity.History{}
		dao.History.Ctx(ctx).Where("device_no", deviceNo).Order("id DESC").Limit(1).Scan(&lastEvent)
		if lastEvent.EventId == event.Id {
			// 是同一事件，则更新结束时间
			_, err = dao.History.Ctx(ctx).Where("device_no", deviceNo).Where("event_id", event.Id).Update(g.Map{"end_time": nowTime})
			if err != nil {
				return "更新结束时间失败,请重试", false, true, err
			}
			finalReply = fmt.Sprintf("好的，已记录%s结束", targetName)
			finishTalk = true
			return finalReply, false, finishTalk, nil
		} else {
			// 不是同一事件

			// 则插入新的记录
			_, err = dao.History.Ctx(ctx).Insert(entity.History{
				DeviceNo:  deviceNo,
				EventId:   event.Id,
				EventName: targetName,
				StartTime: nowTime,
				EndTime:   nowTime,
				Remark:    normalizedTranscript,
			})
			if err != nil {
				return "记录事件失败,请重试", false, true, err
			}
			// 上一件事如果没有结束时间,则告知用户上一件事自动结束
			if lastEvent.EndTime == "" {
				dao.History.Ctx(ctx).Where("device_no", deviceNo).Where("event_id", lastEvent.EventId).Update(g.Map{"end_time": nowTime})
				if err != nil {
					return fmt.Sprintf("好的，已记录%s结束，%s结束失败,请手动结束", targetName, lastEvent.EventName), false, true, err
				}
				finalReply = fmt.Sprintf("好的，已记录%s结束，%s自动结束", targetName, lastEvent.EventName)
			} else {
				finalReply = fmt.Sprintf("好的，已记录%s结束", targetName)
			}
			finishTalk = true
			return finalReply, false, finishTalk, nil
		}
	case "one": //记录一次性动作，记录一次
		event, targetName, ok := s.extractEventFromText(ctx, normalizedTranscript, events)
		if !ok { // 没有命中事件名，交给deepseek分析文案中的事件名
			// 交给deepseek分析文案中的事件名,并落库后，再走命中事件流程
			event, targetName, err = s.callDeepSeekEntityExtract(ctx, deviceNo, normalizedTranscript)
			if err != nil {
				glog.Warningf(ctx, "调用 DeepSeek 进行实体抽取失败，deviceNo=%s transcript=%q err=%v", deviceNo, normalizedTranscript, err)
				return "我听不懂你说的事件,请用具体的名称告诉我", false, false, err
			}
			// 打印日志命中事件
			glog.Infof(ctx, "没有命中事件名, 请求deepSeek分析文案中的事件名,并落库后，再走命中事件流程,命中事件: %s", targetName)
		}
		if event.NeedQuantity > 0 {
			quantity, ok := extractNumberFromText(normalizedTranscript)
			if !ok || quantity <= 0 {
				finalReply = "请问 " + action.Name + " " + targetName + " 的数量是" + quantityKeyword
				finishTalk = false
				return finalReply, false, finishTalk, nil
			}
			_, err = dao.History.Ctx(ctx).Insert(entity.History{
				DeviceNo:    deviceNo,
				EventId:     event.Id,
				EventName:   targetName,
				EventNumber: int64(quantity),
				StartTime:   nowTime,
				EndTime:     nowTime,
				Remark:      normalizedTranscript,
			})
			if err != nil {
				return "记录事件失败,请重试", false, true, err
			}
			finalReply = fmt.Sprintf("好的，已记录 %s %d", targetName, quantity)
			finishTalk = true
			return finalReply, false, finishTalk, nil
		} else {
			_, err = dao.History.Ctx(ctx).Insert(entity.History{
				DeviceNo:    deviceNo,
				EventId:     event.Id,
				EventName:   targetName,
				EventNumber: 1,
				StartTime:   nowTime,
				EndTime:     nowTime,
				Remark:      normalizedTranscript,
			})
			if err != nil {
				return "记录事件失败,请重试", false, true, err
			}
			finalReply = fmt.Sprintf("好的，已记录 %s", targetName)
			finishTalk = true
			return finalReply, false, finishTalk, nil
		}
	case "suggest": //成长建议动作
		reply, handleErr := s.callDeepSeekGrowthSuggestion(ctx, deviceNo)
		if handleErr != nil {
			return "获取成长建议失败,请重试", false, true, handleErr
		}
		finalReply = strings.TrimSpace(reply)
		finishTalk = true
		return finalReply, false, finishTalk, nil
	case "search": //搜索动作

		reply, handleErr := s.callDeepSeekHistoryReply(ctx, deviceNo, normalizedTranscript, 12)
		if handleErr != nil {
			return "获取历史记录失败,请重试", false, true, handleErr
		}
		finalReply = strings.TrimSpace(reply)
		finishTalk = true
		return finalReply, false, finishTalk, nil
	case "exit": //退出动作
		return "好的，再见", true, false, nil
	default:
		return "我没有理解你的意思", false, false, nil
	}
}

// 提取文本中的事件对象
func (s *VoiceService) extractEventFromText(ctx context.Context, normalizedTranscript string, events []entity.Event) (entity.Event, string, bool) {
	for _, event := range events {
		// 原事件名称为部分匹配
		if hasSignificantOverlap(normalizedTranscript, event.Name) {
			// 打印命中事件名
			glog.Infof(ctx, "命中事件名: %s", event.Name)
			return event, event.Name, true
		}
		// 额外名称匹配为包含全量匹配，而不是部分匹配
		if event.ExtraNames != "" {
			extraNames := strings.Split(event.ExtraNames, ",")
			for _, extraName := range extraNames {
				if strings.Contains(normalizedTranscript, extraName) && extraName != "" {
					// 打印命中额外名称
					glog.Infof(ctx, "命中额外名称: %s", extraName)
					return event, extraName, true
				}
			}
		}
	}
	return entity.Event{}, "", false
}

// 提取文本中的数量值
func extractNumberFromText(text string) (int, bool) {

	// 把text中的一、二、三、四、五、六、七、八、九转换为1、2、3、4、5、6、7、8、9
	text = strings.ReplaceAll(text, "一", "1")
	text = strings.ReplaceAll(text, "二", "2")
	text = strings.ReplaceAll(text, "三", "3")
	text = strings.ReplaceAll(text, "四", "4")
	text = strings.ReplaceAll(text, "五", "5")
	text = strings.ReplaceAll(text, "六", "6")
	text = strings.ReplaceAll(text, "七", "7")
	text = strings.ReplaceAll(text, "八", "8")
	text = strings.ReplaceAll(text, "九", "9")

	text = strings.TrimSpace(text)
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(text)
	if match == "" {
		return 0, false
	}
	value, err := strconv.Atoi(match)
	if err != nil {
		return 0, false
	}
	return value, true
}

// hasSignificantOverlap 判断两个文本是否有显著的交集（至少两个连续字符）。考虑到用户说D3的情况
func hasSignificantOverlap(text, keyword string) bool {
	textRunes := []rune(text)
	keywordRunes := []rune(keyword)
	if len(textRunes) < 2 || len(keywordRunes) < 2 {
		return false
	}
	for i := 0; i < len(textRunes)-1; i++ {
		for j := 0; j < len(keywordRunes)-1; j++ {
			if textRunes[i] == keywordRunes[j] && textRunes[i+1] == keywordRunes[j+1] {
				return true
			}
		}
	}
	return false
}

// parseEventIntentFromReply 从模型回复中提取结构化 JSON 意图。
func parseEventIntentFromReply(reply string) (eventIntentResult, bool) {
	intent := eventIntentResult{}
	trimmed := strings.TrimSpace(reply)
	if trimmed == "" {
		return intent, false
	}
	trimmed = normalizeIntentCandidateText(trimmed)
	if trimmed == "" {
		return intent, false
	}
	if err := json.Unmarshal([]byte(trimmed), &intent); err == nil {
		intent.Action = strings.ToLower(strings.TrimSpace(intent.Action))
		intent.EventName = strings.TrimSpace(intent.EventName)
		intent.ActionKeyWord = strings.TrimSpace(intent.ActionKeyWord)
		intent.Remark = strings.TrimSpace(intent.Remark)
		intent.Reply = strings.TrimSpace(intent.Reply)
		intent.Reason = strings.TrimSpace(intent.Reason)
		if intent.Action == "" {
			intent.Action = "none"
		}
		return intent, true
	}
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		segment := trimmed[start : end+1]
		if err := json.Unmarshal([]byte(segment), &intent); err == nil {
			intent.Action = strings.ToLower(strings.TrimSpace(intent.Action))
			intent.EventName = strings.TrimSpace(intent.EventName)
			intent.ActionKeyWord = strings.TrimSpace(intent.ActionKeyWord)
			intent.Remark = strings.TrimSpace(intent.Remark)
			intent.Reply = strings.TrimSpace(intent.Reply)
			intent.Reason = strings.TrimSpace(intent.Reason)
			if intent.Action == "" {
				intent.Action = "none"
			}
			return intent, true
		}
	}
	return eventIntentResult{}, false
}

func normalizeIntentCandidateText(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSuffix(strings.TrimSpace(trimmed), "```")
	}
	// 兼容模型返回包裹格式：{"role":"assistant","content":"{...intent...}"}
	var wrapper map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &wrapper); err == nil {
		if content, ok := wrapper["content"].(string); ok {
			content = strings.TrimSpace(content)
			if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
				return content
			}
		}
	}
	return trimmed
}

func parseGeneralChatResult(reply string) (generalChatResult, bool) {
	var out generalChatResult
	trimmed := normalizeIntentCandidateText(reply)
	if strings.TrimSpace(trimmed) == "" {
		return out, false
	}
	if err := json.Unmarshal([]byte(trimmed), &out); err == nil {
		out.Reply = strings.TrimSpace(out.Reply)
		return out, true
	}
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(trimmed[start:end+1]), &out); err == nil {
			out.Reply = strings.TrimSpace(out.Reply)
			return out, true
		}
	}
	return out, false
}

// 普通对话,无用,先保留
func (s *VoiceService) handleIntentGeneral(ctx context.Context, deviceNo, transcript string) (string, bool, error) {
	// 将用户的生日和性别作为宝宝的信息上下文输入，辅助模型判断是否需要回复以及回复内容
	birthday, gender := s.loadDeviceProfile(ctx, deviceNo)
	systemMessage := fmt.Sprintf("用户的宝宝出生日期是%s，性别是%s。你还是一个语音助手，可以解答任何问题,回应内容控制在20字以内。", birthday, gender)
	prompt := fmt.Sprintf("用户输入=%s。请仅输出JSON：{\"reply\":\"回复内容\",\"need_user_reply\":true|false}。", transcript)
	raw, reply, _, err := s.callDeepSeekRaw(ctx, deviceNo, prompt, 5, systemMessage)
	if err != nil {
		return "", false, err
	}
	if parsed, ok := parseGeneralChatResult(raw); ok {
		if s.cfg.DebugLog {
			parsedJSON, _ := json.Marshal(parsed)
			glog.Infof(ctx, "[思考过程] 其它问答结构化解析成功。deviceNo=%s parsed=%s", deviceNo, string(parsedJSON))
		}
		if parsed.Reply == "" {
			parsed.Reply = reply
		}
		return strings.TrimSpace(parsed.Reply), !parsed.NeedUserReply, nil
	}
	if s.cfg.DebugLog {
		glog.Warningf(ctx, "[思考过程] 其它问答结构化解析失败，使用文本回复。deviceNo=%s raw=%s", deviceNo, truncateVoiceLogText(raw, 800))
	}
	return strings.TrimSpace(reply), true, nil
}

func (s *VoiceService) callDeepSeekEntityExtract(ctx context.Context, deviceNo, transcript string) (entity.Event, string, error) {
	out := entity.Event{}

	// deepseek需要分析文本中是否有与原来事件列表中的名称相符的事件类型,如果有提取当前的事件名称并输出json:{"name":"原表中的事件名","extra_name":"当前事件名称"},否则并判断是否需要计数（1表示需要，0表示不需要）输出json:{"name":当前事件名,"extra_name":"","need_quantity":"0或1"}。如果无法确定事件名称，则输出：{\"name\":\"\",\"need_quantity\":\"0\"}"
	// 将数据库中的事件名称拼接起来,用逗号分隔,然后告诉deepseek,事件名称有:xxx,xxx,xxx
	eventList := []entity.Event{}
	dao.Event.Ctx(ctx).Fields(dao.Event.Columns().Name).Scan(&eventList)
	eventNamesStr := ""
	for _, event := range eventList {
		eventNamesStr += event.Name + ","
	}
	systemMessage := fmt.Sprintf(`你是一个精准的事件提取器，严格输出JSON。

事件列表：%s

特别规则：
1. 扩展词从文本提取连续文案
2. 吃奶事件：如无法区分母乳/配方奶，输出{"name":"","extraNames":"","need_quantity":"0"}
3. 不是想记录事件:输出{"name":"","extraNames":"","need_quantity":"0"}

输出规则：
1. 匹配事件列表 → {"name":"原事件名","extraNames":"扩展词"}
2. 不匹配但可识别 → {"name":"新事件名","extraNames":"","need_quantity":"0或1"}
3. 无法确定 → {"name":"","need_quantity":"0"}`, eventNamesStr)

	prompt := fmt.Sprintf("输入=%s。按规则分析并输出JSON。", transcript)
	// 你需要从事件列表中,查看是否有符合的事件类型,如果有则直接返回列表中的事件类型,如果没有则需要从文本中提取事件名称。
	raw, _, _, err := s.callDeepSeekRaw(ctx, deviceNo, prompt, 0, systemMessage)
	if err != nil {
		return out, "", err
	}

	parsed := entity.Event{}
	trimmed := normalizeIntentCandidateText(raw)
	if unmarshalErr := json.Unmarshal([]byte(trimmed), &parsed); unmarshalErr != nil {
		start := strings.Index(trimmed, "{")
		end := strings.LastIndex(trimmed, "}")
		if start >= 0 && end > start {
			if nestedErr := json.Unmarshal([]byte(trimmed[start:end+1]), &parsed); nestedErr != nil {
				return out, "", fmt.Errorf("解析实体抽取结果失败: %w", unmarshalErr)
			}
		} else {
			return out, "", fmt.Errorf("解析实体抽取结果失败: %w", unmarshalErr)
		}
	}

	name := strings.TrimSpace(parsed.Name)
	if name == "" {
		name = strings.TrimSpace(parsed.Name)
	}
	if name == "" {
		return out, "", errors.New("未抽取到事件名称")
	}

	out.Name = name
	out.ExtraNames = parsed.ExtraNames
	// 获取原来的事件的额外名称
	oldEvent := entity.Event{}
	dao.Event.Ctx(ctx).Where("name", name).Scan(&oldEvent)

	targetName := out.ExtraNames
	if oldEvent.Id > 0 {
		out.NeedQuantity = oldEvent.NeedQuantity
		if out.ExtraNames == "" {
			// 遇到重复的事件名
			return out, out.Name, errors.New("事件名称已存在")
		} else {
			// 将新的额外名称添加到原来的额外名称中	,且避免重复
			extraNames := strings.Split(oldEvent.ExtraNames, ",")
			for _, extraName := range extraNames {
				if extraName == out.ExtraNames {
					return out, extraName, errors.New("事件名称已存在")
				}
			}

			if len(oldEvent.ExtraNames) > 0 {
				out.ExtraNames = strings.Join(extraNames, ",") + "," + out.ExtraNames
			}

			// 更新原来的事件表中的额外名称
			dao.Event.Ctx(ctx).Where("name", name).
				Update(g.Map{"extra_names": out.ExtraNames})
		}
	} else {
		targetName = out.Name
		out.NeedQuantity = parsed.NeedQuantity
		// 将抽取到的事件新增到事件表中。
		dao.Event.Ctx(ctx).Insert(&out)
	}

	return out, targetName, nil
}

// matchEventByName 使用宽松规则将模型输出事件名映射到库内事件。
func (s *VoiceService) matchEventByName(events []eventInfo, eventName string) (eventInfo, bool) {
	needle := strings.TrimSpace(strings.ToLower(eventName))
	if needle == "" {
		return eventInfo{}, false
	}
	for _, ev := range events {
		name := strings.TrimSpace(strings.ToLower(ev.Name))
		if name == needle || strings.Contains(name, needle) || strings.Contains(needle, name) {
			return ev, true
		}
	}
	return eventInfo{}, false
}

func (s *VoiceService) setPendingQuantity(deviceNo string, state pendingQuantityState) {
	// 记录“该设备下一轮需要补充数量”的状态。
	// 这个状态是内存态，服务重启后会丢失（符合当前设计预期）。
	s.pendingQuantityMu.Lock()
	defer s.pendingQuantityMu.Unlock()
	if strings.TrimSpace(deviceNo) == "" {
		return
	}
	s.pendingQuantity[deviceNo] = state
}

func (s *VoiceService) clearPendingQuantity(deviceNo string) {
	// 清理待补量词状态：
	// - 已成功补录数量后
	// - 或当前流程判断不再需要补录时
	s.pendingQuantityMu.Lock()
	defer s.pendingQuantityMu.Unlock()
	delete(s.pendingQuantity, deviceNo)
}

func (s *VoiceService) getPendingQuantity(deviceNo string) (pendingQuantityState, bool) {
	// 读取指定设备是否存在待补量词上下文。
	s.pendingQuantityMu.Lock()
	defer s.pendingQuantityMu.Unlock()
	state, ok := s.pendingQuantity[deviceNo]
	return state, ok
}

func (s *VoiceService) hasPendingQuantity(deviceNo string) bool {
	_, ok := s.getPendingQuantity(deviceNo)
	return ok
}

// appendRemark 合并备注内容，避免覆盖已有备注。
func appendRemark(existing, extra string) string {
	existing = strings.TrimSpace(existing)
	extra = strings.TrimSpace(extra)
	if existing == "" {
		return extra
	}
	if extra == "" {
		return existing
	}
	return existing + "；" + extra
}

func autoTimedQuantityFromRange(startTime, endTime string) int {
	startAt, okStart := parseDBTime(startTime)
	endAt, okEnd := parseDBTime(endTime)
	if !okStart || !okEnd || endAt.Before(startAt) {
		return 0
	}
	seconds := int(math.Round(endAt.Sub(startAt).Seconds()))
	if seconds < 0 {
		return 0
	}
	return seconds
}

func parseDBTime(value string) (time.Time, bool) {
	v := strings.TrimSpace(value)
	if v == "" {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", v, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// isGrowthSuggestionIntent 判断用户是否在询问成长建议。
func (s *VoiceService) isGrowthSuggestionIntent(text string) bool {
	keys := []string{"成长建议", "生长建议", "发育建议", "育儿建议", "喂养建议", "孩子建议", "建议"}
	for _, key := range keys {
		if strings.Contains(text, key) {
			return true
		}
	}
	return false
}

// getOrCreateDailySuggestion 每设备每天只生成一次建议，其余请求直接复用当日记录。
func (s *VoiceService) getOrCreateDailySuggestion(ctx context.Context, deviceNo string) (string, error) {
	todayPrefix := time.Now().Format("2006-01-02")
	var existing entity.Suggest
	err := dao.Suggest.Ctx(ctx).
		Where(dao.Suggest.Columns().DeviceNo, deviceNo).
		WhereLike(dao.Suggest.Columns().Time, todayPrefix+"%").
		OrderDesc(dao.Suggest.Columns().Id).
		Limit(1).
		Scan(&existing)
	if err == nil && existing.Id > 0 && strings.TrimSpace(existing.Suggest) != "" {
		return existing.Suggest, nil
	}

	suggestion, callErr := s.callDeepSeekGrowthSuggestion(ctx, deviceNo)
	if callErr != nil {
		return "", callErr
	}
	suggestion = strings.TrimSpace(suggestion)
	if suggestion == "" {
		suggestion = "今天建议保持规律作息，按需喂养，关注精神状态与排便情况。"
	}

	_, insertErr := dao.Suggest.Ctx(ctx).Data(g.Map{
		dao.Suggest.Columns().DeviceNo: deviceNo,
		dao.Suggest.Columns().Suggest:  suggestion,
		dao.Suggest.Columns().Time:     nowText(),
	}).Insert()
	if insertErr != nil {
		glog.Warningf(ctx, "insert suggest failed: %v", insertErr)
	}
	return suggestion, nil
}

func (s *VoiceService) loadDeviceProfile(ctx context.Context, deviceNo string) (string, string) {
	var user entity.User
	_ = dao.User.Ctx(ctx).
		Fields(dao.User.Columns().Birthday, dao.User.Columns().Sex).
		Where(dao.User.Columns().DeviceNo, deviceNo).
		Limit(1).
		Scan(&user)
	birthday := strings.TrimSpace(user.Birthday)
	if birthday == "" {
		birthday = "未设置"
	}
	sexText := "女"
	if user.Sex > 0 {
		sexText = "男"
	}
	return birthday, sexText
}

const growthSuggestUserPrompt = "你是育儿助手。请根据提供的育儿历史记录，给出今日成长建议（100字以内，实用、温和）。"
const growthSuggestDataHint = "下列 JSON 含 child_info 与 history。你必须严格依据这些记录作答，不要编造未出现的情节。\n输出为一段纯中文正文，不要 Markdown、不要输出 JSON。"

// growthSuggestPayload 成长建议专用 DeepSeek 请求体（含 child_info / history）。
type growthSuggestPayload struct {
	Messages  []growthSuggestMessage   `json:"messages"`
	Model     string                   `json:"model"`
	Stream    bool                     `json:"stream"`
	ChildInfo growthSuggestChildInfo   `json:"child_info"`
	History   []map[string]interface{} `json:"history"`
}

type growthSuggestMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type growthSuggestChildInfo struct {
	Birthday string `json:"birthday"`
	Gender   string `json:"gender"`
}

func defaultSystemPromptForSuggest(cfg VoiceChatConfig) string {
	p := strings.TrimSpace(cfg.DeepSeek.SystemPrompt)
	if p == "" {
		return "你是语音助手。"
	}
	return p
}

func birthdayForSuggestAPI(birthday string) string {
	b := strings.TrimSpace(birthday)
	if b == "" || b == "未设置" {
		return ""
	}
	return b
}

func suggestDurationMinutes(startStr, endStr string) int {
	startAt, ok1 := parseDBTime(strings.TrimSpace(startStr))
	endAt, ok2 := parseDBTime(strings.TrimSpace(endStr))
	if !ok1 || !ok2 || endAt.Before(startAt) {
		return 0
	}
	sec := endAt.Sub(startAt).Seconds()
	if sec < 0 {
		return 0
	}
	return int(math.Round(sec / 60))
}

// loadEventNameAndUnitByID 事件表 id -> 名称、单位（成长建议 history.type 用名称；amount 需带单位）。
func loadEventNameAndUnitByID(ctx context.Context) (names map[int64]string) {
	names = make(map[int64]string)
	var events []entity.Event
	err := dao.Event.Ctx(ctx).Fields(dao.Event.Columns().Id, dao.Event.Columns().Name, dao.Event.Columns().NeedQuantity).Scan(&events)
	if err != nil || len(events) == 0 {
		return names
	}
	for _, e := range events {
		names[e.Id] = strings.TrimSpace(e.Name)
	}
	return names
}

// growthSuggestHistoryCutoff 成长建议只取最近 48 小时内的记录（滚动「两天」）。
func growthSuggestHistoryCutoff() string {
	return time.Now().Add(-48 * time.Hour).Format("2006-01-02 15:04:05")
}

func (s *VoiceService) buildGrowthSuggestHistory(ctx context.Context, deviceNo string) ([]map[string]interface{}, error) {
	cutoff := growthSuggestHistoryCutoff()
	startCol := dao.History.Columns().StartTime
	endCol := dao.History.Columns().EndTime
	// 开始或结束时间任一落在窗口内即纳入（跨天会话也能保留）
	whereOverlap := fmt.Sprintf("(%s >= ? OR %s >= ?)", startCol, endCol)

	var rows []entity.History
	err := dao.History.Ctx(ctx).
		Where(dao.History.Columns().DeviceNo, deviceNo).
		Where(whereOverlap, cutoff, cutoff).
		OrderDesc(dao.History.Columns().Id).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	eventNames := loadEventNameAndUnitByID(ctx)
	out := make([]map[string]interface{}, 0, len(rows))
	for _, h := range rows {
		typeName := eventNames[h.EventId]
		if typeName == "" {
			typeName = strings.TrimSpace(h.EventName)
		}
		if typeName == "" {
			typeName = "未知事件"
		}
		start := strings.TrimSpace(h.StartTime)
		end := strings.TrimSpace(h.EndTime)
		note := strings.TrimSpace(h.Remark)
		amt := h.EventNumber
		if amt < 0 {
			amt = 0
		}
		item := map[string]interface{}{
			"type":         typeName,
			"start_time":   start,
			"end_time":     end,
			"amount_value": amt,
			"note":         note,
		}
		if dm := suggestDurationMinutes(start, end); dm > 0 {
			item["duration_minutes"] = dm
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *VoiceService) buildRecentHistory(ctx context.Context, deviceNo string, hours int) ([]map[string]interface{}, error) {
	if hours <= 0 {
		hours = 12
	}
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour).Format("2006-01-02 15:04:05")
	startCol := dao.History.Columns().StartTime
	endCol := dao.History.Columns().EndTime
	whereOverlap := fmt.Sprintf("(%s >= ? OR %s >= ?)", startCol, endCol)
	var rows []entity.History
	err := dao.History.Ctx(ctx).
		Where(dao.History.Columns().DeviceNo, deviceNo).
		Where(whereOverlap, cutoff, cutoff).
		OrderDesc(dao.History.Columns().Id).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, h := range rows {
		out = append(out, map[string]interface{}{
			"event_name":   strings.TrimSpace(h.EventName),
			"start_time":   strings.TrimSpace(h.StartTime),
			"end_time":     strings.TrimSpace(h.EndTime),
			"event_number": h.EventNumber,
			"remark":       strings.TrimSpace(h.Remark),
		})
	}
	return out, nil
}

func buildGrowthSuggestUserContent(child growthSuggestChildInfo, history []map[string]interface{}) (string, error) {
	ctxBlock := map[string]interface{}{
		"child_info": child,
		"history":    history,
	}
	ctxBytes, err := json.Marshal(ctxBlock)
	if err != nil {
		return "", err
	}
	// 标准 Chat Completions 只消费 messages；顶层 child_info/history 会被忽略，必须把数据写进 user.content。
	return growthSuggestUserPrompt + "\n\n" + growthSuggestDataHint + "\n\n" + string(ctxBytes), nil
}

func (s *VoiceService) buildGrowthSuggestPayload(ctx context.Context, deviceNo string) (growthSuggestPayload, error) {
	birthday, gender := s.loadDeviceProfile(ctx, deviceNo)
	history, err := s.buildGrowthSuggestHistory(ctx, deviceNo)
	if err != nil {
		return growthSuggestPayload{}, err
	}
	child := growthSuggestChildInfo{
		Birthday: birthdayForSuggestAPI(birthday),
		Gender:   gender,
	}
	userContent, err := buildGrowthSuggestUserContent(child, history)
	if err != nil {
		return growthSuggestPayload{}, err
	}
	return growthSuggestPayload{
		Messages: []growthSuggestMessage{
			{Content: "您是育儿助手，主要帮助家长根据历史记录提供成长建议。", Role: "system"},
			{Content: userContent, Role: "user"},
		},
		Model:     s.cfg.DeepSeek.Model,
		Stream:    false,
		ChildInfo: child,
		History:   history,
	}, nil
}

// callDeepSeekGrowthSuggestion 成长建议：按结构化 child_info + history 请求 DeepSeek。
func (s *VoiceService) callDeepSeekGrowthSuggestion(ctx context.Context, deviceNo string) (string, error) {
	if s.cfg.DeepSeek.Endpoint == "" {
		return "", StageError{Stage: "chat", Detail: "DeepSeek endpoint 未配置"}
	}
	payload, err := s.buildGrowthSuggestPayload(ctx, deviceNo)
	if err != nil {
		return "", err
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	if s.cfg.DebugLog {
		glog.Infof(ctx, "[大模型请求] 发送 DeepSeek 请求（成长建议）。deviceNo=%s 请求体=%s", deviceNo, string(bodyBytes))
	}
	cctx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.DeepSeek.TimeoutSeconds)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(cctx, http.MethodPost, s.cfg.DeepSeek.Endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.DeepSeek.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.DeepSeek.APIKey)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	if s.cfg.DebugLog {
		glog.Infof(ctx, "[大模型响应] 收到 DeepSeek 原始响应（成长建议）。deviceNo=%s 响应体=%s", deviceNo, string(body))
	}
	rawContent, replyNormalized, _, err := extractChatReplyRaw(body)
	if err != nil {
		return "", err
	}
	reply := pickGrowthSuggestionDisplayText(rawContent, replyNormalized)
	if s.cfg.DebugLog {
		glog.Infof(ctx, "[大模型解析] 解析回复完成（成长建议）。deviceNo=%s 回复文本=%s", deviceNo, reply)
	}

	// 将成长建议回复存入数据库，便于后续查询和分析。
	_, insertErr := dao.Suggest.Ctx(ctx).Data(g.Map{
		dao.Suggest.Columns().DeviceNo: deviceNo,
		dao.Suggest.Columns().Suggest:  reply,
		dao.Suggest.Columns().Time:     nowText(),
	}).Insert()
	if insertErr != nil {
		glog.Warningf(ctx, "insert suggest failed: %v", insertErr)
	}

	return reply, nil
}

func (s *VoiceService) callDeepSeekRaw(ctx context.Context, deviceNo, prompt string, historyLimit int, systemMessage ...string) (rawContent string, reply string, exit bool, err error) {
	if s.cfg.DeepSeek.Endpoint == "" {
		return "", "", false, StageError{Stage: "chat", Detail: "DeepSeek endpoint 未配置"}
	}
	messages := s.buildChatMessagesWithLimit(deviceNo, prompt, historyLimit, systemMessage...)
	payload := map[string]interface{}{
		"model":    s.cfg.DeepSeek.Model,
		"messages": messages,
		"stream":   false,
	}
	bodyBytes, _ := json.Marshal(payload)
	if s.cfg.DebugLog {
		glog.Infof(ctx, "[大模型请求] 发送 DeepSeek 请求（统一调用）。deviceNo=%s historyLimit=%d 请求体=%s", deviceNo, historyLimit, string(bodyBytes))
	}
	cctx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.DeepSeek.TimeoutSeconds)*time.Second)
	defer cancel()

	req, reqErr := http.NewRequestWithContext(cctx, http.MethodPost, s.cfg.DeepSeek.Endpoint, bytes.NewReader(bodyBytes))
	if reqErr != nil {
		return "", "", false, reqErr
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.DeepSeek.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.DeepSeek.APIKey)
	}
	resp, doErr := s.httpClient.Do(req)
	if doErr != nil {
		return "", "", false, doErr
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", "", false, readErr
	}
	if resp.StatusCode >= 300 {
		return "", "", false, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	if s.cfg.DebugLog {
		glog.Infof(ctx, "[大模型响应] 收到 DeepSeek 原始响应（统一调用）。deviceNo=%s 响应体=%s", deviceNo, string(body))
	}
	return extractChatReplyRaw(body)
}

func (s *VoiceService) callDeepSeekHistoryReply(ctx context.Context, deviceNo, transcript string, hours int) (string, error) {
	if hours <= 0 {
		hours = 12
	}
	history, err := s.buildRecentHistory(ctx, deviceNo, hours)
	if err != nil {
		return "", err
	}
	historyBytes, _ := json.Marshal(history)
	prompt := fmt.Sprintf("用户输入=%s。请基于最近%d小时记录回答。记录=%s。仅输出JSON：{\"reply\":\"回复内容\"}。", transcript, hours, string(historyBytes))
	systemMessage := fmt.Sprintf("您是育儿助手，主要帮助家长根据历史事件回应用户提问。")
	raw, reply, _, callErr := s.callDeepSeekRaw(ctx, deviceNo, prompt, 5, systemMessage)
	if callErr != nil {
		return "", callErr
	}
	if parsed, ok := parseGeneralChatResult(raw); ok && strings.TrimSpace(parsed.Reply) != "" {
		if s.cfg.DebugLog {
			parsedJSON, _ := json.Marshal(parsed)
			glog.Infof(ctx, "[思考过程] 历史问答结构化解析成功。deviceNo=%s parsed=%s", deviceNo, string(parsedJSON))
		}
		return parsed.Reply, nil
	}
	if s.cfg.DebugLog {
		glog.Warningf(ctx, "[思考过程] 历史问答结构化解析失败，使用文本回复。deviceNo=%s raw=%s", deviceNo, truncateVoiceLogText(raw, 800))
	}
	return strings.TrimSpace(reply), nil
}

func truncateVoiceLogText(input string, max int) string {
	s := strings.TrimSpace(input)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "...(truncated)"
}

func normalizeEventRemark(input string) string {
	remark := strings.TrimSpace(input)
	if remark == "" {
		return ""
	}
	switch strings.ToLower(remark) {
	case "无", "无备注", "不需要备注", "无须备注", "none", "null":
		return ""
	}
	runes := []rune(remark)
	if len(runes) > 10 {
		remark = string(runes[:10])
	}
	return strings.TrimSpace(remark)
}

// pickGrowthSuggestionDisplayText 成长建议：优先取模型 JSON 内的 reply 字段，避免把整段外层 JSON 写入数据库。
func pickGrowthSuggestionDisplayText(rawContent, replyNormalized string) string {
	if t := extractReplyFieldFromJSONText(rawContent); t != "" {
		return t
	}
	if t := extractReplyFieldFromJSONText(replyNormalized); t != "" {
		return t
	}
	rn := strings.TrimSpace(replyNormalized)
	rc := strings.TrimSpace(rawContent)
	if rn != "" && !strings.HasPrefix(rn, "{") {
		return rn
	}
	if rc != "" && !strings.HasPrefix(rc, "{") {
		return rc
	}
	if rn != "" {
		return rn
	}
	return rc
}

// extractReplyFieldFromJSONText 从模型输出中解析 JSON 并取出 reply（兼容代码块、assistant 包裹、首尾夹杂文本）。
func extractReplyFieldFromJSONText(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = normalizeIntentCandidateText(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(strings.TrimSpace(s), "```")
		s = strings.TrimSpace(s)
	}
	tryMap := func(jsonBytes []byte) string {
		var obj map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &obj); err != nil {
			return ""
		}
		if v, ok := obj["reply"].(string); ok {
			return strings.TrimSpace(v)
		}
		return ""
	}
	if t := tryMap([]byte(s)); t != "" {
		return t
	}
	start, end := strings.Index(s, "{"), strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		if t := tryMap([]byte(s[start : end+1])); t != "" {
			return t
		}
	}
	intent, ok := parseEventIntentFromReply(s)
	if ok {
		return strings.TrimSpace(intent.Reply)
	}
	return ""
}

// compactIntentText 归一化文本，去空白和标点，便于意图关键词匹配。
func compactIntentText(text string) string {
	trimmed := strings.TrimSpace(strings.ToLower(text))
	if trimmed == "" {
		return ""
	}
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			return -1
		}
		switch r {
		case '，', '。', '！', '？', '、', '；', '：', '（', '）', '【', '】', '《', '》', '"', '\'', '`':
			return -1
		}
		return r
	}, trimmed)
}

// isTimeoutErr 判断错误是否属于超时，便于日志快速定位“卡住直到超时”的阶段。
func isTimeoutErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "deadline exceeded") || strings.Contains(errText, "timeout")
}

func normalizeReplyAndDetectExit(reply string) (string, bool) {
	trimmed := strings.TrimSpace(reply)
	if trimmed == "" {
		return "", false
	}
	var obj map[string]interface{}
	if json.Unmarshal([]byte(trimmed), &obj) != nil {
		return trimmed, false
	}
	if hasExitMarker(obj) {
		return "", true
	}
	if v, ok := obj["reply"].(string); ok {
		return strings.TrimSpace(v), false
	}
	if v, ok := obj["content"].(string); ok {
		return strings.TrimSpace(v), false
	}
	return trimmed, false
}

func hasExitMarker(obj map[string]interface{}) bool {
	if exit, ok := obj["exit"].(bool); ok && exit {
		return true
	}
	if intent, ok := obj["intent"].(string); ok && strings.EqualFold(strings.TrimSpace(intent), "exit") {
		return true
	}
	return false
}

func extractReplyFromChoice(choice map[string]interface{}) string {
	if msg, ok := choice["message"].(map[string]interface{}); ok {
		if content, ok := msg["content"].(string); ok {
			return content
		}
	}
	if delta, ok := choice["delta"].(map[string]interface{}); ok {
		if content, ok := delta["content"].(string); ok {
			return content
		}
	}
	return ""
}

func extractChatRawContent(body []byte) (string, bool, error) {
	var obj map[string]interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		return "", false, err
	}
	if hasExitMarker(obj) {
		return "", true, nil
	}
	if v, ok := obj["reply"].(string); ok {
		return strings.TrimSpace(v), false, nil
	}
	if choices, ok := obj["choices"].([]interface{}); ok && len(choices) > 0 {
		if first, ok := choices[0].(map[string]interface{}); ok {
			if hasExitMarker(first) {
				return "", true, nil
			}
			if msg, ok := first["message"].(map[string]interface{}); ok && hasExitMarker(msg) {
				return "", true, nil
			}
			if content := strings.TrimSpace(extractReplyFromChoice(first)); content != "" {
				return content, false, nil
			}
		}
	}
	return "", false, errors.New("未找到聊天回复字段")
}

// lockDevice 基于 deviceNo 的细粒度互斥，避免同设备并发写冲突。
func (s *VoiceService) lockDevice(deviceNo string) func() {
	if strings.TrimSpace(deviceNo) == "" {
		return func() {}
	}
	v, _ := s.deviceLocks.LoadOrStore(deviceNo, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return func() {
		mu.Unlock()
	}
}

// buildChatMessages 组装模型请求消息（system + 会话历史 + 当前用户消息）。
func (s *VoiceService) buildChatMessages(deviceNo, currentPrompt string) []map[string]string {
	return s.buildChatMessagesWithLimit(deviceNo, currentPrompt, s.cfg.Session.MaxRounds*2)
}

func (s *VoiceService) buildChatMessagesWithLimit(deviceNo, currentPrompt string, historyLimit int, systemMessage ...string) []map[string]string {
	now := time.Now()
	messages := []map[string]string{}
	if s.cfg.DeepSeek.SystemPrompt != "" {
		messages = append(messages, map[string]string{"role": "system", "content": s.cfg.DeepSeek.SystemPrompt})
	}

	// 将systemMessage循环，加入到messages中，允许调用方传入多个systemMessage
	for _, msg := range systemMessage {
		if strings.TrimSpace(msg) != "" {
			messages = append(messages, map[string]string{"role": "system", "content": msg})
		}
	}
	if strings.TrimSpace(deviceNo) != "" {
		s.sessionMu.Lock()
		s.pruneSessionsLocked(now)
		if sess, ok := s.sessions[deviceNo]; ok {
			if !s.isExpired(sess.LastActive, now) {
				historyMessages := sess.Messages
				if historyLimit >= 0 && len(historyMessages) > historyLimit {
					historyMessages = historyMessages[len(historyMessages)-historyLimit:]
				}
				for _, msg := range historyMessages {
					messages = append(messages, map[string]string{"role": msg.Role, "content": msg.Content})
				}
			} else {
				delete(s.sessions, deviceNo)
				s.deviceLocks.Delete(deviceNo)
			}
		}
		s.sessionMu.Unlock()
	}

	messages = append(messages, map[string]string{"role": "user", "content": currentPrompt})
	return messages
}

// appendChatHistory 将本轮问答追加到设备会话缓存，并按 maxRounds 裁剪。
func (s *VoiceService) appendChatHistory(deviceNo, userPrompt, assistantReply string) {
	if strings.TrimSpace(deviceNo) == "" {
		return
	}
	now := time.Now()
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()

	s.pruneSessionsLocked(now)
	s.evictIfNeededLocked(deviceNo)

	sess, ok := s.sessions[deviceNo]
	if !ok || s.isExpired(sess.LastActive, now) {
		sess = &deviceChatSession{}
		s.sessions[deviceNo] = sess
	}
	sess.Messages = append(sess.Messages,
		chatHistoryMessage{Role: "user", Content: userPrompt},
		chatHistoryMessage{Role: "assistant", Content: assistantReply},
	)

	maxMessages := s.cfg.Session.MaxRounds * 2
	if maxMessages > 0 && len(sess.Messages) > maxMessages {
		sess.Messages = sess.Messages[len(sess.Messages)-maxMessages:]
	}
	sess.LastActive = now
}

func (s *VoiceService) isExpired(lastActive, now time.Time) bool {
	if s.cfg.Session.TTLSeconds <= 0 {
		return false
	}
	return now.Sub(lastActive) > time.Duration(s.cfg.Session.TTLSeconds)*time.Second
}

func (s *VoiceService) pruneSessionsLocked(now time.Time) {
	for deviceNo, sess := range s.sessions {
		if s.isExpired(sess.LastActive, now) {
			delete(s.sessions, deviceNo)
			s.deviceLocks.Delete(deviceNo)
		}
	}
}

func (s *VoiceService) evictIfNeededLocked(exemptDeviceNo string) {
	if s.cfg.Session.MaxDeviceSessions <= 0 || len(s.sessions) < s.cfg.Session.MaxDeviceSessions {
		return
	}

	victim := ""
	var oldest time.Time
	for deviceNo, sess := range s.sessions {
		if deviceNo == exemptDeviceNo {
			continue
		}
		if victim == "" || sess.LastActive.Before(oldest) {
			victim = deviceNo
			oldest = sess.LastActive
		}
	}
	if victim != "" {
		delete(s.sessions, victim)
		s.deviceLocks.Delete(victim)
	}
}

func (s *VoiceService) evictExcessSessionsLocked() {
	maxSessions := s.cfg.Session.MaxDeviceSessions
	if maxSessions <= 0 {
		return
	}
	for len(s.sessions) > maxSessions {
		victim := ""
		var oldest time.Time
		for deviceNo, sess := range s.sessions {
			if victim == "" || sess.LastActive.Before(oldest) {
				victim = deviceNo
				oldest = sess.LastActive
			}
		}
		if victim == "" {
			return
		}
		delete(s.sessions, victim)
		s.deviceLocks.Delete(victim)
	}
}

// synthesize 按配置分发到对应 TTS 实现。
func (s *VoiceService) synthesize(ctx context.Context, meta AudioMeta, reply string) ([]byte, error) {
	switch strings.ToLower(s.cfg.TTS.Provider) {
	case "baidu":
		return s.synthesizeBaidu(ctx, meta, reply)
	default:
		return s.synthesizeGeneric(ctx, meta, reply)
	}
}

// synthesizeGeneric 通用 TTS 调用。
func (s *VoiceService) synthesizeGeneric(ctx context.Context, meta AudioMeta, reply string) ([]byte, error) {
	if s.cfg.TTS.Endpoint == "" {
		return nil, StageError{Stage: "tts", Detail: "TTS endpoint 未配置"}
	}

	cctx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.TTS.TimeoutSeconds)*time.Second)
	defer cancel()

	payload := map[string]interface{}{
		"text":        reply,
		"voice":       s.cfg.TTS.Voice,
		"model":       s.cfg.TTS.Model,
		"sample_rate": meta.SampleRate,
		"format":      "wav",
	}

	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(cctx, http.MethodPost, s.cfg.TTS.Endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, StageError{Stage: "tts", Detail: err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/octet-stream")
	if s.cfg.TTS.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.TTS.APIKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, StageError{Stage: "tts", Detail: err.Error()}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, StageError{Stage: "tts", Detail: err.Error()}
	}
	if resp.StatusCode >= 300 {
		return nil, StageError{Stage: "tts", Detail: fmt.Sprintf("status %d: %s", resp.StatusCode, string(respBody))}
	}
	if len(respBody) == 0 {
		return nil, StageError{Stage: "tts", Detail: "合成音频为空"}
	}

	return respBody, nil
}

// synthesizeBaidu 百度 TTS 调用，并按上限自动分段合成后拼接。
func (s *VoiceService) synthesizeBaidu(ctx context.Context, meta AudioMeta, reply string) ([]byte, error) {
	if s.cfg.TTS.Endpoint == "" {
		return nil, StageError{Stage: "tts", Detail: "Baidu TTS endpoint 未配置"}
	}
	token, err := s.getBaiduAccessToken(ctx, &s.ttsToken, s.cfg.TTS.APIKey, s.cfg.TTS.APISecret, s.cfg.TTS.TokenEndpoint, s.cfg.TTS.TimeoutSeconds)
	if err != nil {
		return nil, StageError{Stage: "tts", Detail: err.Error()}
	}

	segments := chunkBaiduText(reply)
	if len(segments) == 0 {
		return nil, StageError{Stage: "tts", Detail: "待合成文本为空"}
	}

	var combined bytes.Buffer
	for idx, segment := range segments {
		audio, chunkErr := s.invokeBaiduTTSChunk(ctx, meta, token, segment)
		if chunkErr != nil {
			if se, ok := chunkErr.(StageError); ok {
				se.Detail = fmt.Sprintf("chunk %d/%d: %s", idx+1, len(segments), se.Detail)
				return nil, se
			}
			return nil, chunkErr
		}
		combined.Write(audio)
	}

	return combined.Bytes(), nil
}

// invokeBaiduTTSChunk 调用百度 TTS 的单段合成。
func (s *VoiceService) invokeBaiduTTSChunk(ctx context.Context, meta AudioMeta, token, text string) ([]byte, error) {
	cctx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.TTS.TimeoutSeconds)*time.Second)
	defer cancel()

	if strings.TrimSpace(text) == "" {
		return nil, StageError{Stage: "tts", Detail: "chunk 文本为空"}
	}

	form := url.Values{}
	form.Set("tex", text)
	form.Set("tok", token)
	form.Set("cuid", s.cfg.TTS.CUID)
	form.Set("ctp", "1")
	form.Set("lan", s.cfg.TTS.Language)
	form.Set("per", s.cfg.TTS.Voice)
	form.Set("aue", s.cfg.TTS.AUE)
	form.Set("spd", s.cfg.TTS.Speed)
	form.Set("pit", s.cfg.TTS.Pitch)
	form.Set("vol", s.cfg.TTS.Volume)
	sampleRate := meta.SampleRate
	if sampleRate <= 0 {
		sampleRate = s.cfg.Audio.SampleRate
	}
	if sampleRate <= 0 {
		sampleRate = 16000
	}
	form.Set("audio_ctrl", fmt.Sprintf("{\"sampling_rate\":%d}", sampleRate))

	req, err := http.NewRequestWithContext(cctx, http.MethodPost, s.cfg.TTS.Endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, StageError{Stage: "tts", Detail: err.Error()}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, StageError{Stage: "tts", Detail: err.Error()}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, StageError{Stage: "tts", Detail: err.Error()}
	}
	contentType := resp.Header.Get("Content-Type")
	if resp.StatusCode >= 300 || strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/plain") {
		var errObj struct {
			ErrNo  int    `json:"err_no"`
			ErrMsg string `json:"err_msg"`
		}
		if err := json.Unmarshal(respBody, &errObj); err == nil && errObj.ErrNo != 0 {
			return nil, StageError{Stage: "tts", Detail: fmt.Sprintf("baidu err %d: %s", errObj.ErrNo, errObj.ErrMsg)}
		}
		return nil, StageError{Stage: "tts", Detail: fmt.Sprintf("status %d: %s", resp.StatusCode, string(respBody))}
	}
	if len(respBody) == 0 {
		return nil, StageError{Stage: "tts", Detail: "合成音频为空"}
	}

	return respBody, nil
}

func chunkBaiduText(text string) []string {
	clean := sanitizeBaiduText(text)
	if clean == "" {
		return nil
	}
	if len(url.QueryEscape(clean)) <= baiduTTSMaxTextBytes {
		return []string{clean}
	}

	var chunks []string
	current := ""
	for _, r := range clean {
		candidate := current + string(r)
		if len(url.QueryEscape(candidate)) > baiduTTSMaxTextBytes {
			trimmed := strings.TrimSpace(current)
			if trimmed != "" {
				chunks = append(chunks, trimmed)
			}
			current = string(r)
			if len(url.QueryEscape(current)) > baiduTTSMaxTextBytes {
				current = ""
			}
			continue
		}
		current = candidate
	}
	trimmed := strings.TrimSpace(current)
	if trimmed != "" {
		chunks = append(chunks, trimmed)
	}
	return chunks
}

func sanitizeBaiduText(text string) string {
	if text == "" {
		return ""
	}
	replacer := strings.NewReplacer("\r", " ", "\n", " ")
	collapsed := replacer.Replace(text)
	fields := strings.Fields(collapsed)
	return strings.Join(fields, " ")
}

// decodeBase64Audio 兼容多种 base64 编码并进行解码。
func decodeBase64Audio(input string) ([]byte, error) {
	normalized, err := normalizeBase64Audio(input)
	if err != nil {
		return nil, err
	}
	decoders := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}
	var lastErr error
	for _, enc := range decoders {
		decoded, err := enc.DecodeString(normalized)
		if err == nil {
			if len(decoded) == 0 {
				return nil, errors.New("音频为空")
			}
			return decoded, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		return nil, errors.New("音频 Base64 解码失败")
	}
	return nil, fmt.Errorf("音频 Base64 解码失败: %w", lastErr)
}

func stripBase64Whitespace(data string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\ufeff', '\u200b', '\u200c', '\u200d':
			return -1
		}
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, data)
}

// normalizeBase64Audio 清洗 dataURL 前缀、空白字符并自动补齐 padding。
func normalizeBase64Audio(input string) (string, error) {
	data := strings.TrimSpace(input)
	if data == "" {
		return "", errors.New("音频为空")
	}
	if idx := strings.Index(data, ","); idx >= 0 && strings.Contains(data[:idx], ";base64") {
		data = data[idx+1:]
	}
	data = stripBase64Whitespace(data)
	if data == "" {
		return "", errors.New("音频为空")
	}
	if remainder := len(data) % 4; remainder != 0 {
		data += strings.Repeat("=", 4-remainder)
	}
	return data, nil
}

func estimateBase64DecodedLen(data string) int {
	if data == "" {
		return 0
	}
	length := len(data)
	padding := 0
	if strings.HasSuffix(data, "==") {
		padding = 2
	} else if strings.HasSuffix(data, "=") {
		padding = 1
	}
	return length/4*3 - padding
}

// stripWavDataChunkIfPresent returns PCM payload if raw is RIFF/WAVE with a data chunk; otherwise returns raw unchanged.
func stripWavDataChunkIfPresent(raw []byte) ([]byte, bool) {
	if len(raw) < 12 {
		return raw, false
	}
	if string(raw[0:4]) != "RIFF" || string(raw[8:12]) != "WAVE" {
		return raw, false
	}
	pos := 12
	for pos+8 <= len(raw) {
		id := string(raw[pos : pos+4])
		sz := int(binary.LittleEndian.Uint32(raw[pos+4 : pos+8]))
		pos += 8
		if sz < 0 || pos+sz > len(raw) {
			break
		}
		if id == "data" {
			return raw[pos : pos+sz], true
		}
		pos += sz
		if sz%2 == 1 {
			pos++
		}
	}
	return raw, false
}

func trimBaiduSTTResultPunctuation(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, "。. ")
	return strings.TrimSpace(s)
}

// Baidu may return err_no=0 but put an error phrase in result[0] instead of speech text.
func isBaiduSTTPlaceholderTranscript(s string) bool {
	t := trimBaiduSTTResultPunctuation(s)
	switch t {
	case "请求错误", "识别失败", "无识别结果", "参数错误", "网络错误", "请求超时",
		"服务繁忙", "无效音频", "音频为空", "无识别内容", "请检查参数", "鉴权失败":
		return true
	default:
		return false
	}
}

// getBaiduAccessToken 获取并缓存百度 access_token。
func (s *VoiceService) getBaiduAccessToken(ctx context.Context, cache *baiduTokenCache, apiKey, apiSecret, endpoint string, timeoutSeconds int) (string, error) {
	if apiKey == "" || apiSecret == "" {
		return "", errors.New("缺少百度 API Key/Secret")
	}
	if endpoint == "" {
		endpoint = "https://aip.baidubce.com/oauth/2.0/token"
	}
	cache.mu.Lock()
	if cache.token != "" && time.Until(cache.expiresAt) > time.Minute {
		token := cache.token
		cache.mu.Unlock()
		return token, nil
	}
	cache.mu.Unlock()

	timeout := time.Duration(timeoutSeconds)
	if timeout == 0 {
		timeout = 10
	}
	cctx, cancel := context.WithTimeout(ctx, timeout*time.Second)
	defer cancel()

	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Set("client_id", apiKey)
	values.Set("client_secret", apiSecret)
	tokenURL := endpoint + "?" + values.Encode()

	req, err := http.NewRequestWithContext(cctx, http.MethodGet, tokenURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("token status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}
	if tokenResp.Error != "" {
		return "", fmt.Errorf("baidu token error %s: %s", tokenResp.Error, tokenResp.ErrorDesc)
	}
	if tokenResp.AccessToken == "" {
		return "", errors.New("未获取到百度 access_token")
	}
	expires := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	if tokenResp.ExpiresIn == 0 {
		expires = time.Now().Add(30 * time.Minute)
	}

	cache.mu.Lock()
	cache.token = tokenResp.AccessToken
	cache.expiresAt = expires.Add(-1 * time.Minute)
	cache.mu.Unlock()

	return tokenResp.AccessToken, nil
}

func (s *VoiceService) acquireLimiter(ch chan struct{}) func() {
	if ch == nil {
		return func() {}
	}
	ch <- struct{}{}
	return func() {
		<-ch
	}
}

// extractText 从通用 STT 返回中提取转写文本字段。
func extractText(body []byte) (string, error) {
	var obj map[string]interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		return "", err
	}

	if v, ok := obj["text"].(string); ok {
		return v, nil
	}
	if data, ok := obj["data"].(map[string]interface{}); ok {
		if v, ok := data["text"].(string); ok {
			return v, nil
		}
	}
	return "", errors.New("未找到转写文本字段")
}

// extractChatReply 从 DeepSeek 风格返回中提取回复文本和退出标记。
func extractChatReply(body []byte) (string, bool, error) {
	_, reply, exit, err := extractChatReplyRaw(body)
	return reply, exit, err
}

// extractChatReplyRaw 返回原始 content 与标准化 reply（避免上层丢失 JSON 结构）。
func extractChatReplyRaw(body []byte) (rawContent string, reply string, exit bool, err error) {
	rawContent, exit, err = extractChatRawContent(body)
	if err != nil || exit {
		return rawContent, "", exit, err
	}
	reply, modelExit := normalizeReplyAndDetectExit(rawContent)
	return rawContent, reply, modelExit, nil
}

func (s *VoiceService) collectChatStreamReply(resp *http.Response) (string, error) {
	if resp == nil || resp.Body == nil {
		return "", errors.New("流式响应为空")
	}
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	// 默认 64KB，适当提高上限避免长 token 触发 scanner.Err
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var builder strings.Builder
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}

		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" {
			continue
		}
		if payload == "[DONE]" {
			break
		}

		chunk, err := extractChatStreamChunk(payload)
		if err != nil {
			return "", err
		}
		if chunk != "" {
			builder.WriteString(chunk)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.TrimSpace(builder.String()), nil
}

func extractChatStreamChunk(data string) (string, error) {
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		return "", err
	}

	choices, ok := obj["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", nil
	}
	first, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", nil
	}

	if delta, ok := first["delta"].(map[string]interface{}); ok {
		if content, ok := delta["content"].(string); ok {
			return content, nil
		}
	}
	if msg, ok := first["message"].(map[string]interface{}); ok {
		if content, ok := msg["content"].(string); ok {
			return content, nil
		}
	}

	return "", nil
}
