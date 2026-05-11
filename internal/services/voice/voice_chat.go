package voice

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
)

const baiduTTSMaxTextBytes = 1024
const (
	voiceStageSlowThreshold = 6 * time.Second
	voiceTotalSlowThreshold = 10 * time.Second
	quantityKeyword         = "多少?"
	voiceSessionRedisPrefix = "VOICE_SESSION_REDIS_PREFIX"
	voiceGuardRedisEnv      = "VOICE_GUARD_REDIS_ENABLED"
	voiceIdempotencyTTL     = "VOICE_TEXT_IDEMPOTENCY_TTL_SECONDS"
	voiceRateLimitPerMinute = "VOICE_TEXT_RATE_LIMIT_PER_MINUTE"
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
	cache                  cachekit.Cache
	sttToken               baiduTokenCache
	ttsToken               baiduTokenCache
	sttLimiter             chan struct{}
	chatLimiter            chan struct{}
	modeMu                 sync.Mutex
	deviceModes            map[string]string
	pendingQuantityMu      sync.Mutex
	pendingQuantity        map[string]pendingQuantityState
	deviceLocks            sync.Map
	ensureDeviceRegistered func(ctx context.Context, deviceNo string) error
	persistTalkRecord      func(ctx context.Context, deviceNo, ask, answer string) error
	taskProducer           *voiceTaskProducer
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

// startSessionJanitor 保留函数签名用于兼容历史调用。
func (s *VoiceService) startSessionJanitor() {
}

// stopSessionJanitor 保留函数签名用于兼容历史调用。
func (s *VoiceService) stopSessionJanitor() {
}

// NewVoiceService 创建服务实例并应用默认配置。
func NewVoiceService(cfg VoiceChatConfig) *VoiceService {
	applyConfigDefaults(&cfg)
	return &VoiceService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache:                  cachekit.WithObserver(cachekit.NewRedisCache(), cachekit.LoggingObserver{}),
		sttLimiter:             newLimiter(cfg.STT.MaxConcurrency),
		chatLimiter:            newLimiter(cfg.DeepSeek.MaxConcurrency),
		deviceModes:            make(map[string]string),
		pendingQuantity:        make(map[string]pendingQuantityState),
		ensureDeviceRegistered: func(ctx context.Context, deviceNo string) error { return nil },
		persistTalkRecord:      func(ctx context.Context, deviceNo, ask, answer string) error { return nil },
		taskProducer:           newVoiceTaskProducer(),
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
	if cfg.TTS.StreamIdleTimeoutSeconds <= 0 {
		cfg.TTS.StreamIdleTimeoutSeconds = 15
	}
	if cfg.TTS.StreamFinishTimeoutSeconds <= 0 {
		cfg.TTS.StreamFinishTimeoutSeconds = 8
	}
	if cfg.TTS.Provider == "baidu" {
		if cfg.TTS.TokenEndpoint == "" {
			cfg.TTS.TokenEndpoint = "https://aip.baidubce.com/oauth/2.0/token"
		}
		if cfg.TTS.StreamEndpoint == "" {
			cfg.TTS.StreamEndpoint = "wss://aip.baidubce.com/ws/2.0/speech/publiccloudspeech/v1/tts"
		}
	}
}

// defaultVoiceChatSharedRel 为仓库内 voiceChat 唯一数据源的相对路径（经 gfile.Search 解析）。
const defaultVoiceChatSharedRel = "manifest/config/voice-chat.shared.yaml"

// loadVoiceConfig 加载 voiceChat：优先共享 YAML（voice / history 共用），若 DeepSeek 等仍为空则回读 GF_GCFG 中的 voiceChat（便于单文件迁移或测试覆盖）。
func loadVoiceConfig(ctx context.Context) VoiceChatConfig {
	var cfg VoiceChatConfig
	rel := strings.TrimSpace(os.Getenv("GF_VOICE_CHAT_FILE"))
	if rel == "" {
		rel = defaultVoiceChatSharedRel
	}
	if path, err := gfile.Search(rel); err == nil && path != "" {
		if j, errLoad := gjson.Load(path); errLoad == nil {
			vc := j.Get("voiceChat")
			if vc != nil && !vc.IsNil() {
				_ = vc.Scan(&cfg)
			}
		} else {
			glog.Warningf(ctx, "[voiceChat] 共享配置解析失败 path=%s err=%v", path, errLoad)
		}
	} else if err != nil {
		glog.Warningf(ctx, "[voiceChat] 未找到共享配置文件 rel=%q err=%v，将仅依赖 GF_GCFG 中的 voiceChat", rel, err)
	}
	if strings.TrimSpace(cfg.DeepSeek.Endpoint) == "" {
		value, err := g.Cfg().Get(ctx, "voiceChat")
		if err == nil && value != nil && !value.IsNil() {
			_ = value.Scan(&cfg)
		}
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
	chatRes, err := s.chatWithResult(ctx, deviceNo, transcript, true)
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
		glog.Warningf(ctx, "[语音链路] 对话阶段耗时偏长。deviceNo=%s 耗时=%s exit=%v askLen=%d replyLen=%d", deviceNo, chatCost, chatRes.Exit, utf8.RuneCountInString(strings.TrimSpace(chatRes.Ask)), utf8.RuneCountInString(strings.TrimSpace(chatRes.Reply)))
	}

	if chatRes.Exit {
		// 当检测到退出意图时，不合成音频，直接返回 exit 标志
		totalCost := time.Since(requestStart)
		if totalCost >= voiceTotalSlowThreshold {
			glog.Warningf(ctx, "[语音链路] 退出请求总耗时偏长。deviceNo=%s 总耗时=%s", deviceNo, totalCost)
		}
		return nil, meta, chatRes.Ask, "", true, false, nil
	}

	ttsStart := time.Now()
	audio, err := s.synthesize(ctx, meta, chatRes.Reply)
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

	return audio, meta, chatRes.Ask, chatRes.Reply, false, chatRes.FinishTalk, nil
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
	chatRes, err := s.chatWithResult(ctx, deviceNo, transcript, true)
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
		glog.Warningf(ctx, "[语音链路] 对话阶段耗时偏长。deviceNo=%s 耗时=%s exit=%v askLen=%d replyLen=%d", deviceNo, chatCost, chatRes.Exit, utf8.RuneCountInString(strings.TrimSpace(chatRes.Ask)), utf8.RuneCountInString(strings.TrimSpace(chatRes.Reply)))
	}

	if chatRes.Exit {
		totalCost := time.Since(requestStart)
		if totalCost >= voiceTotalSlowThreshold {
			glog.Warningf(ctx, "[语音链路] 退出请求总耗时偏长。deviceNo=%s 总耗时=%s", deviceNo, totalCost)
		}
		return nil, meta, chatRes.Ask, "", true, false, nil
	}

	ttsStart := time.Now()
	audio, err := s.synthesize(ctx, meta, chatRes.Reply)
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

	return audio, meta, chatRes.Ask, chatRes.Reply, false, chatRes.FinishTalk, nil
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
	chatRes, err := s.chatWithResult(ctx, deviceNo, transcript, true)
	return chatRes.Reply, err
}

