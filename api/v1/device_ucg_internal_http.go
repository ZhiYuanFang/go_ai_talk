package v1

import "github.com/gogf/gf/v2/frame/g"

// DeviceUcgWxValidateReq ucg 内部校验 wx 是否存在。
type DeviceUcgWxValidateReq struct {
	g.Meta `path:"/device/internal/api/ucg/wx/validate" method:"post" tags:"device" summary:"内部-UCG 校验 wxId"`
	WxId   int64 `json:"wxId" v:"required|min:1"`
}

type DeviceUcgWxValidateRes struct {
	WxId     int64  `json:"wxId"`
	Exists   bool   `json:"exists"`
	DeviceNo string `json:"deviceNo"`
	BabyName string `json:"babyName"`
}

// DeviceUcgWxBatchReq 批量拉取 wx 展示字段。
type DeviceUcgWxBatchReq struct {
	g.Meta `path:"/device/internal/api/ucg/wx/batch" method:"post" tags:"device" summary:"内部-UCG 批量 wx 展示字段"`
	WxIds  []int64 `json:"wxIds" v:"required"`
}

type DeviceUcgWxBatchItem struct {
	WxId        int64  `json:"wxId"`
	Exists      bool   `json:"exists"`
	DeviceNo    string `json:"deviceNo"`
	BabyName    string `json:"babyName"`
	IpLocation  string `json:"ipLocation,omitempty"`
	IsSimulated bool   `json:"isSimulated"`
}

// DeviceUcgWxIpLocationPutReq 更新 wx IP 属地（网关解析后由 ucg-service 写入）。
type DeviceUcgWxIpLocationPutReq struct {
	g.Meta     `path:"/device/internal/api/ucg/wx/{wxId}/ip-location" method:"put" tags:"device" summary:"内部-UCG 更新 wx IP 属地"`
	WxId       int64  `json:"wxId" in:"path" p:"wxId" v:"required|min:1"`
	IpLocation string `json:"ipLocation" v:"required|length:1,64"`
}

type DeviceUcgWxIpLocationPutRes struct{}

type DeviceUcgWxBatchRes struct {
	List []DeviceUcgWxBatchItem `json:"list"`
}

// DeviceUcgWxBabyNameReq 按 wx 主键取 baby_name（默认昵称 `{babyName}的家长`）。
type DeviceUcgWxBabyNameReq struct {
	g.Meta `path:"/device/internal/api/ucg/wx/{wxId}/baby-name" method:"get" tags:"device" summary:"内部-UCG 取 baby_name"`
	WxId   int64 `json:"wxId" in:"path" p:"wxId" v:"required|min:1"`
}

type DeviceUcgWxBabyNameRes struct {
	BabyName string `json:"babyName"`
}
