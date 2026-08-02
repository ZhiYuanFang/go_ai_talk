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

// IntentStreamCallback 流式意图分析对外回调。
// 仅推送思考过程；意图 JSON 仅在 voice 内部累积解析，不经本回调外泄。
// 业务话术由 HandleTranscriptForIntentStream 返回值 Reply 承载，由调用方写入 SSE answer。
type IntentStreamCallback struct {
	OnThinking func(delta string) error // 收到思考过程片段时的回调
}

// IntentStreamResult 流式意图落地后的聊天结果（与 voice 包内 chatResult 语义对齐）。
// Thinking / 意图 JSON 不在此返回；UI 思考内容走 OnThinking，业务话术用 Reply。
type IntentStreamResult struct {
	Ask        string // 用户原始输入 / 规范化问句
	Reply      string // 业务话术（落库/确认/闲聊等最终回复）
	Exit       bool   // 是否退出对话
	FinishTalk bool   // 是否结束本轮对话
}

// TipStreamCallback 小贴士流式回调
type TipStreamCallback struct {
	OnThinking func(delta string) error   // 收到思考过程片段时的回调
	OnAnswer   func(delta string) error   // 收到回答内容片段时的回调
	OnDone     func(answerID string) error // 收到完成事件时的回调（包含 answer_id 用于反馈）
}

// TipStreamResponse 小贴士流式响应结果
type TipStreamResponse struct {
	Thinking  string // 完整的思考过程
	Answer    string // 完整的回答内容
	AnswerID  string // 回答 ID（用于提交反馈）
}

type VoiceContract interface {
	HandleWithDialogue(ctx context.Context, deviceNo string, meta AudioMeta, audioBase64 string) ([]byte, AudioMeta, string, string, bool, bool, error)
	HandleWithTranscript(ctx context.Context, deviceNo string, meta AudioMeta, transcript string) ([]byte, AudioMeta, string, string, bool, bool, error)
	HandleTranscriptChatOnly(ctx context.Context, deviceNo, transcript string) (ask string, answer string, exit bool, finishTalk bool, err error)
	HandleTranscriptForStreaming(ctx context.Context, deviceNo, transcript string) (ask string, answer string, exit bool, finishTalk bool, err error)
	CreateStreamTTSSession(ctx context.Context, meta AudioMeta, onAudioChunk func(audio []byte, meta AudioMeta) error) (StreamTTSSession, error)
	StreamReplyWithBaiduTTS(ctx context.Context, meta AudioMeta, reply string, onAudioChunk func(audio []byte, meta AudioMeta, seq int) error) (chunks int, err error)
	CreateStreamASRSession(ctx context.Context, meta AudioMeta, onPartial func(text string), onFinal func(text string)) (StreamASRSession, error)
	StreamRealtimeOptions() (time.Duration, int)
	TextChat(ctx context.Context, deviceNo, transcript string) (string, error)
	// HandleTranscriptForIntentStream 流式意图分析入口（intent_path=stream_land）。
	// 流式过程仅经 cb.OnThinking 推送思考话术；意图 JSON 内部解析并落地后，
	// 返回 Ask/Reply/Exit/FinishTalk。TTS 语音场景继续使用 HandleTranscriptForStreaming（非流式）。
	HandleTranscriptForIntentStream(ctx context.Context, deviceNo, transcript string, cb *IntentStreamCallback) (*IntentStreamResult, error)
	// TipStream 流式小贴士生成入口
	// 以流式方式接收思考过程与建议文案，内部调用 PythonAIClient.TipStream。
	// 月龄与当前时间由 Python 派生，调用方无需传入。
	TipStream(ctx context.Context, deviceNo string, eventID int64, eventName string, cb *TipStreamCallback) (*TipStreamResponse, error)
}

