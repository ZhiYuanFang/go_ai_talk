// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UcgFollow is the golang structure of table ucg_follow for DAO operations like Where/Data.
type UcgFollow struct {
	g.Meta       `orm:"table:ucg_follow, do:true"`
	Id           interface{} //
	FollowerWxId interface{} //
	FolloweeWxId interface{} //
	CreatedAt    interface{} //
}
