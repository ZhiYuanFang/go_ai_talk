// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Version is the golang structure of table version for DAO operations like Where/Data.
type Version struct {
	g.Meta        `orm:"table:version, do:true"`
	Id            interface{} //
	LatestVersion interface{} // 线上最新版本号
	ForceUpdate   interface{} // 整包强制更新开关
	ReleaseNotes  interface{} // 更新内容
	DownloadUrl   interface{} // 下载链接
	MinVersion    interface{} // 最低支持版本
	ReleaseDate   interface{} // 发布时间戳
}
