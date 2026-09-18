package v1

import "github.com/gogf/gf/v2/frame/g"

// CashAdminRevenueSummaryReq GET 收益合计（须 X-Admin-Password）。
type CashAdminRevenueSummaryReq struct {
	g.Meta `path:"/cash/admin/api/revenue/summary" method:"get" tags:"cash-admin" summary:"管理端收益合计"`
	Start  int64 `json:"start" in:"query" dc:"起始 unix 秒，含；0 表示不限"`
	End    int64 `json:"end" in:"query" dc:"结束 unix 秒，不含；0 表示不限"`
}

// CashAdminRevenueFeatureItem 按功能名拆开的实付。
type CashAdminRevenueFeatureItem struct {
	Name      string `json:"name"`
	AmountFen int    `json:"amountFen"`
}

// CashAdminRevenueSummaryRes 差额 = vipFen + 各功能 amountFen + manualInFen - manualOutFen。
type CashAdminRevenueSummaryRes struct {
	VipFen       int                           `json:"vipFen"`
	Features     []CashAdminRevenueFeatureItem `json:"features"`
	ManualInFen  int                           `json:"manualInFen"`
	ManualOutFen int                           `json:"manualOutFen"`
	NetFen       int                           `json:"netFen"`
}

// CashAdminRevenueEntriesReq GET 手填流水（须 X-Admin-Password）。
type CashAdminRevenueEntriesReq struct {
	g.Meta `path:"/cash/admin/api/revenue/entries" method:"get" tags:"cash-admin" summary:"管理端手填进账出账列表"`
	Start  int64 `json:"start" in:"query" dc:"起始 unix 秒，含；0 表示不限"`
	End    int64 `json:"end" in:"query" dc:"结束 unix 秒，不含；0 表示不限"`
}

// CashAdminRevenueEntryItem 一条手填流水。
type CashAdminRevenueEntryItem struct {
	Id         int64  `json:"id"`
	Direction  string `json:"direction" dc:"in 进账 / out 出账"`
	Name       string `json:"name"`
	AmountFen  int    `json:"amountFen"`
	OccurredAt int64  `json:"occurredAt"`
}

// CashAdminRevenueEntriesRes 手填列表。
type CashAdminRevenueEntriesRes struct {
	List []CashAdminRevenueEntryItem `json:"list"`
}

// CashAdminRevenueEntryCreateReq POST 新增手填流水。
type CashAdminRevenueEntryCreateReq struct {
	g.Meta     `path:"/cash/admin/api/revenue/entries" method:"post" tags:"cash-admin" summary:"管理端新增手填进账或出账"`
	Direction  string `json:"direction" v:"required" dc:"in 或 out"`
	Name       string `json:"name" v:"required"`
	AmountFen  int    `json:"amountFen" v:"required|min:1"`
	OccurredAt int64  `json:"occurredAt" dc:"unix 秒；0 表示现在"`
}

// CashAdminRevenueEntryCreateRes 新增结果。
type CashAdminRevenueEntryCreateRes struct {
	Id int64 `json:"id"`
}

// CashAdminRevenueEntryUpdateReq POST 修改手填流水。
type CashAdminRevenueEntryUpdateReq struct {
	g.Meta     `path:"/cash/admin/api/revenue/entries/update" method:"post" tags:"cash-admin" summary:"管理端修改手填进账或出账"`
	Id         int64  `json:"id" v:"required|min:1"`
	Direction  string `json:"direction" v:"required"`
	Name       string `json:"name" v:"required"`
	AmountFen  int    `json:"amountFen" v:"required|min:1"`
	OccurredAt int64  `json:"occurredAt" v:"required|min:1"`
}

// CashAdminRevenueEntryUpdateRes 修改结果。
type CashAdminRevenueEntryUpdateRes struct {
	Id int64 `json:"id"`
}

// CashAdminRevenueEntryDeleteReq POST 删除手填流水。
type CashAdminRevenueEntryDeleteReq struct {
	g.Meta `path:"/cash/admin/api/revenue/entries/delete" method:"post" tags:"cash-admin" summary:"管理端删除手填进账或出账"`
	Id     int64 `json:"id" v:"required|min:1"`
}

// CashAdminRevenueEntryDeleteRes 删除结果。
type CashAdminRevenueEntryDeleteRes struct{}