type DeviceAdminContract interface {
	VerifyPassword(password string) bool
	Register(ctx context.Context, deviceNo string) (int64, error)
	EnsureRegistered(ctx context.Context, deviceNo string) error
	UpdateLastTalk(ctx context.Context, deviceNo, ask, answer string) error
	List(ctx context.Context) ([]entity.User, error)
	// ListUsersPage 管理端设备分页（user 表，device_no 可模糊过滤）。
	ListUsersPage(ctx context.Context, page, pageSize int, q string) (UserPageResult, error)
	// ListWxPage 管理端 wx 账号分页。
	ListWxPage(ctx context.Context, page, pageSize int, q string) (WxPageResult, error)
	// TouchLastAPIAccess 记录设备最近一次对外 HTTP API（网关边缘或 internal 调用）。
	TouchLastAPIAccess(ctx context.Context, deviceNo, apiPath string, atUnixSec int64) error
	AddEvent(ctx context.Context, name string, eventType string, extraNames, color, unit, logoPath string, parentID int64) (int64, error)
	ListEvents(ctx context.Context) ([]entity.Event, error)
	UpdateEvent(ctx context.Context, id int64, name string, eventType string, extraNames, color, unit, logoPath string, parentID *int64) error
	DeleteEvent(ctx context.Context, id int64) error
	ListQAPage(ctx context.Context, page, pageSize int) (QaPageResult, error)
	DeleteQA(ctx context.Context, id int64) error
	ListFeedbackByWxID(ctx context.Context, wxID int64) ([]entity.Feedback, error)
	SubmitFeedback(ctx context.Context, wxID int64, question string) (entity.Feedback, error)
	ListFeedbackPage(ctx context.Context, page, pageSize int, unrepliedOnly bool) (FeedbackPageResult, error)
	ReplyFeedback(ctx context.Context, id int64, officialReply string) error
	ListActionsForAdmin(ctx context.Context) ([]sharedtypes.AdminActionItem, error)
	UpdateAction(ctx context.Context, id int64, name, targetType string) error
	DeleteAction(ctx context.Context, id int64) error
	// SaveUserProfile 持久化用户画像（宝宝名字、生日 Unix 秒时间戳、性别），仅由 device 库承载。
	SaveUserProfile(ctx context.Context, deviceNo string, babyName string, birthdayUnixSec int64, sex int) error
	// InsertVoiceActionRecord 语音 DeepSeek 新增动作词典，名称冲突时返回错误。
	InsertVoiceActionRecord(ctx context.Context, name, targetType string) error
	// InsertOrGetEventByNeedle 统一意图路径创建事件并回读最新行。
	InsertOrGetEventByNeedle(ctx context.Context, needle string, eventType, unit string) (entity.Event, error)
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

// FeedbackPageResult 用户反馈分页列表（feedback 表权威在 device 库）。
type FeedbackPageResult struct {
	List     []entity.Feedback
	Total    int
	Page     int
	PageSize int
}

// UserPageResult 管理端设备分页列表（user 表权威在 device 库）。
type UserPageResult struct {
	List     []entity.User
	Total    int
	Page     int
	PageSize int
}

// AdminWxListItem 管理端 wx 列表项（不含 password）。
type AdminWxListItem struct {
	Id        int64  `json:"id"`
	DeviceNo  string `json:"deviceNo"`
	Unionid   string `json:"unionid"`
	Platform  string `json:"platform"`
	Account   string `json:"account"`
	CreatedAt int64  `json:"createdAt"`
}

// WxPageResult 管理端 wx 账号分页列表。
type WxPageResult struct {
	List     []AdminWxListItem
	Total    int
	Page     int
	PageSize int
}

type DeviceHistoryContract interface {
	ListHistory(ctx context.Context, deviceNo string) ([]entity.History, error)
	ListHistoryPage(ctx context.Context, deviceNo string, page int, pageSize int) (HistoryPageResult, error)
	ListHistoryFilter(ctx context.Context, deviceNo string, eventIds []int64, startTime int64, endTime int64, limit int) ([]entity.History, error)
	ListHistoryPageV2(ctx context.Context, deviceNo string, page int, pageSize int, startTime int64, endTime int64, limit int) (HistoryPageResult, error)
	GetLatestHistory(ctx context.Context, deviceNo string) (entity.History, error)
	// EndLatestHistoryIfMatch 若该设备存在 eventID 对应且未闭合（end_time=0）的历史，则闭合其中 id 最大的一条并更新结束时间；
	// 不要求该行是全局最新一条。remark 非空时同时覆盖备注，空串表示不修改原备注。无未闭合匹配时返回 updated=false。
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