// TextChat 文本对话：在校验设备已注册后调用 chat，与 HandleWithTranscript 中对话阶段一致（不经 STT/TTS）。
func (s *VoiceService) TextChat(ctx context.Context, deviceNo, transcript string) (string, error) {
	if err := s.ensureDeviceRegistered(ctx, deviceNo); err != nil {
		return "", err
	}
	if err := s.applyTextChatGuards(ctx, deviceNo, transcript); err != nil {
		return "", err
	}
	if s.taskProducer != nil {
		if err := s.taskProducer.publishTaskRequested(ctx, deviceNo, transcript, "text-chat"); err != nil {
			return "", StageError{Stage: "event_publish", Detail: fmt.Sprintf("发布事件失败: %v", err)}
		}
	}
	return s.chat(ctx, deviceNo, transcript)
}

func (s *VoiceService) DetectChatMode(deviceNo, transcript string) string {
	return s.resolveChatMode(deviceNo, transcript)
}

func (s *VoiceService) CreateStreamTTSSession(ctx context.Context, meta AudioMeta, onAudioChunk func(audio []byte, meta AudioMeta) error) (StreamTTSSession, error) {
	// 统一从配置选择流式 TTS provider，调用侧无需感知具体厂商实现。
	switch strings.ToLower(strings.TrimSpace(s.cfg.TTS.Provider)) {
	case "baidu":
		if s.cfg.TTS.StreamEnabled {
			return newBaiduStreamTTSSession(ctx, s, meta, onAudioChunk)
		}
		return nil, StageError{Stage: "tts", Detail: "未启用流式TTS（tts.streamEnabled=false）"}
	default:
		return nil, StageError{Stage: "tts", Detail: "当前 provider 不支持流式TTS"}
	}
}

