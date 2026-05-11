package entity

// Wx 对应 ai_voice_device.wx 表：微信侧标识与设备绑定（wx_code 唯一；device_no 可空）。
type Wx struct {
	Id       int64  `json:"id"        ` //
	WxCode   string `json:"wxCode"    ` //
	DeviceNo string `json:"deviceNo"  ` //
	Platform string `json:"platform"  ` //
}
