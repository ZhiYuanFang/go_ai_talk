package controller

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/services/cash"
	"hello/internal/services/gatewayapp"
	"hello/internal/services/voice"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// CashFeatureController 商业功能开通 App/Admin API（宿主 cash-service）。
type CashFeatureController struct{}

// cashDeviceNoFromHeader 只信网关注入的 X-Internal-Device-No。
func cashDeviceNoFromHeader(ctx context.Context) (string, error) {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "缺少设备号")
	}
	dn := strings.TrimSpace(r.GetHeader(gatewayapp.HeaderInternalDeviceNo))
	if dn == "" {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "缺少设备号（须登录并绑定设备）")
	}
	return dn, nil
}

func cashWxIDOptional(ctx context.Context) int64 {
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return 0
	}
	return voice.ParseHeaderWxID(r.GetHeader(gatewayapp.HeaderInternalWxId))
}

// UCGEligibility GET /cash/app/api/ucg/eligibility
func (c *CashFeatureController) UCGEligibility(ctx context.Context, _ *v1.CashUCGEligibilityReq) (*v1.CashUCGEligibilityRes, error) {
	dn, err := cashDeviceNoFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	out, err := cash.GetUCGEligibility(ctx, dn)
	if err != nil {
		return nil, err
	}
	return &v1.CashUCGEligibilityRes{
		Qualified: out.Qualified, RequiredDays: out.RequiredDays,
		EffectiveDays: out.EffectiveDays, RemainingDays: out.RemainingDays, Message: out.Message,
	}, nil
}

// Catalog GET /cash/app/api/feature/catalog
func (c *CashFeatureController) Catalog(ctx context.Context, _ *v1.CashFeatureCatalogReq) (*v1.CashFeatureCatalogRes, error) {
	dn, err := cashDeviceNoFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	list, err := cash.GetFeatureCatalog(ctx, dn)
	if err != nil {
		return nil, err
	}
	res := &v1.CashFeatureCatalogRes{List: make([]v1.CashFeatureCatalogItem, 0, len(list))}
	for _, it := range list {
		item := v1.CashFeatureCatalogItem{
			FeatureId: it.FeatureId, Title: it.Title, Description: it.Description,
			UnlockMethods: it.UnlockMethods, Unlocked: it.Unlocked,
			UnlockMethod: it.UnlockMethod, ExpiresAt: it.ExpiresAt, AllowedCount: it.AllowedCount,
			Products: make([]v1.CashFeatureCatalogProductItem, 0, len(it.Products)),
		}
		for _, p := range it.Products {
			item.Products = append(item.Products, v1.CashFeatureCatalogProductItem{
				ProductCode: p.ProductCode, PriceFen: p.PriceFen, OriginalPriceFen: p.OriginalPriceFen,
				DurationDays: p.DurationDays, GrantKind: p.GrantKind, GrantQuantity: p.GrantQuantity,
				AppleProductId: p.AppleProductId,
			})
		}
		res.List = append(res.List, item)
	}
	return res, nil
}

// Orders POST /cash/app/api/feature/orders
func (c *CashFeatureController) Orders(ctx context.Context, req *v1.CashFeatureCreateOrderReq) (*v1.CashFeatureCreateOrderRes, error) {
	dn, err := cashDeviceNoFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	wxID := cashWxIDOptional(ctx)
	out, err := cash.CreateFeatureOrder(ctx, dn, wxID, req.ProductCode, req.Channel)
	if err != nil {
		return nil, err
	}
	return &v1.CashFeatureCreateOrderRes{
		OrderNo: out.OrderNo, ProductCode: out.ProductCode, Channel: out.Channel,
		AmountFen: out.AmountFen, AppleProductId: out.AppleProductId,
		AlipayOrderStr: out.AlipayOrderStr, PayTip: out.PayTip,
	}, nil
}

// InviteCodesRedeem POST /cash/app/api/feature/invite-codes/redeem
func (c *CashFeatureController) InviteCodesRedeem(ctx context.Context, req *v1.CashFeatureInviteRedeemReq) (*v1.CashFeatureInviteRedeemRes, error) {
	dn, err := cashDeviceNoFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	wxID, err := cashWxIDFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	if err := cash.RedeemInviteCode(ctx, wxID, dn, req.Code, req.FeatureId); err != nil {
		return nil, err
	}
	return &v1.CashFeatureInviteRedeemRes{}, nil
}

// AdComplete POST /cash/app/api/feature/ad/complete
func (c *CashFeatureController) AdComplete(ctx context.Context, req *v1.CashFeatureAdCompleteReq) (*v1.CashFeatureAdCompleteRes, error) {
	dn, err := cashDeviceNoFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	if err := cash.CompleteFeatureAd(ctx, dn, req.FeatureId, req.IdempotencyKey, 1, 0); err != nil {
		return nil, err
	}
	return &v1.CashFeatureAdCompleteRes{}, nil
}

// —— Admin ——

// AdminFeatureDefs GET
func (c *CashFeatureController) AdminFeatureDefs(ctx context.Context, _ *v1.CashAdminFeatureDefsListReq) (*v1.CashAdminFeatureDefsListRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	list, err := cash.AdminListFeatureDefs(ctx)
	if err != nil {
		return nil, err
	}
	res := &v1.CashAdminFeatureDefsListRes{List: make([]v1.CashAdminFeatureDefItem, 0, len(list))}
	for _, it := range list {
		res.List = append(res.List, v1.CashAdminFeatureDefItem{
			FeatureId: it.FeatureId, Title: it.Title, Description: it.Description,
			UnlockMethods: it.UnlockMethods, DurationDays: it.DurationDays,
			Status: it.Status, SortOrder: it.SortOrder,
		})
	}
	return res, nil
}

