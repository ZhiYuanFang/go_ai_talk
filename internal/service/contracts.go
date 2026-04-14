package service

import (
	"context"
	"fmt"
	"time"

	"hello/internal/model/entity"
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

// StreamASRSession 表示一个流式 ASR 会话。
type StreamASRSession interface {
	WriteAudio(chunk []byte) error
	Commit(ctx context.Context) (string, error)
	Close() error
}

// VoiceContract 语音对话服务契约。
type VoiceContract interface {
	HandleWithDialogue(ctx context.Context, deviceNo string, meta AudioMeta, audioBase64 string) ([]byte, AudioMeta, string, string, bool, bool, error)
	HandleWithTranscript(ctx context.Context, deviceNo string, meta AudioMeta, transcript string) ([]byte, AudioMeta, string, string, bool, bool, error)
	CreateStreamASRSession(ctx context.Context, meta AudioMeta, onPartial func(text string), onFinal func(text string)) (StreamASRSession, error)
	StreamRealtimeOptions() (time.Duration, int)
	TextChat(ctx context.Context, deviceNo, transcript string) (string, error)
}

// DeviceAdminContract 设备注册与事件管理契约。
type DeviceAdminContract interface {
	VerifyPassword(password string) bool
	Register(ctx context.Context, deviceNo string) (string, error)
	EnsureRegistered(ctx context.Context, deviceNo string) error
	UpdateLastTalk(ctx context.Context, deviceNo, ask, answer string) error
	List(ctx context.Context) ([]entity.User, error)
	AddEvent(ctx context.Context, name string, needQuantity int, extraNames string) error
	ListEvents(ctx context.Context) ([]entity.Event, error)
	UpdateEvent(ctx context.Context, id int64, name string, needQuantity int, extraNames string) error
	DeleteEvent(ctx context.Context, id int64) error
	ListQA(ctx context.Context) ([]entity.Qa, error)
	ListActionsForAdmin(ctx context.Context) ([]AdminActionItem, error)
	UpdateAction(ctx context.Context, id int64, name, targetType string) error
	DeleteAction(ctx context.Context, id int64) error
}

// DeviceHistoryContract 设备历史与建议、生日查询契约。
type DeviceHistoryContract interface {
	ListHistory(ctx context.Context, deviceNo string) ([]entity.History, error)
	ListSuggest(ctx context.Context, deviceNo string) ([]entity.Suggest, error)
	DeleteSuggest(ctx context.Context, id int64, deviceNo string) error
	ListEventOptions(ctx context.Context) ([]entity.Event, error)
	GetBirthday(ctx context.Context, deviceNo string) (string, int, error)
	SaveBirthday(ctx context.Context, deviceNo, birthday string, sex int) error
	AddHistory(ctx context.Context, item entity.History) (int64, error)
	UpdateHistory(ctx context.Context, item entity.History) error
	DeleteHistory(ctx context.Context, id int64, deviceNo string) error
}