func (s *VoiceService) StreamCasualReplyWithBaiduTTS(
	ctx context.Context,
	deviceNo string,
	meta AudioMeta,
	transcript string,
	onTextDelta func(text string) error,
	onAudioChunk func(audio []byte, meta AudioMeta, seq int) error,
) (string, error) {
	// 进入流式回复前先校验设备，避免下游处理后才失败。
	if err := s.ensureDeviceRegistered(ctx, deviceNo); err != nil {
		return "", err
	}
	return s.streamCasualReplyWithBaiduTTS(ctx, deviceNo, meta, transcript, onTextDelta, onAudioChunk)
}

func (s *VoiceService) HandleTranscriptChatOnly(ctx context.Context, deviceNo, transcript string) (ask string, answer string, exit bool, finishTalk bool, err error) {
	if err = s.ensureDeviceRegistered(ctx, deviceNo); err != nil {
		return "", "", false, false, err
	}
	if err = s.applyTextChatGuards(ctx, deviceNo, transcript); err != nil {
		return "", "", false, false, err
	}
	chatRes, cErr := s.chatWithResult(ctx, deviceNo, transcript, true)
	if cErr != nil {
		return chatRes.Ask, chatRes.Reply, chatRes.Exit, chatRes.FinishTalk, cErr
	}
	return chatRes.Ask, chatRes.Reply, chatRes.Exit, chatRes.FinishTalk, nil
}

func (s *VoiceService) HandleTranscriptForStreaming(ctx context.Context, deviceNo, transcript string) (ask string, answer string, mode string, needCasualStream bool, exit bool, finishTalk bool, err error) {
	if err = s.ensureDeviceRegistered(ctx, deviceNo); err != nil {
		return "", "", "", false, false, false, err
	}
	if err = s.applyTextChatGuards(ctx, deviceNo, transcript); err != nil {
		return "", "", "", false, false, false, err
	}
	chatRes, cErr := s.chatWithResult(ctx, deviceNo, transcript, false)
	if cErr != nil {
		return chatRes.Ask, chatRes.Reply, chatRes.Mode, chatRes.NeedCasualStream, chatRes.Exit, chatRes.FinishTalk, cErr
	}
	return chatRes.Ask, chatRes.Reply, chatRes.Mode, chatRes.NeedCasualStream, chatRes.Exit, chatRes.FinishTalk, nil
}

