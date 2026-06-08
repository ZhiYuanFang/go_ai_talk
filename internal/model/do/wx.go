// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Wx is the golang structure of table wx for DAO operations like Where/Data.
type Wx struct {
	g.Meta     `orm:"table:wx, do:true"`
	Id         interface{} //
	DeviceNo   interface{} //
	Unionid    interface{} //
	AppleSub   interface{} // Apple JWT sub
	Platform   interface{} // 平台来源
	IpLocation interface{} // IP属地展示文案（省/市，客户端上报）
	Account    interface{} // 账户
	Password   interface{} // 密码哈希（bcrypt，不可逆）
}
