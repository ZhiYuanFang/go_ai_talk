// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Wx is the golang structure of table wx for DAO operations like Where/Data.
type Wx struct {
	g.Meta   `orm:"table:wx, do:true"`
	Id       interface{} //
	DeviceNo interface{} //
	Unionid  interface{} //
	Platform interface{} // 平台来源
	UserName interface{} //
	Password interface{} //
}
