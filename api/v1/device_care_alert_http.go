package v1

import "github.com/gogf/gf/v2/frame/g"

// CareAlertReasonDTO 护理留意结构化原因（对齐 Flutter CareAlertReason）。
type CareAlertReasonDTO struct {
	Type            string   `json:"type" dc:"原因类型"`
	Score           float64  `json:"score" dc:"得分"`
	ExpectationUsed bool     `json:"expectationUsed" dc:"是否使用月龄期望"`
	AgeMonths       int      `json:"ageMonths" dc:"月龄"`
	MedianGapMs     int64    `json:"medianGapMs" dc:"中位间隔毫秒"`
	LastGapMs       int64    `json:"lastGapMs" dc:"最近间隔毫秒"`
	ExpectGapMaxMs  int64    `json:"expectGapMaxMs" dc:"期望间隔上限毫秒"`
	P75DurMs        int64    `json:"p75DurMs" dc:"时长 P75 毫秒"`
	ElapsedMs       int64    `json:"elapsedMs" dc:"已持续毫秒"`
	ExpectDurMaxMs  int64    `json:"expectDurMaxMs" dc:"期望时长上限毫秒"`
	DailyAvg        float64  `json:"dailyAvg" dc:"日均次数"`
	Recent48hCount  int      `json:"recent48hCount" dc:"近 48h 次数"`
	StillExpected   bool     `json:"stillExpected" dc:"是否仍预期会发生"`
	DetailLines     []string `json:"detailLines" dc:"补充说明行"`
}

// CareAlertItemDTO 单条护理留意建议（对齐 Flutter CareAlertEventItem）。
type CareAlertItemDTO struct {
	SuggestionId   string               `json:"suggestionId" dc:"当日作用域 UUID"`
	EventId        string               `json:"eventId" dc:"事件 ID"`
	EventName      string               `json:"eventName" dc:"事件名称"`
	SummaryLine    string               `json:"summaryLine" dc:"摘要一行"`
	FollowUpPrompt string               `json:"followUpPrompt" dc:"树洞追问原文"`
	Reasons        []CareAlertReasonDTO `json:"reasons" dc:"结构化原因列表"`
}

// DeviceCareAlertDailyReq GET 宝宝日护理留意列表。
// force=1（或 true）时删除当日 Redis 日缓存后重新生成；仍须 App 鉴权注入 wxId>0。
type DeviceCareAlertDailyReq struct {
	g.Meta   `path:"/device/api/care-alert/daily" method:"get" tags:"device" summary:"护理留意日列表（可选 force 强刷）"`
	DeviceNo string `json:"deviceNo" p:"deviceNo" v:"required" dc:"设备号（宝宝维度）"`
	Force    string `json:"force" p:"force" dc:"强刷：1/true 时清当日缓存后重生"`
}

// DeviceCareAlertDailyRes 日列表 data（经 MiddlewareHandlerResponse 包为 envelope）。
type DeviceCareAlertDailyRes struct {
	Day   string             `json:"day" dc:"Asia/Shanghai 自然日 YYYY-MM-DD"`
	Items []CareAlertItemDTO `json:"items" dc:"留意建议列表"`
}

// DeviceCareAlertDailyItemDeleteReq 从当日缓存删除单条 suggestionId。
// Query 与 Flutter ApiClient.deleteEnvelope 对齐；亦接受 JSON body。
type DeviceCareAlertDailyItemDeleteReq struct {
	g.Meta       `path:"/device/api/care-alert/daily/item" method:"delete" tags:"device" summary:"删除当日护理留意项"`
	DeviceNo     string `json:"deviceNo" p:"deviceNo" v:"required" dc:"设备号"`
	SuggestionId string `json:"suggestionId" p:"suggestionId" v:"required" dc:"建议 UUID"`
}

// DeviceCareAlertDailyItemDeleteRes 删除后的当日列表（与 GET 同形，便于多看护对齐）。
type DeviceCareAlertDailyItemDeleteRes struct {
	Day   string             `json:"day" dc:"Asia/Shanghai 自然日 YYYY-MM-DD"`
	Items []CareAlertItemDTO `json:"items" dc:"更新后的列表"`
}

// DeviceCareAlertFeedbackReq 固定意图飞轮（无 NLP）。
type DeviceCareAlertFeedbackReq struct {
	g.Meta       `path:"/device/api/care-alert/feedback" method:"post" tags:"device" summary:"护理留意飞轮反馈"`
	DeviceNo     string `json:"deviceNo" v:"required" dc:"设备号"`
	SuggestionId string `json:"suggestionId" v:"required" dc:"建议 UUID"`
	Intent       string `json:"intent" v:"required" dc:"ignore|follow_up"`
}

// DeviceCareAlertFeedbackRes 飞轮提交结果（空 data 即可；外层 code=0 表示成功）。
type DeviceCareAlertFeedbackRes struct{}
