// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// History is the golang structure for table history.
type History struct {
	Id          int64  `json:"id"          ` //
	DeviceNo    string `json:"deviceNo"    ` // 设备号
	EventId     int64  `json:"eventId"     ` // 事件id
	EventName   string `json:"eventName"   ` // 事件名
	EventNumber int64  `json:"eventNumber" ` // 奶量等
	EventUnit   string `json:"eventUnit"   ` // 写入时反规范化的事件单位
	StartTime   int64  `json:"startTime"   ` // 开始时间戳
	EndTime     int64  `json:"endTime"     ` // 结束时间戳
	Remark      string `json:"remark"      ` // 备注
	PostId      uint64 `json:"postId"      ` //
	MediaType   int    `json:"mediaType"   ` //
	ImageKeys   string `json:"imageKeys"   ` //
	VideoKey    string `json:"videoKey"    ` //
}
