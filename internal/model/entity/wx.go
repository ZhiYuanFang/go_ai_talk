// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Wx is the golang structure for table wx.
type Wx struct {
	Id       int64  `json:"id"       ` //
	DeviceNo string `json:"deviceNo" ` //
	UnionId  string `json:"unionId"  ` // 微信开放平台 unionid（库列 union_id）
	Platform string `json:"platform" ` // 平台来源
}
