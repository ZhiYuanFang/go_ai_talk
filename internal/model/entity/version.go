// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Version is the golang structure for table version.
type Version struct {
	Id            int64  `json:"id"            ` //
	LatestVersion string `json:"latestVersion" ` // 线上最新版本号
	ForceUpdate   int    `json:"forceUpdate"   ` // 整包强制更新开关
	ReleaseNotes  string `json:"releaseNotes"  ` // 更新内容
	DownloadUrl   string `json:"downloadUrl"   ` // 下载链接
	MinVersion    string `json:"minVersion"    ` // 最低支持版本
	ReleaseDate   int64  `json:"releaseDate"   ` // 发布时间戳
}
