package cashctrl

import (
	"hello/internal/platform/httpmeta"
	"context"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/services/cash"

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
	dn := strings.TrimSpace(r.GetHeader(httpmeta.HeaderInternalDeviceNo))
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
	return httpmeta.ParseHeaderWxID(r.GetHeader(httpmeta.HeaderInternalWxId))
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

// CareAlertEligibility GET /cash/app/api/care-alert/eligibility
func (c *CashFeatureController) CareAlertEligibility(ctx context.Context, _ *v1.CashCareAlertEligibilityReq) (*v1.CashCareAlertEligibilityRes, error) {
	dn, err := cashDeviceNoFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	out, err := cash.GetCareAlertFeedingEligibility(ctx, dn)
	if err != nil {
		return nil, err
	}
	return &v1.CashCareAlertEligibilityRes{
		Qualified: out.Qualified, RequiredDays: out.RequiredDays,
		EffectiveDays: out.EffectiveDays, RemainingDays: out.RemainingDays, Message: out.Message,
	}, nil
}

// InternalCareAlertAccess GET /cash/internal/api/care-alert/access
// 供 voice 双门禁：喂养资格 ∧（设备开通 ∨ VIP）；须内部密钥。
func (c *CashFeatureController) InternalCareAlertAccess(ctx context.Context, req *v1.CashInternalCareAlertAccessReq) (*v1.CashInternalCareAlertAccessRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	if !cash.ValidateInternalSecret(cash.InternalSecretFromRequest(r)) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "内部接口未授权")
	}
	dn := strings.TrimSpace(req.DeviceNo)
	if dn == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if req.WxId <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	out, err := cash.GetCareAlertAccess(ctx, dn, req.WxId)
	if err != nil {
		return nil, err
	}
	return &v1.CashInternalCareAlertAccessRes{
		Allowed: out.Allowed, FeedingQualified: out.FeedingQualified,
		FeatureActive: out.FeatureActive, EntitlementExpiresAt: out.EntitlementExpiresAt,
	}, nil
}

