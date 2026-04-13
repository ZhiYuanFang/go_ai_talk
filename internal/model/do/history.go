// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// History is the golang structure of table history for DAO operations like Where/Data.
type History struct {
	g.Meta      `orm:"table:history, do:true"`
	Id          interface{} //
	DeviceNo    interface{} // 设备号
	EventId     interface{} // 事件id
	EventName   interface{} // 事件名
	EventNumber interface{} // 奶量等
	EventUnit   interface{} // 事件数量对应的单位
	StartTime   interface{} // 开始时间戳
	EndTime     interface{} // 结束时间戳
	Remark      interface{} // 备注
}
