// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Event is the golang structure for table event.
type Event struct {
	Id         int64  `json:"id"         ` //
	Name       string `json:"name"       ` // 吃奶/睡觉/尿/屎等
	EventType  string `json:"eventType"  ` // 事件类型，number: 计数，time:计时，one:一次性
	Unit       string `json:"unit"       ` // 计数单位，如 ml、次
	ExtraNames string `json:"extraNames" ` // name的其它表达方式
	Color      string `json:"color"      ` //
	Logo       string `json:"logo"       ` //
	ParentId   int64  `json:"parentId"   ` // 父类ID
}
