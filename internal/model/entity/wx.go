// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Wx is the golang structure for table wx.
type Wx struct {
	Id       int64  `json:"id"       ` //
	DeviceNo string `json:"deviceNo" ` //
	Unionid  string `json:"unionid"  ` //
	Platform string `json:"platform" ` // 平台来源
	Account  string `json:"account"  ` // 账户
	Password string `json:"password" ` // 密码哈希（bcrypt，不可逆）
}
