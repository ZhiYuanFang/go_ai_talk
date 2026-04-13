// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Event is the golang structure for table event.
type Event struct {
	Id           int64  `json:"id"           ` //
	Name         string `json:"name"         ` // 吃奶/睡觉/尿/屎等
	NeedTime     int    `json:"needTime"     ` // 是否需要计时1要
	NeedQuantity int    `json:"needQuantity" ` // 是否需要计数1要
}
