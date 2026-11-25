package cash

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// MerchantProductDTO 商户订阅商品。
type MerchantProductDTO struct {
	ProductCode string `json:"productCode"`
	Title       string `json:"title"`
	PriceFen    int    `json:"priceFen"`
	PeriodDays  int    `json:"periodDays"`
	Status      int    `json:"status"`
}

// GetMerchantMonthlyProduct 读取月费商品（默认种子）。
func GetMerchantMonthlyProduct(ctx context.Context) (*MerchantProductDTO, error) {
	one, err := g.DB().Model("merchant_product").Ctx(ctx).
		Where("product_code", MerchantProductCodeMonthly).Limit(1).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() || one["status"].Int() != 1 {
		return nil, ErrMerchantProductMissing
	}
	return &MerchantProductDTO{
		ProductCode: one["product_code"].String(),
		Title:       one["title"].String(),
		PriceFen:    one["price_fen"].Int(),
		PeriodDays:  one["period_days"].Int(),
		Status:      one["status"].Int(),
	}, nil
}

// AdminUpdateMerchantMonthlyPrice Hub 更新月费分。
func AdminUpdateMerchantMonthlyPrice(ctx context.Context, priceFen int) error {
	if priceFen < 1 {
		return ErrMerchantPriceInvalid
	}
	now := time.Now().Unix()
	res, err := g.DB().Model("merchant_product").Ctx(ctx).
		Where("product_code", MerchantProductCodeMonthly).
		Data(g.Map{"price_fen": priceFen, "updated_at": now}).Update()
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		// 行不存在则种子
		_, err = g.DB().Exec(ctx, `
INSERT INTO merchant_product (product_code, title, price_fen, period_days, status, updated_at)
VALUES (?, '商户连续订阅（月）', ?, ?, 1, ?)
ON DUPLICATE KEY UPDATE price_fen=VALUES(price_fen), updated_at=VALUES(updated_at)`,
			MerchantProductCodeMonthly, priceFen, MerchantDefaultPeriodDays, now)
		return err
	}
	return nil
}

// MerchantEntitlementStatus 列表/详情状态。
type MerchantEntitlementStatus struct {
	MerchantID int64  `json:"merchantId"`
	ExpireAt   int64  `json:"expireAt"`
	Source     string `json:"source"`
	Entitled   bool   `json:"entitled"`
	Arrears    bool   `json:"arrears"`
}

// GetMerchantEntitlementStatus 读权益状态。
func GetMerchantEntitlementStatus(ctx context.Context, merchantID int64) (*MerchantEntitlementStatus, error) {
	ok, expireAt, err := IsMerchantEntitled(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	src := ""
	one, err := g.DB().Model("merchant_entitlement").Ctx(ctx).
		Where("merchant_id", merchantID).Limit(1).One()
	if err != nil {
		return nil, err
	}
	if !one.IsEmpty() {
		src = strings.TrimSpace(one["source"].String())
		expireAt = one["expire_at"].Int64()
		ok = expireAt > time.Now().Unix()
	}
	return &MerchantEntitlementStatus{
		MerchantID: merchantID,
		ExpireAt:   expireAt,
		Source:     src,
		Entitled:   ok,
		Arrears:    !ok,
	}, nil
}

// ListMerchantEntitlements Hub 列表（全表；一期可接受）。
func ListMerchantEntitlements(ctx context.Context) ([]MerchantEntitlementStatus, error) {
	rows, err := g.DB().Model("merchant_entitlement").Ctx(ctx).OrderDesc("updated_at").All()
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	out := make([]MerchantEntitlementStatus, 0, len(rows))
	for _, r := range rows {
		exp := r["expire_at"].Int64()
		ok := exp > now
		out = append(out, MerchantEntitlementStatus{
			MerchantID: r["merchant_id"].Int64(),
			ExpireAt:   exp,
			Source:     r["source"].String(),
			Entitled:   ok,
			Arrears:    !ok,
		})
	}
	return out, nil
}
