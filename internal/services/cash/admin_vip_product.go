package cash

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// AdminUpdateVipProductInput 开通功能管理更新一期 VIP 套餐。
type AdminUpdateVipProductInput struct {
	ProductCode      string
	Title            string
	PriceFen         int
	OriginalPriceFen int
	DurationDays     int
	AppleProductId   string
	Status           int
}

// AdminUpdateVipProduct 仅更新 vip_monthly_19，禁止插入其它 product_code。
//
// 业务：现价/原价单位为分；Apple 商品 ID 只写库，不读 env。
func AdminUpdateVipProduct(ctx context.Context, in AdminUpdateVipProductInput) error {
	code := strings.TrimSpace(in.ProductCode)
	if code == "" {
		code = ProductMonthly19
	}
	if code != ProductMonthly19 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "一期仅支持 productCode="+ProductMonthly19)
	}
	if in.PriceFen < 0 || in.OriginalPriceFen < 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "价格不能为负")
	}
	if in.DurationDays < 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "时长不能为负")
	}
	status := in.Status
	if status != 0 && status != 1 {
		status = 1
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = "VIP月会员"
	}
	exist, err := g.DB().Model("vip_product").Ctx(ctx).Where("product_code", ProductMonthly19).Count()
	if err != nil {
		return err
	}
	if exist == 0 {
		return gerror.NewCode(gcode.CodeNotFound, "VIP 商品尚未种子，禁止插入其它 SKU")
	}
	_, err = g.DB().Model("vip_product").Ctx(ctx).Where("product_code", ProductMonthly19).Data(g.Map{
		"title":              title,
		"price_fen":          in.PriceFen,
		"original_price_fen": in.OriginalPriceFen,
		"duration_days":      in.DurationDays,
		"apple_product_id":   strings.TrimSpace(in.AppleProductId),
		"status":             status,
		"updated_at":         time.Now().Unix(),
	}).Update()
	return err
}
