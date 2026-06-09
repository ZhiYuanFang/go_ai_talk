// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Event is the golang structure of table event for DAO operations like Where/Data.
type Event struct {
	g.Meta     `orm:"table:event, do:true"`
	Id         interface{} //
	Name       interface{} // 吃奶/睡觉/尿/屎等
	EventType  interface{} // 事件类型，number: 计数，time:计时，one:一次性
	Unit       interface{} // 计数单位，如 ml、次
	ExtraNames interface{} // name的其它表达方式
	Color      interface{} //
	Logo       interface{} //
	ParentId   interface{} // 父类ID
}
