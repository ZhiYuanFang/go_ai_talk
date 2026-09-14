package v1

import "github.com/gogf/gf/v2/frame/g"

// —— App：客户端功能使用上报（gateway-app 本机；须登录；不计入 App API usage）——

// GatewayAppClientUsageReportReq POST 上报功能事件。
type GatewayAppClientUsageReportReq struct {
	g.Meta      `path:"/device/app/api/client-usage/report" method:"post" tags:"gateway-app" summary:"客户端功能使用上报"`
	FeatureId   string `json:"featureId" v:"required" dc:"事件名（自由字符串，最长 128；勿含 |）"`
	Description string `json:"description" v:"required" dc:"运维可读描述（最长 128；勿含 |）"`
}

// GatewayAppClientUsageReportRes 上报结果（空 data）。
type GatewayAppClientUsageReportRes struct{}

// —— Admin：客户端使用统计读接口 ——

// DeviceAdminClientUsageFeaturesReq 按功能聚合列表。
type DeviceAdminClientUsageFeaturesReq struct {
	g.Meta `path:"/device/admin/api/client-usage/features" method:"get" tags:"admin" summary:"客户端使用统计-按功能"`
	Days   int    `json:"days" p:"days" d:"7" dc:"统计天数，默认 7，最大 30"`
	SortBy string `json:"sortBy" p:"sortBy" d:"count" dc:"count|lastAt"`
}

// DeviceAdminClientUsageFeatureItem 功能行。
type DeviceAdminClientUsageFeatureItem struct {
	FeatureId   string `json:"featureId"`
	Description string `json:"description" dc:"最近一次上报的描述（可空）"`
	Count       int64  `json:"count"`
	LastAt      int64  `json:"lastAt"`
}

// DeviceAdminClientUsageFeaturesRes 功能列表。
type DeviceAdminClientUsageFeaturesRes struct {
	List   []DeviceAdminClientUsageFeatureItem `json:"list"`
	Days   int                                 `json:"days"`
	SortBy string                              `json:"sortBy"`
}

// DeviceAdminClientUsageFeatureUsersReq 某功能下用户分布。
type DeviceAdminClientUsageFeatureUsersReq struct {
	g.Meta    `path:"/device/admin/api/client-usage/feature-users" method:"get" tags:"admin" summary:"客户端使用统计-功能下钻用户"`
	FeatureId string `json:"featureId" p:"featureId" v:"required"`
	Days      int    `json:"days" p:"days" d:"7"`
	SortBy    string `json:"sortBy" p:"sortBy" d:"count"`
}

// DeviceAdminClientUsageFeatureUserItem 用户行。
type DeviceAdminClientUsageFeatureUserItem struct {
	WxId     int64  `json:"wxId"`
	Nickname string `json:"nickname"`
	Count    int64  `json:"count"`
	LastAt   int64  `json:"lastAt"`
}

// DeviceAdminClientUsageFeatureUsersRes 功能下钻。
type DeviceAdminClientUsageFeatureUsersRes struct {
	FeatureId string                                   `json:"featureId"`
	List      []DeviceAdminClientUsageFeatureUserItem `json:"list"`
	Days      int                                      `json:"days"`
	SortBy    string                                   `json:"sortBy"`
}

// DeviceAdminClientUsageUserReq 某用户功能分布 + 时间线。
type DeviceAdminClientUsageUserReq struct {
	g.Meta `path:"/device/admin/api/client-usage/user" method:"get" tags:"admin" summary:"客户端使用统计-按用户"`
	WxId   int64  `json:"wxId" p:"wxId" v:"required|min:1"`
	Days   int    `json:"days" p:"days" d:"7"`
	SortBy string `json:"sortBy" p:"sortBy" d:"count"`
	// TimelineLimit 时间线条数，默认 200，最大 1000。
	TimelineLimit int `json:"timelineLimit" p:"timelineLimit" d:"200"`
}

// DeviceAdminClientUsageUserFeatureItem 用户功能行。
type DeviceAdminClientUsageUserFeatureItem struct {
	FeatureId   string `json:"featureId"`
	Description string `json:"description" dc:"最近一次上报的描述（可空）"`
	Count       int64  `json:"count"`
	LastAt      int64  `json:"lastAt"`
}

// DeviceAdminClientUsageTimelineItem 时间线行。
type DeviceAdminClientUsageTimelineItem struct {
	FeatureId   string `json:"featureId"`
	Description string `json:"description" dc:"该次上报描述（旧数据可空）"`
	At          int64  `json:"at" dc:"服务端 Unix 秒"`
}

// DeviceAdminClientUsageUserRes 用户详情。
type DeviceAdminClientUsageUserRes struct {
	WxId     int64                                    `json:"wxId"`
	List     []DeviceAdminClientUsageUserFeatureItem  `json:"list"`
	Timeline []DeviceAdminClientUsageTimelineItem     `json:"timeline"`
	Days     int                                      `json:"days"`
	SortBy   string                                   `json:"sortBy"`
}

// DeviceAdminClientUsageWxListReq 复用 usage 页的 wx 列表形态。
type DeviceAdminClientUsageWxListReq struct {
	g.Meta   `path:"/device/admin/api/client-usage/wx-list" method:"get" tags:"admin" summary:"客户端使用统计-wx 列表"`
	Page     int    `json:"page" p:"page"`
	PageSize int    `json:"pageSize" p:"pageSize"`
	Q        string `json:"q" p:"q"`
}

// DeviceAdminClientUsageWxListItem wx 行。
type DeviceAdminClientUsageWxListItem struct {
	Id        int64  `json:"id"`
	DeviceNo  string `json:"deviceNo"`
	Unionid   string `json:"unionid"`
	Platform  string `json:"platform"`
	Account   string `json:"account"`
	CreatedAt int64  `json:"createdAt"`
	Nickname  string `json:"nickname"`
}

// DeviceAdminClientUsageWxListRes wx 列表。
type DeviceAdminClientUsageWxListRes struct {
	List     []DeviceAdminClientUsageWxListItem `json:"list"`
	Total    int                                `json:"total"`
	Page     int                                `json:"page"`
	PageSize int                                `json:"pageSize"`
}
