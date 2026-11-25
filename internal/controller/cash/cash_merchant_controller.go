package cashctrl

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/platform/httpmeta"
	"hello/internal/services/cash"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// CashMerchantController 商户订阅商品/权益 + Hub + 内部探活。
type CashMerchantController struct{}

func merchantIDFromCashHeader(r *ghttp.Request) (int64, error) {
	raw := strings.TrimSpace(r.GetHeader(httpmeta.HeaderInternalMerchantId))
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, gerror.NewCode(gcode.CodeNotAuthorized, "缺少商户身份")
	}
	return id, nil
}

// Product GET /cash/merchant/api/product
func (c *CashMerchantController) Product(ctx context.Context, req *v1.CashMerchantProductReq) (*v1.CashMerchantProductRes, error) {
	_ = c
	_ = req
	p, err := cash.GetMerchantMonthlyProduct(ctx)
	if err != nil {
		if errors.Is(err, cash.ErrMerchantProductMissing) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		return nil, err
	}
	return &v1.CashMerchantProductRes{
		ProductCode: p.ProductCode,
		Title:       p.Title,
		PriceFen:    p.PriceFen,
		PeriodDays:  p.PeriodDays,
	}, nil
}

// Subscribe POST /cash/merchant/api/subscribe
func (c *CashMerchantController) Subscribe(ctx context.Context, req *v1.CashMerchantSubscribeReq) (*v1.CashMerchantSubscribeRes, error) {
	_ = req
	r := ghttp.RequestFromCtx(ctx)
	mid, err := merchantIDFromCashHeader(r)
	if err != nil {
		return nil, err
	}
	orderNo, amount, payload, err := cash.CreateMerchantAlipayOrder(ctx, mid)
	if err != nil {
		if errors.Is(err, cash.ErrMerchantProductMissing) || errors.Is(err, cash.ErrMerchantIDInvalid) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		return nil, err
	}
	tip := ""
	if strings.Contains(payload, "placeholder") {
		tip = "开发占位：配置支付宝网页/周期扣款后替换 payPayload；DEV_BYPASS 可模拟 notify 履约"
	}
	return &v1.CashMerchantSubscribeRes{
		OrderNo: orderNo, AmountFen: amount, PayPayload: payload, Tip: tip,
	}, nil
}

// SubscribeSimulatePaid POST /cash/merchant/api/subscribe/simulate-paid（仅 CASH_PAYMENT_DEV_BYPASS=1）
func (c *CashMerchantController) SubscribeSimulatePaid(ctx context.Context, req *v1.CashMerchantSubscribeSimulateReq) (*v1.CashMerchantSubscribeSimulateRes, error) {
	_ = c
	r := ghttp.RequestFromCtx(ctx)
	if _, err := merchantIDFromCashHeader(r); err != nil {
		return nil, err
	}
	if strings.TrimSpace(os.Getenv("CASH_PAYMENT_DEV_BYPASS")) != "1" {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "仅开发旁路可用")
	}
	form := map[string]string{
		"out_trade_no":  strings.TrimSpace(req.OrderNo),
		"trade_no":      "SIM" + strings.TrimSpace(req.OrderNo),
		"trade_status":  "TRADE_SUCCESS",
		"total_amount":  "0.01", // 金额校验：若订单 amount>0 且不一致会失败；旁路用订单金额覆盖
	}
	// 用订单真实金额填 total_amount
	one, err := cash.PeekMerchantOrderAmountYuan(ctx, req.OrderNo)
	if err == nil && one != "" {
		form["total_amount"] = one
	}
	ack, err := cash.HandleAlipayNotify(ctx, form)
	if err != nil || ack != "success" {
		if err != nil {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, "simulate failed")
	}
	return &v1.CashMerchantSubscribeSimulateRes{Ok: true}, nil
}

// Entitlement GET /cash/merchant/api/entitlement
func (c *CashMerchantController) Entitlement(ctx context.Context, req *v1.CashMerchantEntitlementReq) (*v1.CashMerchantEntitlementRes, error) {
	_ = req
	r := ghttp.RequestFromCtx(ctx)
	mid, err := merchantIDFromCashHeader(r)
	if err != nil {
		return nil, err
	}
	st, err := cash.GetMerchantEntitlementStatus(ctx, mid)
	if err != nil {
		return nil, err
	}
	return &v1.CashMerchantEntitlementRes{
		ExpireAt: st.ExpireAt,
		Source:   st.Source,
		Entitled: st.Entitled,
		Arrears:  st.Arrears,
	}, nil
}

// AdminList GET /cash/admin/api/merchant/entitlements
func (c *CashMerchantController) AdminList(ctx context.Context, req *v1.CashAdminMerchantListReq) (*v1.CashAdminMerchantListRes, error) {
	_ = c
	_ = req
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	list, err := cash.ListMerchantEntitlements(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]v1.CashAdminMerchantItem, 0, len(list))
	for _, it := range list {
		items = append(items, v1.CashAdminMerchantItem{
			MerchantId: it.MerchantID,
			ExpireAt:   it.ExpireAt,
			Source:     it.Source,
			Entitled:   it.Entitled,
			Arrears:    it.Arrears,
		})
	}
	return &v1.CashAdminMerchantListRes{Items: items}, nil
}

// AdminGrant POST /cash/admin/api/merchant/grant
func (c *CashMerchantController) AdminGrant(ctx context.Context, req *v1.CashAdminMerchantGrantReq) (*v1.CashAdminMerchantGrantRes, error) {
	_ = c
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	if err := cash.SetMerchantExpireAtAbsolute(ctx, req.MerchantId, req.ExpireAt, cash.MerchantEntitlementSourceHub); err != nil {
		if errors.Is(err, cash.ErrMerchantExpireAtInvalid) || errors.Is(err, cash.ErrMerchantIDInvalid) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		return nil, err
	}
	return &v1.CashAdminMerchantGrantRes{ExpireAt: req.ExpireAt}, nil
}

// AdminPrice POST /cash/admin/api/merchant/price
func (c *CashMerchantController) AdminPrice(ctx context.Context, req *v1.CashAdminMerchantPriceReq) (*v1.CashAdminMerchantPriceRes, error) {
	_ = c
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	if err := cash.AdminUpdateMerchantMonthlyPrice(ctx, req.PriceFen); err != nil {
		if errors.Is(err, cash.ErrMerchantPriceInvalid) {
			return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, err.Error())
		}
		return nil, err
	}
	return &v1.CashAdminMerchantPriceRes{PriceFen: req.PriceFen}, nil
}

// InternalEntitled GET /cash/internal/api/merchant/entitled
func (c *CashMerchantController) InternalEntitled(ctx context.Context, req *v1.CashInternalMerchantEntitledReq) (*v1.CashInternalMerchantEntitledRes, error) {
	_ = c
	r := ghttp.RequestFromCtx(ctx)
	if !cash.ValidateInternalSecret(cash.InternalSecretFromRequest(r)) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "内部接口未授权")
	}
	ok, exp, err := cash.IsMerchantEntitled(ctx, req.MerchantId)
	if err != nil {
		return nil, err
	}
	return &v1.CashInternalMerchantEntitledRes{Entitled: ok, ExpireAt: exp}, nil
}
