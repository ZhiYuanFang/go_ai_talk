package contracts

import (
	"context"
	"fmt"
	"time"

	"hello/internal/model/entity"
	sharedtypes "hello/internal/shared/types"
)

// AudioMeta 描述一路 PCM 音频的基本参数（HTTP/WS 边界与语音服务共用）。
type AudioMeta struct {
	SampleRate int
	Bits       int
	Channels   int
	Length     int
}

// StageError 表示语音链路某一阶段失败。
type StageError struct {
	Stage  string `json:"stage"`
	Detail string `json:"detail"`
}

func (e StageError) Error() string {
	return fmt.Sprintf("%s: %s", e.Stage, e.Detail)
}

type StreamASRSession interface {
	WriteAudio(chunk []byte) error
	Commit(ctx context.Context) (string, error)
	Close() error
}

type StreamTTSSession interface {
	WriteText(text string) error
	Finish(ctx context.Context) error
	Close() error
}

type VoiceContract interface {
	HandleWithDialogue(ctx context.Context, deviceNo string, meta AudioMeta, audioBase64 string) ([]byte, AudioMeta, string, string, bool, bool, error)
	HandleWithTranscript(ctx context.Context, deviceNo string, meta AudioMeta, transcript string) ([]byte, AudioMeta, string, string, bool, bool, error)
	HandleTranscriptChatOnly(ctx context.Context, deviceNo, transcript string) (ask string, answer string, exit bool, finishTalk bool, err error)
	HandleTranscriptForStreaming(ctx context.Context, deviceNo, transcript string) (ask string, answer string, mode string, needCasualStream bool, exit bool, finishTalk bool, err error)
	DetectChatMode(deviceNo, transcript string) string
	CreateStreamTTSSession(ctx context.Context, meta AudioMeta, onAudioChunk func(audio []byte, meta AudioMeta) error) (StreamTTSSession, error)
	StreamCasualReplyWithBaiduTTS(ctx context.Context, deviceNo string, meta AudioMeta, transcript string, onTextDelta func(text string) error, onAudioChunk func(audio []byte, meta AudioMeta, seq int) error) (string, error)
	StreamReplyWithBaiduTTS(ctx context.Context, meta AudioMeta, reply string, onAudioChunk func(audio []byte, meta AudioMeta, seq int) error) (chunks int, err error)
	CreateStreamASRSession(ctx context.Context, meta AudioMeta, onPartial func(text string), onFinal func(text string)) (StreamASRSession, error)
	StreamRealtimeOptions() (time.Duration, int)
	TextChat(ctx context.Context, deviceNo, transcript string) (string, error)
}

type DeviceAdminContract interface {
	VerifyPassword(password string) bool
	Register(ctx context.Context, deviceNo string) (int64, error)
	EnsureRegistered(ctx context.Context, deviceNo string) error
	UpdateLastTalk(ctx context.Context, deviceNo, ask, answer string) error
	List(ctx context.Context) ([]entity.User, error)
	AddEvent(ctx context.Context, name string, eventType string, extraNames, color, logoPath string, parentID int64) (int64, error)
	ListEvents(ctx context.Context) ([]entity.Event, error)
	UpdateEvent(ctx context.Context, id int64, name string, eventType string, extraNames, color, logoPath string, parentID *int64) error
	DeleteEvent(ctx context.Context, id int64) error
	ListQAPage(ctx context.Context, page, pageSize int) (QaPageResult, error)
	DeleteQA(ctx context.Context, id int64) error
	ListActionsForAdmin(ctx context.Context) ([]sharedtypes.AdminActionItem, error)
	UpdateAction(ctx context.Context, id int64, name, targetType string) error
	DeleteAction(ctx context.Context, id int64) error
	// SaveUserProfile 持久化用户画像（宝宝名字、生日 Unix 秒时间戳、性别），仅由 device 库承载。
	SaveUserProfile(ctx context.Context, deviceNo string, babyName string, birthdayUnixSec int64, sex int) error
	// InsertVoiceActionRecord 语音 DeepSeek 新增动作词典，名称冲突时返回错误。
	InsertVoiceActionRecord(ctx context.Context, name, targetType string) error
	// InsertOrGetEventByNeedle 统一意图路径创建事件并回读最新行。
	InsertOrGetEventByNeedle(ctx context.Context, needle string, eventType string) (entity.Event, error)
	// ApplyDeepSeekEventExtractPersistence DeepSeek 实体抽取管线写事件字典（新增或合并 extra_names）。
	ApplyDeepSeekEventExtractPersistence(ctx context.Context, out entity.Event) (entity.Event, string, error)
}

// HistoryPageResult 表示历史记录分页结果。
// 该结构只服务于外部历史列表接口，避免影响内部全量读取场景。
type HistoryPageResult struct {
	List     []entity.History
	Total    int
	Page     int
	PageSize int
}

// QaPageResult 问答库分页列表（qa 表权威在 voice 库）。
type QaPageResult struct {
	List     []entity.Qa
	Total    int
	Page     int
	PageSize int
}

type DeviceHistoryContract interface {
	ListHistory(ctx context.Context, deviceNo string) ([]entity.History, error)
	ListHistoryPage(ctx context.Context, deviceNo string, page int, pageSize int) (HistoryPageResult, error)
	GetLatestHistory(ctx context.Context, deviceNo string) (entity.History, error)
	// EndLatestHistoryIfMatch 若最近一条历史与 eventID 匹配则更新结束时间；remark 非空时同时覆盖备注，空串表示不修改原备注。
	EndLatestHistoryIfMatch(ctx context.Context, deviceNo string, eventID int64, endTimeUnixSec int64, remark string) (bool, error)
	ListSuggest(ctx context.Context, deviceNo string) ([]entity.Suggest, error)
	DeleteSuggest(ctx context.Context, id int64, deviceNo string) error
	ListEventOptions(ctx context.Context) ([]entity.Event, error)
	GetBirthday(ctx context.Context, deviceNo string) (babyName string, birthdayUnixSec int64, sex int, err error)
	SaveBirthday(ctx context.Context, deviceNo string, babyName string, birthdayUnixSec int64, sex int) error
	AddHistory(ctx context.Context, item entity.History) (int64, error)
	UpdateHistory(ctx context.Context, item entity.History) error
	DeleteHistory(ctx context.Context, id int64, deviceNo string) error
}
