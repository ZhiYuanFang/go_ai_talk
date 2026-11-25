package cash

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// CreateMerchantAlipayOrder 建商户连续订阅单（网页支付宝）；返回 orderNo 与调起占位。
// 完整周期扣款签约字段以实现期对接开放平台为准；本函数落库 created 单并返回金额。
func CreateMerchantAlipayOrder(ctx context.Context, merchantID int64) (orderNo string, amountFen int, payPayload string, err error) {
	if merchantID <= 0 {
		return "", 0, "", ErrMerchantIDInvalid
	}
	p, err := GetMerchantMonthlyProduct(ctx)
	if err != nil {
		return "", 0, "", err
	}
	orderNo, err = newOrderNo()
	if err != nil {
		return "", 0, "", err
	}
	orderNo = "M" + orderNo
	now := time.Now().Unix()
	_, err = g.DB().Model("merchant_order").Ctx(ctx).Data(g.Map{
		"order_no":     orderNo,
		"merchant_id":  merchantID,
		"product_code": p.ProductCode,
		"channel":      ChannelAlipay,
		"amount_fen":   p.PriceFen,
		"status":       OrderCreated,
		"created_at":   now,
	}).Insert()
	if err != nil {
		return "", 0, "", err
	}
	// 占位：真实环境应返回支付宝 page pay / 周期扣款签约串。
	payPayload = fmt.Sprintf("merchant_alipay_period_placeholder:%s:%d", orderNo, p.PriceFen)
	return orderNo, p.PriceFen, payPayload, nil
}

// PeekMerchantOrderAmountYuan 返回订单金额的元字符串（模拟 notify 用）；不存在返回错误。
func PeekMerchantOrderAmountYuan(ctx context.Context, orderNo string) (string, error) {
	orderNo = strings.TrimSpace(orderNo)
	one, err := g.DB().Model("merchant_order").Ctx(ctx).Where("order_no", orderNo).Limit(1).One()
	if err != nil {
		return "", err
	}
	if one.IsEmpty() {
		return "", ErrMerchantOrderNotFound
	}
	fen := one["amount_fen"].Int()
	return fmt.Sprintf("%.2f", float64(fen)/100.0), nil
}

// FulfillMerchantAlipayPaid 幂等履约：置 paid 并按商品 period_days 延长 expire_at。
func FulfillMerchantAlipayPaid(ctx context.Context, orderNo, channelTxnID string) error {
	orderNo = strings.TrimSpace(orderNo)
	if orderNo == "" {
		return ErrMerchantOrderNotFound
	}
	one, err := g.DB().Model("merchant_order").Ctx(ctx).Where("order_no", orderNo).Limit(1).One()
	if err != nil {
		return err
	}
	if one.IsEmpty() {
		return ErrMerchantOrderNotFound
	}
	if one["status"].String() == OrderPaid {
		return nil
	}
	merchantID := one["merchant_id"].Int64()
	productCode := one["product_code"].String()
	pOne, err := g.DB().Model("merchant_product").Ctx(ctx).Where("product_code", productCode).Limit(1).One()
	if err != nil {
		return err
	}
	period := MerchantDefaultPeriodDays
	if !pOne.IsEmpty() && pOne["period_days"].Int() > 0 {
		period = pOne["period_days"].Int()
	}
	now := time.Now().Unix()
	_, err = g.DB().Model("merchant_order").Ctx(ctx).Where("order_no", orderNo).Data(g.Map{
		"status":         OrderPaid,
		"channel_txn_id": strings.TrimSpace(channelTxnID),
		"paid_at":        now,
	}).Update()
	if err != nil {
		return err
	}
	_, err = ExtendMerchantExpireAt(ctx, merchantID, period, MerchantEntitlementSourceAlipay)
	return err
}
