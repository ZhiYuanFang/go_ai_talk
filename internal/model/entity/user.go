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
	ActiveTime     int64  `json:"activeTime"     ` // 激活时间戳
	LastTalkTime   int64  `json:"lastTalkTime"   ` // 最后对话时间戳
	Birthday       int64  `json:"birthday"       ` // 生日时间戳
	BabyName       string `json:"babyName"       ` // 宝宝名字
	LastApiPath    string `json:"lastApiPath"    ` // 最近 HTTP 接口 METHOD /path
	LastApiAt      int64  `json:"lastApiAt"      ` // 最近 HTTP 接口时间 Unix 秒
}
