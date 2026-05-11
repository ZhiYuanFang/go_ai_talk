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
	Birthday string `json:"birthday"`
	Sex      int    `json:"sex"`
}

// DeviceHistoryBirthdaySaveReq 保存设备生日（JSON body）。
type DeviceHistoryBirthdaySaveReq struct {
	g.Meta   `path:"/device/history/api/birthday/save" method:"post" tags:"device" summary:"保存设备生日"`
	DeviceNo string `json:"deviceNo" dc:"设备号"`
	Birthday string `json:"birthday" dc:"生日"`
	Sex      int    `json:"sex" dc:"性别（0女1男）"`
}

// DeviceHistoryBirthdaySaveRes 保存成功（无额外字段）。
type DeviceHistoryBirthdaySaveRes struct{}

// DeviceHistoryChatReq 文本对话（不走 STT/TTS）。
type DeviceHistoryChatReq struct {
	g.Meta     `path:"/device/history/api/chat" method:"post" tags:"device" summary:"设备文本对话"`
	DeviceNo   string `json:"deviceNo" dc:"设备号"`
	Transcript string `json:"transcript" dc:"文本输入"`
}

// DeviceHistoryChatRes 文本对话结果。
type DeviceHistoryChatRes struct {
	Reply string `json:"reply"`
}

// DeviceHistoryEventAddReq 手动新增历史事件。
type DeviceHistoryEventAddReq struct {
	g.Meta      `path:"/device/history/api/event/add" method:"post" tags:"device" summary:"新增历史事件"`
	DeviceNo    string `json:"deviceNo" dc:"设备号"`
	EventId     int64  `json:"eventId" dc:"事件ID"`
	EventName   string `json:"eventName" dc:"事件名"`
	EventUnit   string `json:"eventUnit" dc:"事件单位"`
	EventNumber int    `json:"eventNumber" dc:"数量"`
	StartTime   string `json:"startTime" dc:"开始时间"`
	EndTime     string `json:"endTime" dc:"结束时间"`
	Remark      string `json:"remark" dc:"备注"`
}

// DeviceHistoryEventAddRes 新增结果。
type DeviceHistoryEventAddRes struct {
	Id int64 `json:"id"`
}

// DeviceHistoryEventUpdateReq 手动修改历史事件。
type DeviceHistoryEventUpdateReq struct {
	g.Meta      `path:"/device/history/api/event/update" method:"post" tags:"device" summary:"修改历史事件"`
	Id          int64  `json:"id" dc:"主键ID"`
	DeviceNo    string `json:"deviceNo" dc:"设备号"`
	EventId     int64  `json:"eventId" dc:"事件ID"`
	EventName   string `json:"eventName" dc:"事件名"`
	EventUnit   string `json:"eventUnit" dc:"事件单位"`
	EventNumber int    `json:"eventNumber" dc:"数量"`
	StartTime   string `json:"startTime" dc:"开始时间"`
	EndTime     string `json:"endTime" dc:"结束时间"`
	Remark      string `json:"remark" dc:"备注"`
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

// DeviceHistoryEndLatestReq 条件结束最近一条历史事件。
type DeviceHistoryEndLatestReq struct {
	g.Meta   `path:"/device/history/api/event/end-latest" method:"post" tags:"device" summary:"条件结束最近一条历史事件"`
	DeviceNo string `json:"deviceNo" dc:"设备号"`
	EventId  int64  `json:"eventId" dc:"事件ID"`
	EndTime  string `json:"endTime" dc:"结束时间"`
}

// DeviceHistoryEndLatestRes 条件结束结果。
type DeviceHistoryEndLatestRes struct {
	Updated bool `json:"updated"`
}

// DeviceHistoryPieceReq 区段内某类事件的历史记录（用于趋势图）。
type DeviceHistoryPieceReq struct {
	g.Meta    `path:"/device/history/api/piece" method:"get" tags:"device" summary:"历史区段查询"`
	EventId   int64  `json:"eventId"   p:"eventId"   dc:"事件 ID"`
	StartTime string `json:"startTime" p:"startTime" dc:"区间开始（与库内 start_time 可比）"`
	EndTime   string `json:"endTime"   p:"endTime"   dc:"区间结束"`
	DeviceNo  string `json:"deviceNo"  p:"deviceNo"  dc:"设备号"`
}

// DeviceHistoryPieceRes 区段历史列表。
type DeviceHistoryPieceRes struct {
	List []entity.History `json:"list"`
}
