package v1

import "github.com/gogf/gf/v2/frame/g"

// —— App：小智 MCP 永久开通 ——

// CashXiaozhiMcpUnlockReq GET 开通态与可售 SKU。
type CashXiaozhiMcpUnlockReq struct {
	g.Meta `path:"/cash/app/api/xiaozhi-mcp/unlock" method:"get" tags:"cash" summary:"小智 MCP 开通态与可售 SKU"`
}

// CashXiaozhiMcpProductItem 可售/展示 SKU。
type CashXiaozhiMcpProductItem struct {
	ProductCode      string `json:"productCode"`
	Title            string `json:"title"`
	PriceFen         int    `json:"priceFen"`
	OriginalPriceFen int    `json:"originalPriceFen"`
	AppleProductId   string `json:"appleProductId,omitempty"`
	Status           int    `json:"status"`
}

// CashXiaozhiMcpUnlockRes 开通态 data。
type CashXiaozhiMcpUnlockRes struct {
	Unlocked       bool                       `json:"unlocked"`
	TrialAvailable bool                       `json:"trialAvailable"`
	ExpiresAt      int64                      `json:"expiresAt,omitempty" dc:"0=永久；试用为截止 Unix 秒"`
	Product        *CashXiaozhiMcpProductItem `json:"product,omitempty"`
}

// CashXiaozhiMcpCreateOrderReq POST 建单。
type CashXiaozhiMcpCreateOrderReq struct {
	g.Meta  `path:"/cash/app/api/xiaozhi-mcp/orders" method:"post" tags:"cash" summary:"创建小智 MCP 永久开通订单"`
	Channel string `json:"channel" v:"required" dc:"alipay|apple_iap"`
}

// CashXiaozhiMcpCreateOrderRes 建单 data（字段对齐功能/VIP 建单）。
type CashXiaozhiMcpCreateOrderRes struct {
	OrderNo         string `json:"orderNo"`
	ProductCode     string `json:"productCode"`
	Channel         string `json:"channel"`
	AmountFen       int    `json:"amountFen"`
	AppleProductId  string `json:"appleProductId,omitempty"`
	AppAccountToken string `json:"appAccountToken,omitempty" dc:"Apple StoreKit UUID；购买必带"`
	AlipayOrderStr  string `json:"alipayOrderStr,omitempty"`
	PayTip          string `json:"payTip,omitempty"`
}

// —— Internal ——

// CashInternalXiaozhiMcpEntitlementReq 内部：wx 是否已开通小智 MCP。
type CashInternalXiaozhiMcpEntitlementReq struct {
	g.Meta `path:"/cash/internal/api/xiaozhi-mcp/entitlement" method:"get" tags:"cash" summary:"内部小智 MCP 开通态"`
	WxId   int64 `json:"wxId" p:"wxId" dc:"账号 wx 主键"`
}

// CashInternalXiaozhiMcpEntitlementRes 内部开通态。
type CashInternalXiaozhiMcpEntitlementRes struct {
	Unlocked bool `json:"unlocked"`
}

// CashInternalXiaozhiMcpEnsureAccessReq 内部：Add 前确保权益（可触发试用 claim）。
type CashInternalXiaozhiMcpEnsureAccessReq struct {
	g.Meta `path:"/cash/internal/api/xiaozhi-mcp/ensure-access-for-add" method:"post" tags:"cash" summary:"内部小智 MCP Add 前确保开通"`
	WxId   int64 `json:"wxId" v:"required"`
}

// CashInternalXiaozhiMcpEnsureAccessRes 空 data 表示已放行。
type CashInternalXiaozhiMcpEnsureAccessRes struct{}

// —— Admin ——

// CashAdminXiaozhiMcpProductGetReq GET 种子 SKU。
type CashAdminXiaozhiMcpProductGetReq struct {
	g.Meta `path:"/cash/admin/api/xiaozhi-mcp/product" method:"get" tags:"cash-admin" summary:"管理端小智 MCP SKU"`
}

