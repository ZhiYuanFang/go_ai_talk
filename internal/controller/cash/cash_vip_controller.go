package cashctrl

import (
	"context"
	"fmt"
	"hello/internal/platform/httpmeta"
	"net/http"
	"strconv"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/services/cash"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// CashVipController VIP 商品/订单/支付/权益（宿主 cash-service）。
type CashVipController struct{}

func cashWxIDFromHeader(ctx context.Context) (int64, error) {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	wxID := httpmeta.ParseHeaderWxID(r.GetHeader(httpmeta.HeaderInternalWxId))
	if wxID <= 0 {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	return wxID, nil
}

// Product GET /cash/app/api/vip/product — 匿名可读现价/原价；建单仍须登录。
func (c *CashVipController) Product(ctx context.Context, req *v1.CashVipProductReq) (res *v1.CashVipProductRes, err error) {
	prod, err := cash.GetActiveProduct(ctx, cash.ProductMonthly19)
	if err != nil {
		return nil, err
	}
	return &v1.CashVipProductRes{
		ProductCode:      prod.ProductCode,
		Title:            prod.Title,
		PriceFen:         prod.PriceFen,
		OriginalPriceFen: prod.OriginalPriceFen,
		DurationDays:     prod.DurationDays,
		AppleProductId:   prod.AppleProductId,
	}, nil
}

// Status GET /cash/app/api/vip/status
func (c *CashVipController) Status(ctx context.Context, req *v1.CashVipStatusReq) (res *v1.CashVipStatusRes, err error) {
	wxID, err := cashWxIDFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	st, err := cash.GetVipStatus(ctx, wxID)
	if err != nil {
		return nil, err
	}
	return &v1.CashVipStatusRes{WxId: st.WxId, IsVip: st.IsVip, ExpireAt: st.ExpireAt}, nil
}

// Orders POST /cash/app/api/vip/orders
func (c *CashVipController) Orders(ctx context.Context, req *v1.CashVipCreateOrderReq) (res *v1.CashVipCreateOrderRes, err error) {
	wxID, err := cashWxIDFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	out, err := cash.CreateOrder(ctx, wxID, req.ProductCode, req.Channel)
	if err != nil {
		return nil, err
	}
	return &v1.CashVipCreateOrderRes{
		OrderNo:         out.OrderNo,
		ProductCode:     out.ProductCode,
		Channel:         out.Channel,
		AmountFen:       out.AmountFen,
		AppleProductId:  out.AppleProductId,
		AppAccountToken: out.AppAccountToken,
		AlipayOrderStr:  out.AlipayOrderStr,
		PayTip:          out.PayTip,
	}, nil
}

// AppleVerify POST /cash/app/api/vip/apple/verify
// 可选加速履约；权威路径为 ASN（…/apple/notifications），二者共用 Fulfill* 幂等。
func (c *CashVipController) AppleVerify(ctx context.Context, req *v1.CashVipAppleVerifyReq) (res *v1.CashVipAppleVerifyRes, err error) {
	wxID, err := cashWxIDFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	if err := cash.VerifyAppleIAP(ctx, wxID, cash.AppleVerifyInput{
		OrderNo:           req.OrderNo,
		TransactionId:     req.TransactionId,
		ProductId:         req.ProductId,
		SignedTransaction: req.SignedTransaction,
	}); err != nil {
		return nil, err
	}
	return &v1.CashVipAppleVerifyRes{}, nil
}

// InternalVipByWxID GET /cash/internal/api/vip/by-wx-id
func (c *CashVipController) InternalVipByWxID(ctx context.Context, req *v1.CashInternalVipByWxIDReq) (res *v1.CashInternalVipByWxIDRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	if !cash.ValidateInternalSecret(cash.InternalSecretFromRequest(r)) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "内部接口未授权")
	}
	if req.WxId <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	st, err := cash.GetVipStatus(ctx, req.WxId)
	if err != nil {
		return nil, err
	}
	return &v1.CashInternalVipByWxIDRes{WxId: st.WxId, IsVip: st.IsVip, ExpireAt: st.ExpireAt}, nil
}

// requireCashAdmin 校验网关注入的 X-Admin-Password（浏览器不得自带）。
func requireCashAdmin(ctx context.Context) error {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return gerror.NewCode(gcode.CodeNotAuthorized, "口令错误")
	}
	if !cash.VerifyCashAdminPassword(strings.TrimSpace(r.GetHeader(cash.HeaderAdminPassword))) {
		return gerror.NewCode(gcode.CodeNotAuthorized, "口令错误")
	}
	return nil
}

// AdminVipProduct GET /cash/admin/api/vip/product — 开通功能管理读取一期 VIP 套餐。
func (c *CashVipController) AdminVipProduct(ctx context.Context, _ *v1.CashAdminVipProductGetReq) (*v1.CashAdminVipProductGetRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	prod, err := cash.GetVipProductForAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.CashAdminVipProductGetRes{
		ProductCode:      prod.ProductCode,
		Title:            prod.Title,
		PriceFen:         prod.PriceFen,
		OriginalPriceFen: prod.OriginalPriceFen,
		DurationDays:     prod.DurationDays,
		AppleProductId:   prod.AppleProductId,
		Status:           prod.Status,
	}, nil
}

