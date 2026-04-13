// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Qa is the golang structure for table qa.
type Qa struct {
	Id              int64  `json:"id"              ` //
	Question        string `json:"question"        ` // 问题
	IntentionId     int64  `json:"intentionId"     ` // 意图id
	IntentionAnswer string `json:"intentionAnswer" ` // 意图下的回答
	Replay          string `json:"replay"          ` // 回复
	Attack          int    `json:"attack"          ` // 命中次数
}
