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

// —— 引导视频（双固定槽）——

// CashXiaozhiMcpGuideVideoItem 引导视频项（Admin 含未配置；App 仅已配置）。
type CashXiaozhiMcpGuideVideoItem struct {
	Slot        string `json:"slot"`
	Title       string `json:"title"`
	ObjectKey   string `json:"objectKey,omitempty"`
	VideoCdnUrl string `json:"videoCdnUrl,omitempty"`
	CoverCdnUrl string `json:"coverCdnUrl,omitempty"`
	Configured  bool   `json:"configured"`
	UpdatedAt   int64  `json:"updatedAt,omitempty"`
}

// CashAppXiaozhiMcpGuideVideosReq GET App 已配置引导视频列表。
type CashAppXiaozhiMcpGuideVideosReq struct {
	g.Meta `path:"/cash/app/api/xiaozhi-mcp/guide-videos" method:"get" tags:"cash" summary:"小智引导视频清单"`
}

// CashAppXiaozhiMcpGuideVideosRes App 列表 data。
type CashAppXiaozhiMcpGuideVideosRes struct {
	List []CashXiaozhiMcpGuideVideoItem `json:"list"`
}

// CashAdminXiaozhiMcpGuideVideosGetReq GET Admin 两槽状态。
type CashAdminXiaozhiMcpGuideVideosGetReq struct {
	g.Meta `path:"/cash/admin/api/xiaozhi-mcp/guide-videos" method:"get" tags:"cash-admin" summary:"管理端小智引导视频两槽"`
}

// CashAdminXiaozhiMcpGuideVideosGetRes Admin 列表 data。
type CashAdminXiaozhiMcpGuideVideosGetRes struct {
	List []CashXiaozhiMcpGuideVideoItem `json:"list"`
}

// CashAdminXiaozhiMcpGuideVideoPutReq PUT 更新槽标题。
type CashAdminXiaozhiMcpGuideVideoPutReq struct {
	g.Meta `path:"/cash/admin/api/xiaozhi-mcp/guide-videos/{slot}" method:"put" tags:"cash-admin" summary:"管理端更新小智引导视频标题"`
	Slot   string `json:"slot" in:"path" v:"required"`
	Title  string `json:"title" v:"required"`
}

// CashAdminXiaozhiMcpGuideVideoPutRes 空 data。
type CashAdminXiaozhiMcpGuideVideoPutRes struct{}

// CashAdminXiaozhiMcpGuideVideoUploadReq POST multipart 上传/覆盖槽视频。
type CashAdminXiaozhiMcpGuideVideoUploadReq struct {
	g.Meta `path:"/cash/admin/api/xiaozhi-mcp/guide-videos/{slot}/upload" method:"post" tags:"cash-admin" summary:"管理端上传小智引导视频"`
	Slot   string `json:"slot" in:"path" v:"required"`
}

// CashAdminXiaozhiMcpGuideVideoUploadRes 上传后槽态。
type CashAdminXiaozhiMcpGuideVideoUploadRes struct {
	Slot        string `json:"slot"`
	Title       string `json:"title"`
	ObjectKey   string `json:"objectKey"`
	VideoCdnUrl string `json:"videoCdnUrl"`
	CoverCdnUrl string `json:"coverCdnUrl,omitempty"`
	UpdatedAt   int64  `json:"updatedAt"`
}

// CashAdminXiaozhiMcpGuideVideoDeleteReq DELETE 清空槽视频。
type CashAdminXiaozhiMcpGuideVideoDeleteReq struct {
	g.Meta `path:"/cash/admin/api/xiaozhi-mcp/guide-videos/{slot}" method:"delete" tags:"cash-admin" summary:"管理端清空小智引导视频"`
	Slot   string `json:"slot" in:"path" v:"required"`
}

// CashAdminXiaozhiMcpGuideVideoDeleteRes 空 data。
type CashAdminXiaozhiMcpGuideVideoDeleteRes struct{}
