package v1

import (
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// DeviceHistoryListReq 查询设备历史记录。
// 参数通过 query 传递：deviceNo/page/pageSize。
type DeviceHistoryListReq struct {
	g.Meta   `path:"/device/history/api/list" method:"get" tags:"device" summary:"设备历史列表（分页）"`
	DeviceNo string `json:"deviceNo" p:"deviceNo" dc:"设备号"`
	Page     int    `json:"page" p:"page" dc:"页码，从 1 开始"`
	PageSize int    `json:"pageSize" p:"pageSize" dc:"每页条数"`
}

// DeviceHistoryListRes 设备历史列表响应。
type DeviceHistoryListRes struct {
	List     []entity.History `json:"list"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

// DeviceHistoryFilterReq 历史筛选请求，支持按事件ID列表、时间范围、备注模糊、返回条数上限筛选。
// ignoreTimeRange 为 additive 可选参数：默认 false 时时间语义与现网一致；为真时强制忽略 startTime/endTime。
type DeviceHistoryFilterReq struct {
	g.Meta          `path:"/device/history/api/filter" method:"get" tags:"device" summary:"设备历史筛选"`
	DeviceNo        string `json:"deviceNo"  p:"deviceNo"  dc:"设备号"`
	EventIds        string `json:"eventIds"  p:"eventIds"  dc:"事件ID列表，逗号分隔，如 1,2,5；空串表示不过滤"`
	StartTime       int64  `json:"startTime" p:"startTime" dc:"开始时间，Unix 秒；0 表示不限制"`
	EndTime         int64  `json:"endTime"   p:"endTime"   dc:"结束时间，Unix 秒；0 表示不限制"`
	Limit           int    `json:"limit"     p:"limit"     dc:"返回条数上限，默认 100，上限 500；仅备注探针时上限 20"`
	Remark          string `json:"remark"    p:"remark"    dc:"备注模糊关键词；空串表示不按备注过滤。NULL/空备注行不会命中"`
	IgnoreTimeRange bool   `json:"ignoreTimeRange" p:"ignoreTimeRange" dc:"为真时完全忽略 startTime/endTime（即使非 0）；默认 false，行为与现网一致"`
}

// DeviceHistoryFilterRes 历史筛选结果响应。
type DeviceHistoryFilterRes struct {
	List []entity.History `json:"list"`
}

// DeviceHistorySuggestReq 查询设备建议。
type DeviceHistorySuggestReq struct {
	g.Meta   `path:"/device/history/api/suggest" method:"get" tags:"device" summary:"设备建议列表"`
	DeviceNo string `json:"deviceNo" p:"deviceNo" dc:"设备号"`
}

// DeviceHistorySuggestRes 设备建议列表响应。
type DeviceHistorySuggestRes struct {
	List []entity.Suggest `json:"list"`
}

// DeviceHistorySuggestDeleteReq 删除一条成长建议记录。
type DeviceHistorySuggestDeleteReq struct {
	g.Meta   `path:"/device/history/api/suggest/delete" method:"post" tags:"device" summary:"删除成长建议"`
	Id       int64  `json:"id" dc:"主键ID"`
	DeviceNo string `json:"deviceNo" dc:"设备号"`
}

// DeviceHistorySuggestDeleteRes 删除结果。
type DeviceHistorySuggestDeleteRes struct{}

// DeviceHistoryEventOptionsReq 查询历史事件可选项。
type DeviceHistoryEventOptionsReq struct {
	g.Meta `path:"/device/history/api/event/options" method:"get" tags:"device" summary:"历史事件可选项"`
}

// DeviceHistoryEventOptionsRes 历史事件可选项。
type DeviceHistoryEventOptionsRes struct {
	List []entity.Event `json:"list"`
}

// DeviceHistoryBirthdayGetReq 查询设备生日。
type DeviceHistoryBirthdayGetReq struct {
	g.Meta   `path:"/device/history/api/birthday" method:"get" tags:"device" summary:"获取设备生日"`
	DeviceNo string `json:"deviceNo" p:"deviceNo" dc:"设备号"`
}

// DeviceHistoryBirthdayGetRes 生日查询响应。
type DeviceHistoryBirthdayGetRes struct {
	BabyName string `json:"babyName" dc:"宝宝名字；未设置时为空串"`
	Birthday int64  `json:"birthday" dc:"生日，Unix 秒时间戳；0 表示未设置"`
	Sex      int    `json:"sex"`
}

// DeviceHistoryBirthdaySaveReq 保存设备生日（JSON body）。
type DeviceHistoryBirthdaySaveReq struct {
	g.Meta   `path:"/device/history/api/birthday/save" method:"post" tags:"device" summary:"保存设备生日"`
	DeviceNo string `json:"deviceNo" dc:"设备号"`
	BabyName string `json:"babyName" dc:"宝宝名字"`
	Birthday int64  `json:"birthday" dc:"生日，Unix 秒时间戳"`
	Sex      int    `json:"sex" dc:"性别（0女1男）"`
}

// DeviceHistoryBirthdaySaveRes 保存成功（无额外字段）。
type DeviceHistoryBirthdaySaveRes struct{}

// DeviceHistoryEventAddReq 手动新增历史事件。
type DeviceHistoryEventAddReq struct {
	g.Meta      `path:"/device/history/api/event/add" method:"post" tags:"device" summary:"新增历史事件"`
	DeviceNo    string `json:"deviceNo" dc:"设备号"`
	EventId     int64  `json:"eventId" dc:"事件ID"`
	EventName   string `json:"eventName" dc:"事件名"`
	EventUnit   string `json:"eventUnit" dc:"事件单位"`
	EventNumber int    `json:"eventNumber" dc:"数量"`
	StartTime   int64  `json:"startTime" dc:"开始时间，Unix 秒"`
	EndTime     int64  `json:"endTime" dc:"结束时间，Unix 秒"`
	Remark      string `json:"remark" dc:"备注"`
}

// DeviceHistoryEventAddRes 新增结果。
type DeviceHistoryEventAddRes struct {
	Id int64 `json:"id"`
}

// DeviceHistoryEventUpdateReq 手动修改历史事件。
type DeviceHistoryEventUpdateReq struct {
	g.Meta      `path:"/device/history/api/event/update" method:"post" tags:"device" summary:"修改历史事件"`
	Id          int64    `json:"id" dc:"主键ID"`
	DeviceNo    string   `json:"deviceNo" dc:"设备号"`
	EventId     int64    `json:"eventId" dc:"事件ID"`
	EventName   string   `json:"eventName" dc:"事件名"`
	EventUnit   string   `json:"eventUnit" dc:"事件单位"`
	EventNumber int      `json:"eventNumber" dc:"数量"`
	StartTime   int64    `json:"startTime" dc:"开始时间，Unix 秒"`
	EndTime     int64    `json:"endTime" dc:"结束时间，Unix 秒"`
	Remark      string   `json:"remark" dc:"备注"`
	PostId      *uint64  `json:"postId" dc:"关联 UCG 帖子；0 表示清除"`
	MediaType   *int     `json:"mediaType" dc:"0 无 / 1 图 / 2 视频"`
	ImageKeys   []string `json:"imageKeys" dc:"有序图片 objectKey"`
	VideoKey    *string  `json:"videoKey" dc:"视频 objectKey"`
}

// DeviceHistoryEventUpdateRes 修改结果。
type DeviceHistoryEventUpdateRes struct{}

// DeviceHistoryEventDeleteReq 手动删除历史事件。
type DeviceHistoryEventDeleteReq struct {
	g.Meta   `path:"/device/history/api/event/delete" method:"post" tags:"device" summary:"删除历史事件"`
	Id       int64  `json:"id" dc:"主键ID"`
	DeviceNo string `json:"deviceNo" dc:"设备号"`
}

// DeviceHistoryEventDeleteRes 删除结果。
type DeviceHistoryEventDeleteRes struct{}

// DeviceHistoryLatestReq 查询最近一条历史记录。
type DeviceHistoryLatestReq struct {
	g.Meta   `path:"/device/history/api/event/latest" method:"get" tags:"device" summary:"查询最近一条历史记录"`
	DeviceNo string `json:"deviceNo" p:"deviceNo" dc:"设备号"`
}

// DeviceHistoryLatestRes 最近一条历史记录响应。
type DeviceHistoryLatestRes struct {
	Item entity.History `json:"item"`
}

// DeviceHistoryEndLatestReq 条件结束指定 eventId 的最近一条未闭合历史（end_time=0），不要求该行是全局最新。
type DeviceHistoryEndLatestReq struct {
	g.Meta   `path:"/device/history/api/event/end-latest" method:"post" tags:"device" summary:"结束指定事件最近一条未闭合历史"`
	DeviceNo string `json:"deviceNo" dc:"设备号"`
	EventId  int64  `json:"eventId" dc:"事件ID"`
	EndTime  int64  `json:"endTime" dc:"结束时间，Unix 秒"`
	Remark   string `json:"remark" dc:"备注；非空时与结束时间一并写入，空串表示保持原备注"`
}

// DeviceHistoryEndLatestRes 条件结束结果。
type DeviceHistoryEndLatestRes struct {
	Updated bool `json:"updated"`
}

// DeviceHistoryEventBatchItem 批量写库的单条操作。
// op：create|update|delete|end；create 可用 action=start|end|one 区分开始/结束/单次。
type DeviceHistoryEventBatchItem struct {
	Op          string `json:"op" dc:"create|update|delete|end"`
	Action      string `json:"action" dc:"create 时的 start|end|one；end 可与 op=end 等价"`
	Id          int64  `json:"id" dc:"update/delete 的历史行主键"`
	EventId     int64  `json:"eventId" dc:"事件ID"`
	EventName   string `json:"eventName" dc:"事件名"`
	EventUnit   string `json:"eventUnit" dc:"单位"`
	EventNumber int    `json:"eventNumber" dc:"数量"`
	StartTime   int64  `json:"startTime" dc:"开始时间，Unix 秒"`
	EndTime     int64  `json:"endTime" dc:"结束时间，Unix 秒"`
	Remark      string `json:"remark" dc:"备注；可空"`
}

// DeviceHistoryEventBatchReq 意图/语音专用批量写历史；部分成功不整单回滚。
type DeviceHistoryEventBatchReq struct {
	g.Meta   `path:"/device/history/api/event/batch" method:"post" tags:"device" summary:"批量增删改历史事件"`
	DeviceNo string                        `json:"deviceNo" dc:"设备号"`
	Items    []DeviceHistoryEventBatchItem `json:"items" dc:"操作列表，按顺序执行"`
}

// DeviceHistoryEventBatchItemRes 单条批量结果。
type DeviceHistoryEventBatchItemRes struct {
	Index  int    `json:"index" dc:"items 下标"`
	Ok     bool   `json:"ok" dc:"是否成功"`
	Reason string `json:"reason" dc:"失败原因；成功为空"`
	Id     int64  `json:"id" dc:"create 新行 id，或 update/delete 的目标 id"`
}

// DeviceHistoryEventBatchRes 批量写结果。
type DeviceHistoryEventBatchRes struct {
	Results []DeviceHistoryEventBatchItemRes `json:"results"`
}

// DeviceHistoryPieceReq 区段内某类事件的历史记录（用于趋势图）。
type DeviceHistoryPieceReq struct {
	g.Meta    `path:"/device/history/api/piece" method:"get" tags:"device" summary:"历史区段查询"`
	EventId   int64  `json:"eventId"   p:"eventId"   dc:"事件 ID"`
	StartTime int64  `json:"startTime" p:"startTime" dc:"区间开始，Unix 秒"`
	EndTime   int64  `json:"endTime"   p:"endTime"   dc:"区间结束，Unix 秒"`
	DeviceNo  string `json:"deviceNo"  p:"deviceNo"  dc:"设备号"`
}

// DeviceHistoryPieceRes 区段历史列表。
type DeviceHistoryPieceRes struct {
	List []entity.History `json:"list"`
}
