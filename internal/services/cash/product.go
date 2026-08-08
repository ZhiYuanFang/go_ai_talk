package cash

import (
	"context"
	"os"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// Product 一期 VIP 商品。
type Product struct {
	ProductCode      string `json:"productCode"`
	Title            string `json:"title"`
	PriceFen         int    `json:"priceFen"`         // 现价（分），建单用
	OriginalPriceFen int    `json:"originalPriceFen"` // 原价（分），0=不展示划线
	DurationDays     int    `json:"durationDays"`
	AppleProductId   string `json:"appleProductId"`
}

// GetActiveProduct 读取上架商品；默认 vip_monthly_19。
func GetActiveProduct(ctx context.Context, productCode string) (*Product, error) {
	productCode = strings.TrimSpace(productCode)
	if productCode == "" {
		productCode = ProductMonthly19
	}
	one, err := g.DB().Model("vip_product").Ctx(ctx).
		Where("product_code", productCode).Where("status", 1).Limit(1).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "商品不存在或已下架")
	}
	applePID := strings.TrimSpace(one["apple_product_id"].String())
	if applePID == "" {
		applePID = strings.TrimSpace(os.Getenv("CASH_APPLE_PRODUCT_ID"))
	}
	return &Product{
		ProductCode:      one["product_code"].String(),
		Title:            one["title"].String(),
		PriceFen:         one["price_fen"].Int(),
		OriginalPriceFen: one["original_price_fen"].Int(),
		DurationDays:     one["duration_days"].Int(),
		AppleProductId:   applePID,
	}, nil
}
