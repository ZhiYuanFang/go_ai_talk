// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Intention is the golang structure of table intention for DAO operations like Where/Data.
type Intention struct {
	g.Meta     `orm:"table:intention, do:true"`
	Id         interface{} //
	Name       interface{} // 意图名
	UpperLimit interface{} // 附带历史消息上限
}
