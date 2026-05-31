// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// User is the golang structure of table user for DAO operations like Where/Data.
type User struct {
	g.Meta         `orm:"table:user, do:true"`
	Id             interface{} //
	DeviceNo       interface{} // 唯一设备号
	Sex            interface{} // 性别（0女1男）
	LastTalkAsk    interface{} // 最后对话的问题
	LastTalkAnswer interface{} // 最后对话的答案
	ActiveTime     interface{} // 激活时间戳
	LastTalkTime   interface{} // 最后对话时间戳
	Birthday       interface{} // 生日时间戳
	BabyName       interface{} // 宝宝名字
}
