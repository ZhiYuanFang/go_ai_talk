// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UcgPushDevice is the golang structure for table ucg_push_device.
type UcgPushDevice struct {
	Id        uint64 `json:"id"        ` //
	WxId      int64  `json:"wxId"      ` // recipient wx id
	Channel   string `json:"channel"   ` // apns | hms | mipush
	Token     string `json:"token"     ` //
	DeviceKey string `json:"deviceKey" ` // stable client installation id
	UpdatedAt int64  `json:"updatedAt" ` //
}
