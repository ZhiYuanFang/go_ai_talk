// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// User is the golang structure for table user.
type User struct {
	Id             int64  `json:"id"             ` //
	DeviceNo       string `json:"deviceNo"       ` // 唯一设备号
	Sex            int    `json:"sex"            ` // 性别（0女1男）
	LastTalkAsk    string `json:"lastTalkAsk"    ` // 最后对话的问题
	LastTalkAnswer string `json:"lastTalkAnswer" ` // 最后对话的答案
	ActiveTime     string `json:"activeTime"     ` // 激活时间UTF8
	LastTalkTime   string `json:"lastTalkTime"   ` // 最后对话时间UTF8
	Birthday       string `json:"birthday"       ` // 生日时间戳
}
