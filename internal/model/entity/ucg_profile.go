// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UcgProfile is the golang structure for table ucg_profile.
type UcgProfile struct {
	Id        uint64 `json:"id"        ` //
	WxId      uint64 `json:"wxId"      ` // device wx.id
	Nickname  string `json:"nickname"  ` //
	AvatarKey string `json:"avatarKey" ` // OSS objectKey only
	Bio       string `json:"bio"       ` //
	CreatedAt int64  `json:"createdAt" ` // unix seconds
	UpdatedAt int64  `json:"updatedAt" ` //
}