// AdminVipProductUpsert POST /cash/admin/api/vip/product — 仅更新 vip_monthly_19。
func (c *CashVipController) AdminVipProductUpsert(ctx context.Context, req *v1.CashAdminVipProductUpsertReq) (*v1.CashAdminVipProductUpsertRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	if err := cash.AdminUpdateVipProduct(ctx, cash.AdminUpdateVipProductInput{
		ProductCode:      req.ProductCode,
		Title:            req.Title,
		PriceFen:         req.PriceFen,
		OriginalPriceFen: req.OriginalPriceFen,
		DurationDays:     req.DurationDays,
		AppleProductId:   req.AppleProductId,
		Status:           req.Status,
	}); err != nil {
		return nil, err
	}
	return &v1.CashAdminVipProductUpsertRes{ProductCode: cash.ProductMonthly19}, nil
}

// AdminVipEntitlements GET /cash/admin/api/vip/entitlements — 只读分页列表（含已过期 + 最近实付）。
//
// 业务逻辑：口令失败直接拒绝；不写库；响应字段见 design D2。
func (c *CashVipController) AdminVipEntitlements(ctx context.Context, req *v1.CashAdminVipEntitlementsReq) (res *v1.CashAdminVipEntitlementsRes, err error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := cash.ListEntitlementsForAdmin(ctx, req.Page, req.PageSize, req.WxId)
	if err != nil {
		return nil, err
	}
	out := &v1.CashAdminVipEntitlementsRes{
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
		List:     make([]v1.CashAdminVipEntitlementItem, 0, len(result.List)),
	}
	for _, row := range result.List {
		out.List = append(out.List, v1.CashAdminVipEntitlementItem{
			WxId:              row.WxId,
			IsVip:             row.IsVip,
			ExpireAt:          row.ExpireAt,
			RemainingSeconds:  row.RemainingSeconds,
			LastPaidAmountFen: row.LastPaidAmountFen,
			Channel:           row.Channel,
			PaidAt:            row.PaidAt,
			GrantReason:       row.GrantReason,
		})
	}
	return out, nil
}

// AdminVipGrant POST /cash/admin/api/vip/entitlements/grant — 手工授 VIP。
func (c *CashVipController) AdminVipGrant(ctx context.Context, req *v1.CashAdminVipGrantReq) (*v1.CashAdminVipGrantRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	out, err := cash.AdminGrantVip(ctx, req.WxId, req.DurationDays, req.Reason)
	if err != nil {
		return nil, err
	}
	return &v1.CashAdminVipGrantRes{OrderNo: out.OrderNo, WxId: out.WxId, ExpireAt: out.ExpireAt}, nil
}

// registerCashAlipayNotify 支付宝 notify 需返回纯文本 success，不用标准 JSON envelope。
func RegisterAlipayNotify(s *ghttp.Server) {
	s.BindHandler("POST:/cash/app/api/vip/alipay/notify", func(r *ghttp.Request) {
		_ = r.Request.ParseForm()
		form := map[string]string{}
		for k, vals := range r.Request.Form {
			if len(vals) > 0 {
				form[k] = vals[0]
			}
		}
		if len(form) == 0 {
			var body map[string]interface{}
			if err := r.Parse(&body); err == nil {
				for k, v := range body {
					form[k] = stringifyNotifyVal(v)
				}
			}
		}
		ack, err := cash.HandleAlipayNotify(r.Context(), form)
		if err != nil || ack != "success" {
			r.Response.WriteStatus(200)
			r.Response.Write("failure")
			return
		}
		r.Response.WriteStatus(200)
		r.Response.Write("success")
	})
}

// RegisterAppleNotifications Apple ASN V2：读 JSON signedPayload，同步验签履约；失败非 2xx 触发苹果重试。
// 无 MQ、无标准 JSON envelope（对齐支付宝 notify 特殊写出风格）。
func RegisterAppleNotifications(s *ghttp.Server) {
	s.BindHandler("POST:/cash/app/api/vip/apple/notifications", func(r *ghttp.Request) {
		var body struct {
			SignedPayload string `json:"signedPayload"`
		}
		if err := r.Parse(&body); err != nil || strings.TrimSpace(body.SignedPayload) == "" {
			// 兼容原始 body 仅为 JWS 字符串的误配（少见）
			raw := strings.TrimSpace(r.GetBodyString())
			if strings.HasPrefix(raw, "eyJ") {
				body.SignedPayload = raw
			}
		}
		if err := cash.HandleAppleNotification(r.Context(), body.SignedPayload); err != nil {
			r.Response.WriteStatus(http.StatusInternalServerError)
			r.Response.Write([]byte(err.Error()))
			return
		}
		r.Response.WriteStatus(http.StatusOK)
		r.Response.Write([]byte("ok"))
	})
}

func stringifyNotifyVal(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}
