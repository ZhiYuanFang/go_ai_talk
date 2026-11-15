package cash

// Product 支付拼串用的商品形状（功能单 / 小智 MCP 建支付宝 orderStr 时复用）。
//
// 业务说明：VIP 商品表已退役；本结构不再读 vip_product，仅作内存 DTO。
type Product struct {
	ProductCode      string `json:"productCode"`
	Title            string `json:"title"`
	PriceFen         int    `json:"priceFen"`         // 现价（分），建单用
	OriginalPriceFen int    `json:"originalPriceFen"` // 原价（分），0=不展示划线
	DurationDays     int    `json:"durationDays"`
	AppleProductId   string `json:"appleProductId"`
	Status           int    `json:"status"` // 1=上架
}