// Catalog GET /cash/app/api/feature/catalog
func (c *CashFeatureController) Catalog(ctx context.Context, _ *v1.CashFeatureCatalogReq) (*v1.CashFeatureCatalogRes, error) {
	dn, err := cashDeviceNoFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	cat, err := cash.GetFeatureCatalog(ctx, dn)
	if err != nil {
		return nil, err
	}
	res := &v1.CashFeatureCatalogRes{
		List:             make([]v1.CashFeatureCatalogItem, 0, len(cat.List)),
		InviteGroupQrUrl: cat.InviteGroupQrUrl,
	}
	for _, it := range cat.List {
		item := v1.CashFeatureCatalogItem{
			FeatureId: it.FeatureId, Title: it.Title, Description: it.Description,
			UnlockMethods: it.UnlockMethods, Unlocked: it.Unlocked,
			UnlockMethod: it.UnlockMethod, ExpiresAt: it.ExpiresAt, AllowedCount: it.AllowedCount,
			TotalActivatableCount: it.TotalActivatableCount,
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

// InviteMine GET /cash/app/api/invite/mine
func (c *CashFeatureController) InviteMine(ctx context.Context, _ *v1.CashInviteMineReq) (*v1.CashInviteMineRes, error) {
	wxID, err := cashWxIDFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	out, err := cash.GetInviteMine(ctx, wxID)
	if err != nil {
		return nil, err
	}
	return &v1.CashInviteMineRes{Code: out.Code, RedeemedCount: out.RedeemedCount}, nil
}

// InviteInvitees GET /cash/app/api/invite/invitees
func (c *CashFeatureController) InviteInvitees(ctx context.Context, _ *v1.CashInviteInviteesReq) (*v1.CashInviteInviteesRes, error) {
	wxID, err := cashWxIDFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	list, err := cash.ListInviteInvitees(ctx, wxID)
	if err != nil {
		return nil, err
	}
	res := &v1.CashInviteInviteesRes{List: make([]v1.CashInviteInviteeItem, 0, len(list))}
	for _, it := range list {
		res.List = append(res.List, v1.CashInviteInviteeItem{
			WxId: it.WxId, Nickname: it.Nickname, RedeemedAt: it.RedeemedAt,
		})
	}
	return res, nil
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
			DefaultAllowedCount: it.DefaultAllowedCount,
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
	if err := cash.AdminUpdateFeatureDef(ctx, req.FeatureId, req.Title, req.Description, req.UnlockMethods, req.DurationDays, req.Status, req.SortOrder, req.DefaultAllowedCount); err != nil {
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
	prod := &cash.FeatureProduct{
		ProductCode: req.ProductCode, FeatureId: req.FeatureId, GrantKind: req.GrantKind,
		GrantQuantity: req.GrantQuantity, PriceFen: req.PriceFen, OriginalPriceFen: req.OriginalPriceFen,
		DurationDays: req.DurationDays, AppleProductId: req.AppleProductId, Status: req.Status,
	}
	if err := cash.AdminUpsertFeatureProduct(ctx, prod); err != nil {
		return nil, err
	}
	return &v1.CashAdminFeatureProductUpsertRes{ProductCode: prod.ProductCode}, nil
}

// AdminFeedingEligibilityScenes GET
func (c *CashFeatureController) AdminFeedingEligibilityScenes(ctx context.Context, _ *v1.CashAdminFeedingEligibilityScenesListReq) (*v1.CashAdminFeedingEligibilityScenesListRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	list, err := cash.AdminListFeedingEligibilityScenes(ctx)
	if err != nil {
		return nil, err
	}
	res := &v1.CashAdminFeedingEligibilityScenesListRes{List: make([]v1.CashAdminFeedingEligibilitySceneItem, 0, len(list))}
	for _, it := range list {
		res.List = append(res.List, v1.CashAdminFeedingEligibilitySceneItem{
			SceneKey: it.SceneKey, RequiredDays: it.RequiredDays,
			MinRecordsPerDay: it.MinRecordsPerDay, UpdatedAt: it.UpdatedAt,
		})
	}
	return res, nil
}

// AdminFeedingEligibilityScenesUpdate POST
func (c *CashFeatureController) AdminFeedingEligibilityScenesUpdate(ctx context.Context, req *v1.CashAdminFeedingEligibilitySceneUpdateReq) (*v1.CashAdminFeedingEligibilitySceneUpdateRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	if err := cash.AdminUpdateFeedingEligibilityScene(ctx, req.SceneKey, req.RequiredDays, req.MinRecordsPerDay); err != nil {
		return nil, err
	}
	return &v1.CashAdminFeedingEligibilitySceneUpdateRes{}, nil
}

// AdminInviteGroupQrGet GET /cash/admin/api/invite-group-qr
func (c *CashFeatureController) AdminInviteGroupQrGet(ctx context.Context, _ *v1.CashAdminInviteGroupQrGetReq) (*v1.CashAdminInviteGroupQrGetRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	out, err := cash.GetInviteGroupQrAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.CashAdminInviteGroupQrGetRes{
		FileName: out.FileName, ExpiresAt: out.ExpiresAt, UpdatedAt: out.UpdatedAt,
		PreviewPath: out.PreviewPath, AppVisible: out.AppVisible,
	}, nil
}

// AdminInviteGroupQrPut POST /cash/admin/api/invite-group-qr
func (c *CashFeatureController) AdminInviteGroupQrPut(ctx context.Context, req *v1.CashAdminInviteGroupQrPutReq) (*v1.CashAdminInviteGroupQrPutRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	if req.TouchUpdated {
		if err := cash.TouchInviteGroupQrUpdated(ctx); err != nil {
			return nil, err
		}
	}
	if req.ExpiresAt != nil {
		if err := cash.SetInviteGroupQrExpires(ctx, *req.ExpiresAt); err != nil {
			return nil, err
		}
	}
	out, err := cash.GetInviteGroupQrAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.CashAdminInviteGroupQrPutRes{
		FileName: out.FileName, ExpiresAt: out.ExpiresAt, UpdatedAt: out.UpdatedAt,
		PreviewPath: out.PreviewPath, AppVisible: out.AppVisible,
	}, nil
}