// CashAdminXiaozhiMcpProductGetRes SKU data。
type CashAdminXiaozhiMcpProductGetRes struct {
	ProductCode      string `json:"productCode"`
	Title            string `json:"title"`
	PriceFen         int    `json:"priceFen"`
	OriginalPriceFen int    `json:"originalPriceFen"`
	AppleProductId   string `json:"appleProductId"`
	Status           int    `json:"status"`
	UpdatedAt        int64  `json:"updatedAt"`
}

// CashAdminXiaozhiMcpProductPutReq POST 更新种子 SKU。
type CashAdminXiaozhiMcpProductPutReq struct {
	g.Meta           `path:"/cash/admin/api/xiaozhi-mcp/product" method:"post" tags:"cash-admin" summary:"管理端更新小智 MCP SKU"`
	Title            string `json:"title"`
	PriceFen         int    `json:"priceFen"`
	OriginalPriceFen int    `json:"originalPriceFen"`
	AppleProductId   string `json:"appleProductId"`
	Status           int    `json:"status" dc:"1上架 0下架"`
}

// CashAdminXiaozhiMcpProductPutRes 空 data。
type CashAdminXiaozhiMcpProductPutRes struct{}

// CashAdminXiaozhiMcpGrantReq POST 手工授永久能力。
type CashAdminXiaozhiMcpGrantReq struct {
	g.Meta `path:"/cash/admin/api/xiaozhi-mcp/grants" method:"post" tags:"cash-admin" summary:"管理端手工授小智 MCP"`
	WxId   int64  `json:"wxId" v:"required"`
	Reason string `json:"reason" v:"required" dc:"授权理由"`
}

// CashAdminXiaozhiMcpGrantRes 授结果。
type CashAdminXiaozhiMcpGrantRes struct {
	OrderNo string `json:"orderNo"`
	WxId    int64  `json:"wxId"`
}

// CashAdminXiaozhiMcpRevokeReq POST 撤销最近手工授。
type CashAdminXiaozhiMcpRevokeReq struct {
	g.Meta `path:"/cash/admin/api/xiaozhi-mcp/grants/revoke" method:"post" tags:"cash-admin" summary:"管理端撤销小智 MCP 手工授"`
	WxId   int64 `json:"wxId" v:"required"`
}

// CashAdminXiaozhiMcpRevokeRes 空 data。
type CashAdminXiaozhiMcpRevokeRes struct{}

// CashAdminXiaozhiMcpEntitlementsReq GET 已开通人员快照列表。
type CashAdminXiaozhiMcpEntitlementsReq struct {
	g.Meta `path:"/cash/admin/api/xiaozhi-mcp/entitlements" method:"get" tags:"cash-admin" summary:"管理端小智 MCP 开通人员列表"`
	Limit  int `json:"limit" in:"query" d:"50"`
	Offset int `json:"offset" in:"query" d:"0"`
}

// CashAdminXiaozhiMcpEntitlementItem 开通快照行。
type CashAdminXiaozhiMcpEntitlementItem struct {
	WxId             int64  `json:"wxId"`
	Nickname         string `json:"nickname,omitempty"`
	UnlockMethod     string `json:"unlockMethod"`
	ChannelRef       string `json:"channelRef,omitempty"`
	UnlockedAt       int64  `json:"unlockedAt"`
	ExpiresAt        int64  `json:"expiresAt" dc:"0=永久；>0=试用截止 Unix 秒"`
	Active           bool   `json:"active"`
	RemainingSeconds int64  `json:"remainingSeconds,omitempty"`
	Status           int    `json:"status" dc:"1有效 0已撤销"`
	RevokedAt        int64  `json:"revokedAt,omitempty"`
	UpdatedAt        int64  `json:"updatedAt"`
}

// CashAdminXiaozhiMcpEntitlementsRes 开通人员列表 data。
type CashAdminXiaozhiMcpEntitlementsRes struct {
	Note  string                               `json:"note"`
	Total int                                  `json:"total"`
	List  []CashAdminXiaozhiMcpEntitlementItem `json:"list"`
}
