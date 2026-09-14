package cash

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
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
	Status           int    `json:"status"` // 1=上架；Admin 可读下架行
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
	return scanVipProduct(one), nil
}

func scanVipProduct(one gdb.Record) *Product {
	return &Product{
		ProductCode:      one["product_code"].String(),
		Title:            one["title"].String(),
		PriceFen:         one["price_fen"].Int(),
		OriginalPriceFen: one["original_price_fen"].Int(),
		DurationDays:     one["duration_days"].Int(),
		AppleProductId:   strings.TrimSpace(one["apple_product_id"].String()),
		Status:           one["status"].Int(),
	}
}

// GetVipProductForAdmin 读取一期 VIP 商品行（含下架），供开通功能管理编辑。
func GetVipProductForAdmin(ctx context.Context) (*Product, error) {
	one, err := g.DB().Model("vip_product").Ctx(ctx).
		Where("product_code", ProductMonthly19).Limit(1).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotFound, "VIP 商品尚未种子，请重启 cash-service")
	}
	return scanVipProduct(one), nil
}
