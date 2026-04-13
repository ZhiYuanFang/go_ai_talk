// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Qa is the golang structure of table qa for DAO operations like Where/Data.
type Qa struct {
	g.Meta          `orm:"table:qa, do:true"`
	Id              interface{} //
	Question        interface{} // 问题
	IntentionId     interface{} // 意图id
	IntentionAnswer interface{} // 意图下的回答
	Replay          interface{} // 回复
	Attack          interface{} // 命中次数
}
