package v1

import (
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// 运维 Hub 设备历史数据 API：网关校验 Admin JWT 并注入 X-Admin-Password，由 history-service 校验口令。

// DeviceAdminHistoryListReq 运维分页查询设备历史。
type DeviceAdminHistoryListReq struct {
	g.Meta   `path:"/device/admin/api/history/list" method:"get" tags:"admin" summary:"运维-设备历史列表"`
	DeviceNo string `json:"deviceNo" p:"deviceNo" dc:"设备号"`
	Page     int    `json:"page" p:"page" dc:"页码，从 1 开始"`
	PageSize int    `json:"pageSize" p:"pageSize" dc:"每页条数"`
}

// DeviceAdminHistoryListRes 设备历史分页响应。
type DeviceAdminHistoryListRes struct {
	List     []entity.History `json:"list"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

// DeviceAdminHistorySuggestReq 运维查询成长建议。
type DeviceAdminHistorySuggestReq struct {
	g.Meta   `path:"/device/admin/api/history/suggest" method:"get" tags:"admin" summary:"运维-成长建议列表"`
	DeviceNo string `json:"deviceNo" p:"deviceNo" dc:"设备号"`
}

// DeviceAdminHistorySuggestRes 成长建议列表。
type DeviceAdminHistorySuggestRes struct {
	List []entity.Suggest `json:"list"`
}

// DeviceAdminHistoryBirthdayGetReq 运维读取宝宝档案。
type DeviceAdminHistoryBirthdayGetReq struct {
	g.Meta   `path:"/device/admin/api/history/birthday" method:"get" tags:"admin" summary:"运维-获取设备生日档案"`
	DeviceNo string `json:"deviceNo" p:"deviceNo" dc:"设备号"`
}

// DeviceAdminHistoryBirthdayGetRes 生日档案响应。
type DeviceAdminHistoryBirthdayGetRes struct {
	BabyName string `json:"babyName"`
	Birthday int64  `json:"birthday"`
	Sex      int    `json:"sex"`
}

// DeviceAdminHistoryBirthdaySaveReq 运维保存宝宝档案。
type DeviceAdminHistoryBirthdaySaveReq struct {
	g.Meta   `path:"/device/admin/api/history/birthday/save" method:"post" tags:"admin" summary:"运维-保存设备生日档案"`
	DeviceNo string `json:"deviceNo" dc:"设备号"`
	BabyName string `json:"babyName" dc:"宝宝名字"`
	Birthday int64  `json:"birthday" dc:"生日 Unix 秒"`
	Sex      int    `json:"sex" dc:"性别 0女1男"`
}

// DeviceAdminHistoryBirthdaySaveRes 保存成功。
type DeviceAdminHistoryBirthdaySaveRes struct{}

// DeviceAdminHistoryEventAddReq 运维新增历史事件。
type DeviceAdminHistoryEventAddReq struct {
	g.Meta      `path:"/device/admin/api/history/event/add" method:"post" tags:"admin" summary:"运维-新增历史事件"`
	DeviceNo    string `json:"deviceNo" dc:"设备号"`
	EventId     int64  `json:"eventId" dc:"事件 ID"`
	EventName   string `json:"eventName" dc:"事件名"`
	EventUnit   string `json:"eventUnit" dc:"单位"`
	EventNumber int    `json:"eventNumber" dc:"数量"`
	StartTime   int64  `json:"startTime" dc:"开始 Unix 秒"`
	EndTime     int64  `json:"endTime" dc:"结束 Unix 秒"`
	Remark      string `json:"remark" dc:"备注"`
}

// DeviceAdminHistoryEventAddRes 新增结果。
type DeviceAdminHistoryEventAddRes struct {
	Id int64 `json:"id"`
}

// DeviceAdminHistoryEventUpdateReq 运维修改历史事件。
type DeviceAdminHistoryEventUpdateReq struct {
	g.Meta      `path:"/device/admin/api/history/event/update" method:"post" tags:"admin" summary:"运维-修改历史事件"`
	Id          int64    `json:"id" dc:"主键 ID"`
	DeviceNo    string   `json:"deviceNo" dc:"设备号"`
	EventId     int64    `json:"eventId" dc:"事件 ID"`
	EventName   string   `json:"eventName" dc:"事件名"`
	EventUnit   string   `json:"eventUnit" dc:"单位"`
	EventNumber int      `json:"eventNumber" dc:"数量"`
	StartTime   int64    `json:"startTime" dc:"开始 Unix 秒"`
	EndTime     int64    `json:"endTime" dc:"结束 Unix 秒"`
	Remark      string   `json:"remark" dc:"备注"`
	PostId      *uint64  `json:"postId" dc:"关联帖子；0 清除"`
	MediaType   *int     `json:"mediaType" dc:"0无/1图/2视频"`
	ImageKeys   []string `json:"imageKeys" dc:"图片 objectKey 有序列表"`
	VideoKey    *string  `json:"videoKey" dc:"视频 objectKey"`
}

// DeviceAdminHistoryEventUpdateRes 修改结果。
type DeviceAdminHistoryEventUpdateRes struct{}

// DeviceAdminHistoryEventDeleteReq 运维删除历史事件。
type DeviceAdminHistoryEventDeleteReq struct {
	g.Meta   `path:"/device/admin/api/history/event/delete" method:"post" tags:"admin" summary:"运维-删除历史事件"`
	Id       int64  `json:"id" dc:"主键 ID"`
	DeviceNo string `json:"deviceNo" dc:"设备号"`
}

// DeviceAdminHistoryEventDeleteRes 删除结果。
type DeviceAdminHistoryEventDeleteRes struct{}
