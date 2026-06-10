package v1

import "github.com/gogf/gf/v2/frame/g"

// 管理端使用统计读 API 由 gateway-app 本机处理（不经 device-service 反代）。

// DeviceAdminUsageListReq API 频率列表。
type DeviceAdminUsageListReq struct {
	g.Meta `path:"/device/admin/api/usage/list" method:"get" tags:"admin" summary:"App API 使用频率列表"`
	Days   int `json:"days" p:"days" d:"7" dc:"统计天数，默认 7；0 表示 TTL 内全部"`
}

// DeviceAdminUsageListItem 单条 API 统计。
type DeviceAdminUsageListItem struct {
	ApiKey  string `json:"apiKey"`
	Summary string `json:"summary"`
	Count   int64  `json:"count"`
	LastAt  int64  `json:"lastAt" dc:"最近成功调用 Unix 秒"`
}

// DeviceAdminUsageListRes API 列表响应。
type DeviceAdminUsageListRes struct {
	List []DeviceAdminUsageListItem `json:"list"`
	Days int                        `json:"days"`
}

// DeviceAdminUsageDetailReq 某 API 的 wxId 调用分布。
type DeviceAdminUsageDetailReq struct {
	g.Meta  `path:"/device/admin/api/usage/detail" method:"get" tags:"admin" summary:"App API 按接口下钻用户"`
	ApiKey  string `json:"apiKey" p:"apiKey" v:"required" dc:"METHOD /path 归一化键"`
	Days    int    `json:"days" p:"days" d:"7" dc:"统计天数，默认 7；0 表示 TTL 内全部"`
}

// DeviceAdminUsageDetailItem wxId 调用项。
type DeviceAdminUsageDetailItem struct {
	WxId   int64 `json:"wxId"`
	Count  int64 `json:"count"`
	LastAt int64 `json:"lastAt"`
}

// DeviceAdminUsageDetailRes 下钻响应。
type DeviceAdminUsageDetailRes struct {
	ApiKey  string                       `json:"apiKey"`
	Summary string                       `json:"summary"`
	List    []DeviceAdminUsageDetailItem `json:"list"`
	Days    int                          `json:"days"`
}

// DeviceAdminUsageUserReq 某 wxId 的 API 调用分布。
type DeviceAdminUsageUserReq struct {
	g.Meta `path:"/device/admin/api/usage/user" method:"get" tags:"admin" summary:"App API 按用户查看调用"`
	WxId   int64 `json:"wxId" p:"wxId" v:"required|min:1" dc:"微信用户主键"`
	Days   int   `json:"days" p:"days" d:"7" dc:"统计天数，默认 7；0 表示 TTL 内全部"`
}

// DeviceAdminUsageUserItem 用户 API 项。
type DeviceAdminUsageUserItem struct {
	ApiKey  string `json:"apiKey"`
	Summary string `json:"summary"`
	Count   int64  `json:"count"`
	LastAt  int64  `json:"lastAt"`
}

// DeviceAdminUsageUserRes 用户视图响应。
type DeviceAdminUsageUserRes struct {
	WxId int64                      `json:"wxId"`
	List []DeviceAdminUsageUserItem `json:"list"`
	Days int                        `json:"days"`
}