func (s *VoiceService) StreamReplyWithBaiduTTS(
	ctx context.Context,
	meta AudioMeta,
	reply string,
	onAudioChunk func(audio []byte, meta AudioMeta, seq int) error,
) (int, error) {
	if strings.ToLower(strings.TrimSpace(s.cfg.TTS.Provider)) != "baidu" {
		return 0, StageError{Stage: "tts", Detail: "当前仅支持百度TTS流式分段下发"}
	}
	if !s.cfg.TTS.StreamEnabled {
		return 0, StageError{Stage: "tts", Detail: "未启用百度流式TTS（tts.streamEnabled=false）"}
	}
	seq := 0
	session, err := s.CreateStreamTTSSession(ctx, meta, func(audio []byte, chunkMeta AudioMeta) error {
		// 统一在服务层维护分片序号，前端按 seq 可重排/校验完整性。
		seq++
		if onAudioChunk != nil {
			return onAudioChunk(audio, chunkMeta, seq)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	defer session.Close()
	if writeErr := session.WriteText(reply); writeErr != nil {
		return 0, writeErr
	}
	if finishErr := session.Finish(ctx); finishErr != nil {
		return seq, finishErr
	}
	return seq, nil
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

// getOrCreateDailySuggestion 每设备每天只生成一次建议，其余请求直接复用当日记录。
func (s *VoiceService) getOrCreateDailySuggestion(ctx context.Context, deviceNo string) (string, error) {
	loc := time.Local
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Unix()
	endOfDay := startOfDay + 86400
	var existing entity.Suggest
	err := dao.Suggest.Ctx(ctx).
		Where(dao.Suggest.Columns().DeviceNo, deviceNo).
		WhereGTE(dao.Suggest.Columns().Time, startOfDay).
		WhereLT(dao.Suggest.Columns().Time, endOfDay).
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
		dao.Suggest.Columns().Time:     nowUnixSec(),
	}).Insert()
	if insertErr != nil {
		glog.Warningf(ctx, "insert suggest failed: %v", insertErr)
	}
	return suggestion, nil
}

func (s *VoiceService) loadDeviceProfile(ctx context.Context, deviceNo string) (string, string) {
	profile, err := DeviceProfile().GetProfile(ctx, deviceNo)
	if err != nil {
		glog.Warningf(ctx, "load device profile failed: deviceNo=%s err=%v", deviceNo, err)
	}
	birthday := "未设置"
	if profile.Birthday > 0 {
		birthday = time.Unix(profile.Birthday, 0).In(time.Local).Format("2006-01-02")
	}
	sexText := "女"
	if profile.Sex > 0 {
		sexText = "男"
	}
	return birthday, sexText
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
		// 非 JSON 直接按普通文本清洗后返回。
		return sanitizeModelReplyText(trimmed), false
	}
	if hasExitMarker(obj) {
		return "", true
	}
	if v, ok := obj["reply"].(string); ok {
		return sanitizeModelReplyText(v), false
	}
	if v, ok := obj["content"].(string); ok {
		return sanitizeModelReplyText(v), false
	}
	return sanitizeModelReplyText(trimmed), false
}

// sanitizeModelReplyText 清理 emoji/符号字符，提升 TTS 播放与终端展示稳定性。
func sanitizeModelReplyText(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	cleaned := strings.Map(func(r rune) rune {
		if r >= 0x1F300 && r <= 0x1FAFF {
			return -1
		}
		if r >= 0x2600 && r <= 0x27BF {
			return -1
		}
		if unicode.Is(unicode.So, r) {
			return -1
		}
		return r
	}, trimmed)
	return strings.TrimSpace(cleaned)
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
		if sess, ok := s.getSessionByDevice(gctx.New(), deviceNo, now); ok {
			historyMessages := sess.Messages
			if historyLimit >= 0 && len(historyMessages) > historyLimit {
				historyMessages = historyMessages[len(historyMessages)-historyLimit:]
			}
			for _, msg := range historyMessages {
				messages = append(messages, map[string]string{"role": msg.Role, "content": msg.Content})
			}
		}
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
	ctx := gctx.New()
	sess, _ := s.getSessionByDevice(ctx, deviceNo, now)
	if sess == nil {
		sess = &deviceChatSession{}
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
	if err := s.setSessionByDevice(ctx, deviceNo, sess); err != nil {
		glog.Warningf(ctx, "写入redis会话失败。deviceNo=%s err=%v", deviceNo, err)
	}
}

func (s *VoiceService) isExpired(lastActive, now time.Time) bool {
	if s.cfg.Session.TTLSeconds <= 0 {
		return false
	}
	return now.Sub(lastActive) > time.Duration(s.cfg.Session.TTLSeconds)*time.Second
}

func (s *VoiceService) useRedisSession() bool {
	return true
}

func (s *VoiceService) sessionRedisKey(deviceNo string) string {
	prefix := strings.TrimSpace(os.Getenv(voiceSessionRedisPrefix))
	if prefix == "" {
		prefix = "voice:session:"
	}
	return prefix + strings.TrimSpace(deviceNo)
}

func (s *VoiceService) getSessionByDevice(ctx context.Context, deviceNo string, now time.Time) (*deviceChatSession, bool) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil, false
	}
	key := s.sessionRedisKey(deviceNo)
	// 先做 Exists 可避免无效 key 触发额外网络与反序列化开销。
	exists, err := s.cache.Exists(ctx, key)
	if err != nil || !exists {
		return nil, false
	}
	text, ok, err := s.cache.Get(ctx, key)
	if err != nil {
		return nil, false
	}
	if !ok || text == "" {
		return nil, false
	}
	var sess deviceChatSession
	if err := json.Unmarshal([]byte(text), &sess); err != nil {
		return nil, false
	}
	if s.isExpired(sess.LastActive, now) {
		// 发现过期会话即主动删除，减少脏数据长期堆积。
		_ = s.cache.Del(ctx, key)
		return nil, false
	}
	return &sess, true
}

func (s *VoiceService) setSessionByDevice(ctx context.Context, deviceNo string, sess *deviceChatSession) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" || sess == nil {
		return nil
	}
	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	key := s.sessionRedisKey(deviceNo)
	ttl := s.cfg.Session.TTLSeconds
	if ttl <= 0 {
		// 配置缺失时兜底 30 分钟，保证会话缓存可回收。
		ttl = 1800
	}
	return s.cache.SetEX(ctx, key, string(data), time.Duration(ttl)*time.Second)
}

func (s *VoiceService) getLastUserMessageFromSession(ctx context.Context, deviceNo string, now time.Time) string {
	sess, ok := s.getSessionByDevice(ctx, deviceNo, now)
	if !ok || sess == nil {
		return ""
	}
	historyMessages := sess.Messages
	if len(historyMessages) > 1 {
		return historyMessages[len(historyMessages)-1].Content
	}
	return ""
}

func (s *VoiceService) applyTextChatGuards(ctx context.Context, deviceNo, transcript string) error {
	// 守卫链路可按环境开关启用，默认不影响纯离线调试。
	if !s.useRedisGuards() {
		return nil
	}
	if err := s.checkTextRateLimit(ctx, deviceNo); err != nil {
		return err
	}
	if err := s.checkTextIdempotency(ctx, deviceNo, transcript); err != nil {
		return err
	}
	return nil
}