// AdminFeatureDefsUpsert POST
func (c *CashFeatureController) AdminFeatureDefsUpsert(ctx context.Context, req *v1.CashAdminFeatureDefUpsertReq) (*v1.CashAdminFeatureDefUpsertRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	if err := cash.AdminUpsertFeatureDef(ctx, req.FeatureId, req.Title, req.Description, req.UnlockMethods, req.DurationDays, req.Status, req.SortOrder); err != nil {
		return nil, err
	}
	return &v1.CashAdminFeatureDefUpsertRes{}, nil
}

// AdminFeatureProducts GET
func (c *CashFeatureController) AdminFeatureProducts(ctx context.Context, _ *v1.CashAdminFeatureProductsListReq) (*v1.CashAdminFeatureProductsListRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	list, err := cash.AdminListFeatureProducts(ctx)
	if err != nil {
		return nil, err
	}
	res := &v1.CashAdminFeatureProductsListRes{List: make([]v1.CashAdminFeatureProductItem, 0, len(list))}
	for _, it := range list {
		res.List = append(res.List, v1.CashAdminFeatureProductItem{
			ProductCode: it.ProductCode, FeatureId: it.FeatureId, GrantKind: it.GrantKind,
			GrantQuantity: it.GrantQuantity, PriceFen: it.PriceFen, OriginalPriceFen: it.OriginalPriceFen,
			DurationDays: it.DurationDays, AppleProductId: it.AppleProductId, Status: it.Status,
		})
	}
	return res, nil
}

// AdminFeatureProductsUpsert POST
func (c *CashFeatureController) AdminFeatureProductsUpsert(ctx context.Context, req *v1.CashAdminFeatureProductUpsertReq) (*v1.CashAdminFeatureProductUpsertRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	err := cash.AdminUpsertFeatureProduct(ctx, &cash.FeatureProduct{
		ProductCode: req.ProductCode, FeatureId: req.FeatureId, GrantKind: req.GrantKind,
		GrantQuantity: req.GrantQuantity, PriceFen: req.PriceFen, OriginalPriceFen: req.OriginalPriceFen,
		DurationDays: req.DurationDays, AppleProductId: req.AppleProductId, Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CashAdminFeatureProductUpsertRes{}, nil
}

// AdminInviteCodesList GET
func (c *CashFeatureController) AdminInviteCodesList(ctx context.Context, _ *v1.CashAdminInviteCodesListReq) (*v1.CashAdminInviteCodesListRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	list, err := cash.AdminListInviteCodes(ctx)
	if err != nil {
		return nil, err
	}
	res := &v1.CashAdminInviteCodesListRes{List: make([]v1.CashAdminInviteCodeItem, 0, len(list))}
	for _, it := range list {
		res.List = append(res.List, v1.CashAdminInviteCodeItem{
			Code: it.Code, OwnerWxId: it.OwnerWxId, ExpiresAt: it.ExpiresAt,
			MaxRedemptions: it.MaxRedemptions, RedeemedCount: it.RedeemedCount,
			GrantDurationDays: it.GrantDurationDays, Status: it.Status,
			FeatureIds: it.FeatureIds, CreatedAt: it.CreatedAt,
		})
	}
	return res, nil
}

// AdminInviteCodeCreate POST
func (c *CashFeatureController) AdminInviteCodeCreate(ctx context.Context, req *v1.CashAdminInviteCodeCreateReq) (*v1.CashAdminInviteCodeCreateRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	code, err := cash.AdminCreateInviteCode(ctx, req.OwnerWxId, req.Code, req.ExpiresAt, req.MaxRedemptions, req.GrantDurationDays, req.Status, req.FeatureIds)
	if err != nil {
		return nil, err
	}
	return &v1.CashAdminInviteCodeCreateRes{Code: code}, nil
}

// AdminInviteCodeStatus POST
func (c *CashFeatureController) AdminInviteCodeStatus(ctx context.Context, req *v1.CashAdminInviteCodeStatusReq) (*v1.CashAdminInviteCodeStatusRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	if err := cash.AdminUpdateInviteCodeStatus(ctx, req.Code, req.Status); err != nil {
		return nil, err
	}
	return &v1.CashAdminInviteCodeStatusRes{}, nil
}

// AdminInviteRedemptions GET
func (c *CashFeatureController) AdminInviteRedemptions(ctx context.Context, req *v1.CashAdminInviteRedemptionsReq) (*v1.CashAdminInviteRedemptionsRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	list, err := cash.AdminListInviteRedemptions(ctx, req.Code, req.Limit)
	if err != nil {
		return nil, err
	}
	res := &v1.CashAdminInviteRedemptionsRes{List: make([]v1.CashAdminInviteRedemptionItem, 0, len(list))}
	for _, it := range list {
		res.List = append(res.List, v1.CashAdminInviteRedemptionItem{
			Code: it.Code, OwnerWxId: it.OwnerWxId, RedeemerWxId: it.RedeemerWxId,
			DeviceNo: it.DeviceNo, FeatureId: it.FeatureId, RedeemedAt: it.RedeemedAt,
		})
	}
	return res, nil
}
