// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Action is the golang structure of table action for DAO operations like Where/Data.
type Action struct {
	g.Meta     `orm:"table:action, do:true"`
	Id         interface{} //
	Name       interface{} // 动作名
	TargetType interface{} // 目标类型
}