func (s *VoiceService) useRedisGuards() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(voiceGuardRedisEnv)))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func (s *VoiceService) checkTextRateLimit(ctx context.Context, deviceNo string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil
	}
	limit := envIntOrDefault(voiceRateLimitPerMinute, 30)
	if limit <= 0 {
		return nil
	}
	minuteBucket := time.Now().Format("200601021504")
	key := fmt.Sprintf("voice:guard:rate:%s:%s", deviceNo, minuteBucket)
	// 计数与 TTL 分离：首次命中再设置过期，便于窗口平滑滚动。
	count, err := s.cache.Incr(ctx, key)
	if err != nil {
		return StageError{Stage: "rate_limit", Detail: fmt.Sprintf("限流检查依赖异常: %v", err)}
	}
	if count == 1 {
		if err = s.cache.Expire(ctx, key, 90*time.Second); err != nil {
			return StageError{Stage: "rate_limit", Detail: fmt.Sprintf("限流过期设置失败: %v", err)}
		}
	}
	if count > int64(limit) {
		return StageError{Stage: "rate_limit", Detail: "请求频率过高，请稍后再试"}
	}
	return nil
}

func (s *VoiceService) checkTextIdempotency(ctx context.Context, deviceNo, transcript string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil
	}
	ttl := envIntOrDefault(voiceIdempotencyTTL, 3)
	if ttl <= 0 {
		return nil
	}
	sum := fnv.New32a()
	_, _ = sum.Write([]byte(strings.TrimSpace(transcript)))
	// 设备号 + 文本哈希作为幂等键，短窗口内重复请求直接拦截。
	key := fmt.Sprintf("voice:guard:idem:%s:%d", deviceNo, sum.Sum32())
	ok, err := s.cache.SetNXEX(ctx, key, "1", time.Duration(ttl)*time.Second)
	if err != nil {
		return StageError{Stage: "idempotent", Detail: fmt.Sprintf("幂等检查依赖异常: %v", err)}
	}
	if !ok {
		return StageError{Stage: "idempotent", Detail: "重复请求过于频繁，请稍后重试"}
	}
	return nil
}

func envIntOrDefault(key string, d int) int {
	v, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key)))
	if err != nil || v <= 0 {
		return d
	}
	return v
}

// synthesize 按配置分发到对应 TTS 实现。
func (s *VoiceService) synthesize(ctx context.Context, meta AudioMeta, reply string) ([]byte, error) {
	switch strings.ToLower(s.cfg.TTS.Provider) {
	case "baidu":
		return s.synthesizeBaidu(ctx, meta, reply)
	default:
		return nil, StageError{Stage: "tts", Detail: "当前 provider 不支持TTS"}
	}
}

// synthesizeBaidu 百度 TTS 调用，基于流式 TTS 聚合完整音频。
func (s *VoiceService) synthesizeBaidu(ctx context.Context, meta AudioMeta, reply string) ([]byte, error) {
	if !s.cfg.TTS.StreamEnabled {
		return nil, StageError{Stage: "tts", Detail: "未启用百度流式TTS（tts.streamEnabled=false）"}
	}
	var combined bytes.Buffer
	session, err := s.CreateStreamTTSSession(ctx, meta, func(audio []byte, meta AudioMeta) error {
		combined.Write(audio)
		return nil
	})
	if err != nil {
		return nil, err
	}
	defer session.Close()
	if err := session.WriteText(reply); err != nil {
		return nil, err
	}
	if err := session.Finish(ctx); err != nil {
		return nil, err
	}
	return combined.Bytes(), nil
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
		// 兼容 data:audio/...;base64,xxxx 这种前端常见格式。
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

// stripWavDataChunkIfPresent 从 WAV 容器中剥离 data chunk 的 PCM 负载；若不是 WAV 则原样返回。
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

// 百度在 err_no=0 的情况下，仍可能在 result[0] 返回错误提示文本而非真实转写结果。
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
		// 未配置时使用百度默认 OAuth token 地址。
		endpoint = "https://aip.baidubce.com/oauth/2.0/token"
	}
	cache.mu.Lock()
	if cache.token != "" && time.Until(cache.expiresAt) > time.Minute {
		// 命中缓存直接返回，减少 token 接口调用压力。
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
	// 提前 1 分钟失效，避免边界时刻拿到即将过期 token。
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
			// 遵循 SSE 结束标识，收到后立即结束拼接。
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
		// 流式增量优先读取 delta.content。
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
