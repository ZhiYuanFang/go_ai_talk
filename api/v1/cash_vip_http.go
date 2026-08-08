package v1

import "github.com/gogf/gf/v2/frame/g"

// CashVipProductReq GET 一期 VIP 商品（允许未登录）。
type CashVipProductReq struct {
	g.Meta `path:"/cash/app/api/vip/product" method:"get" tags:"cash" summary:"VIP 商品现价与原价（可匿名）"`
}

// CashVipProductRes 商品 data。
type CashVipProductRes struct {
	ProductCode      string `json:"productCode"`
	Title            string `json:"title"`
	PriceFen         int    `json:"priceFen" dc:"现价（分），建单金额"`
	OriginalPriceFen int    `json:"originalPriceFen" dc:"原价（分），0 表示不展示划线"`
	DurationDays     int    `json:"durationDays"`
	AppleProductId   string `json:"appleProductId"`
}

// CashVipStatusReq GET 当前账号 VIP 状态。
type CashVipStatusReq struct {
	g.Meta `path:"/cash/app/api/vip/status" method:"get" tags:"cash" summary:"当前账号 VIP 状态"`
}

// CashVipStatusRes 状态 data。
type CashVipStatusRes struct {
	WxId     int64 `json:"wxId"`
	IsVip    bool  `json:"isVip"`
	ExpireAt int64 `json:"expireAt"`
}

// CashVipCreateOrderReq POST 创建 VIP 订单。
type CashVipCreateOrderReq struct {
	g.Meta      `path:"/cash/app/api/vip/orders" method:"post" tags:"cash" summary:"创建 VIP 订单"`
	ProductCode string `json:"productCode" dc:"默认 vip_monthly_19"`
	Channel     string `json:"channel" v:"required" dc:"alipay|apple_iap"`
}

// CashVipCreateOrderRes 建单 data。
type CashVipCreateOrderRes struct {
	OrderNo        string `json:"orderNo"`
	ProductCode    string `json:"productCode"`
	Channel        string `json:"channel"`
	AmountFen      int    `json:"amountFen"`
	AppleProductId string `json:"appleProductId,omitempty"`
	AlipayOrderStr string `json:"alipayOrderStr,omitempty"`
	PayTip         string `json:"payTip,omitempty"`
}

// CashVipAppleVerifyReq POST Apple IAP 验单。
type CashVipAppleVerifyReq struct {
	g.Meta            `path:"/cash/app/api/vip/apple/verify" method:"post" tags:"cash" summary:"Apple IAP 验单开通 VIP"`
	OrderNo           string `json:"orderNo" dc:"可选，缺省则自动建单"`
	TransactionId     string `json:"transactionId" v:"required"`
	ProductId         string `json:"productId" v:"required"`
	SignedTransaction string `json:"signedTransaction" dc:"JWS，生产必填"`
}

// CashVipAppleVerifyRes 验单结果。
type CashVipAppleVerifyRes struct{}

// CashVipAlipayNotifyReq 支付宝异步通知（表单；白名单）。
// 实际由 controller 读 form；此处仅登记 path。
type CashVipAlipayNotifyReq struct {
	g.Meta `path:"/cash/app/api/vip/alipay/notify" method:"post" tags:"cash" summary:"支付宝 VIP 支付异步通知"`
}

// CashVipAlipayNotifyRes 支付宝要求纯文本 success（controller 特殊写出）。
type CashVipAlipayNotifyRes struct{}

// CashInternalVipByWxIDReq 内部按 wxId 查 VIP。
type CashInternalVipByWxIDReq struct {
	g.Meta `path:"/cash/internal/api/vip/by-wx-id" method:"get" tags:"cash" summary:"内部按 wxId 查 VIP"`
	WxId   int64 `json:"wxId" p:"wxId" dc:"wx 表主键"`
}

// CashInternalVipByWxIDRes 内部 VIP data。
type CashInternalVipByWxIDRes struct {
	WxId     int64 `json:"wxId"`
	IsVip    bool  `json:"isVip"`
	ExpireAt int64 `json:"expireAt"`
}

// CashAdminVipEntitlementsReq GET 管理端 VIP 权益分页列表（只读；须 X-Admin-Password）。
type CashAdminVipEntitlementsReq struct {
	g.Meta   `path:"/cash/admin/api/vip/entitlements" method:"get" tags:"cash-admin" summary:"管理端 VIP 权益列表（含已过期）"`
	Page     int   `json:"page" in:"query" d:"1" dc:"页码，从 1 起"`
	PageSize int   `json:"pageSize" in:"query" d:"20" dc:"每页条数，最大 200"`
	WxId     int64 `json:"wxId" in:"query" dc:"可选，精确过滤 wx 主键"`
}

// CashAdminVipEntitlementItem 管理端权益行。
type CashAdminVipEntitlementItem struct {
	WxId              int64  `json:"wxId"`
	IsVip             bool   `json:"isVip" dc:"expire_at > now"`
	ExpireAt          int64  `json:"expireAt" dc:"到期 unix 秒"`
	RemainingSeconds  int64  `json:"remainingSeconds" dc:"expire_at-now，过期可为负"`
	LastPaidAmountFen int    `json:"lastPaidAmountFen" dc:"最近一次 paid 订单 amount_fen，无则为 0"`
	Channel           string `json:"channel" dc:"最近 paid 订单渠道；无则为空"`
	PaidAt            int64  `json:"paidAt" dc:"最近 paid 订单支付时间；无则为 0"`
}

// CashAdminVipEntitlementsRes 管理端权益分页 data。
type CashAdminVipEntitlementsRes struct {
	List     []CashAdminVipEntitlementItem `json:"list"`
	Page     int                           `json:"page"`
	PageSize int                           `json:"pageSize"`
	Total    int                           `json:"total"`
}
