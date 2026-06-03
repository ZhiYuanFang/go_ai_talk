// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UcgProfile is the golang structure of table ucg_profile for DAO operations like Where/Data.
type UcgProfile struct {
	g.Meta    `orm:"table:ucg_profile, do:true"`
	Id        interface{} //
	WxId      interface{} // device wx.id
	Nickname  interface{} //
	AvatarKey interface{} // OSS objectKey only
	Bio       interface{} //
	CreatedAt interface{} // unix seconds
	UpdatedAt interface{} //
}
