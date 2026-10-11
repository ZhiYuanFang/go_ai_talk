package cashctrl

// cash_xiaozhi_mcp_controller.go：小智 MCP 永久开通 App / Internal / Admin API。
//
// 业务：独立于开通功能管理；支付建单走共用回调；device 经 internal 查开通态。

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/services/cash"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// CashXiaozhiMcpController 小智 MCP 开通控制器（宿主 cash-service）。
type CashXiaozhiMcpController struct{}

func mapXiaozhiMcpProductItem(p *cash.XiaozhiMcpProduct) *v1.CashXiaozhiMcpProductItem {
	if p == nil {
		return nil
	}
	return &v1.CashXiaozhiMcpProductItem{
		ProductCode:      p.ProductCode,
		Title:            p.Title,
		PriceFen:         p.PriceFen,
		OriginalPriceFen: p.OriginalPriceFen,
		AppleProductId:   p.AppleProductId,
		Status:           p.Status,
	}
}

// Unlock GET /cash/app/api/xiaozhi-mcp/unlock
func (c *CashXiaozhiMcpController) Unlock(ctx context.Context, _ *v1.CashXiaozhiMcpUnlockReq) (*v1.CashXiaozhiMcpUnlockRes, error) {
	wxID, err := cashWxIDFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	view, err := cash.GetXiaozhiMcpUnlockView(ctx, wxID)
	if err != nil {
		return nil, err
	}
	return &v1.CashXiaozhiMcpUnlockRes{
		Unlocked:       view.Unlocked,
		TrialAvailable: view.TrialAvailable,
		ExpiresAt:      view.ExpiresAt,
		Product:        mapXiaozhiMcpProductItem(view.Product),
	}, nil
}

// Orders POST /cash/app/api/xiaozhi-mcp/orders
func (c *CashXiaozhiMcpController) Orders(ctx context.Context, req *v1.CashXiaozhiMcpCreateOrderReq) (*v1.CashXiaozhiMcpCreateOrderRes, error) {
	wxID, err := cashWxIDFromHeader(ctx)
	if err != nil {
		return nil, err
	}
	// deviceNo 可选：有绑定时写入订单审查字段。
	deviceNo := ""
	if dn, dErr := cashDeviceNoFromHeader(ctx); dErr == nil {
		deviceNo = dn
	}
	out, err := cash.CreateXiaozhiMcpOrder(ctx, deviceNo, wxID, req.Channel)
	if err != nil {
		return nil, err
	}
	return &v1.CashXiaozhiMcpCreateOrderRes{
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

// InternalEntitlement GET /cash/internal/api/xiaozhi-mcp/entitlement
func (c *CashXiaozhiMcpController) InternalEntitlement(ctx context.Context, req *v1.CashInternalXiaozhiMcpEntitlementReq) (*v1.CashInternalXiaozhiMcpEntitlementRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	if !cash.ValidateInternalSecret(cash.InternalSecretFromRequest(r)) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "内部接口未授权")
	}
	if req.WxId <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	ok, err := cash.HasActiveXiaozhiMcpEntitlement(ctx, req.WxId)
	if err != nil {
		return nil, err
	}
	return &v1.CashInternalXiaozhiMcpEntitlementRes{Unlocked: ok}, nil
}

// InternalEnsureAccessForAdd POST /cash/internal/api/xiaozhi-mcp/ensure-access-for-add
func (c *CashXiaozhiMcpController) InternalEnsureAccessForAdd(ctx context.Context, req *v1.CashInternalXiaozhiMcpEnsureAccessReq) (*v1.CashInternalXiaozhiMcpEnsureAccessRes, error) {
	r := ghttp.RequestFromCtx(ctx)
	if !cash.ValidateInternalSecret(cash.InternalSecretFromRequest(r)) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "内部接口未授权")
	}
	if err := cash.EnsureXiaozhiMcpAccessForAdd(ctx, req.WxId); err != nil {
		return nil, err
	}
	return &v1.CashInternalXiaozhiMcpEnsureAccessRes{}, nil
}

// AdminProductGet GET /cash/admin/api/xiaozhi-mcp/product
func (c *CashXiaozhiMcpController) AdminProductGet(ctx context.Context, _ *v1.CashAdminXiaozhiMcpProductGetReq) (*v1.CashAdminXiaozhiMcpProductGetRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	p, err := cash.AdminGetXiaozhiMcpProduct(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.CashAdminXiaozhiMcpProductGetRes{
		ProductCode:      p.ProductCode,
		Title:            p.Title,
		PriceFen:         p.PriceFen,
		OriginalPriceFen: p.OriginalPriceFen,
		AppleProductId:   p.AppleProductId,
		Status:           p.Status,
		UpdatedAt:        p.UpdatedAt,
	}, nil
}

// AdminProductPut POST /cash/admin/api/xiaozhi-mcp/product
func (c *CashXiaozhiMcpController) AdminProductPut(ctx context.Context, req *v1.CashAdminXiaozhiMcpProductPutReq) (*v1.CashAdminXiaozhiMcpProductPutRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	err := cash.AdminUpdateXiaozhiMcpProduct(ctx, cash.AdminUpdateXiaozhiMcpProductInput{
		Title:            strings.TrimSpace(req.Title),
		PriceFen:         req.PriceFen,
		OriginalPriceFen: req.OriginalPriceFen,
		AppleProductId:   req.AppleProductId,
		Status:           req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CashAdminXiaozhiMcpProductPutRes{}, nil
}

// AdminGrant POST /cash/admin/api/xiaozhi-mcp/grants
func (c *CashXiaozhiMcpController) AdminGrant(ctx context.Context, req *v1.CashAdminXiaozhiMcpGrantReq) (*v1.CashAdminXiaozhiMcpGrantRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	orderNo, err := cash.AdminGrantXiaozhiMcp(ctx, req.WxId, req.Reason)
	if err != nil {
		return nil, err
	}
	return &v1.CashAdminXiaozhiMcpGrantRes{OrderNo: orderNo, WxId: req.WxId}, nil
}

// AdminRevoke POST /cash/admin/api/xiaozhi-mcp/grants/revoke
func (c *CashXiaozhiMcpController) AdminRevoke(ctx context.Context, req *v1.CashAdminXiaozhiMcpRevokeReq) (*v1.CashAdminXiaozhiMcpRevokeRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	if err := cash.AdminRevokeXiaozhiMcp(ctx, req.WxId); err != nil {
		return nil, err
	}
	return &v1.CashAdminXiaozhiMcpRevokeRes{}, nil
}

// AdminEntitlements GET /cash/admin/api/xiaozhi-mcp/entitlements — 开通人员快照。
func (c *CashXiaozhiMcpController) AdminEntitlements(ctx context.Context, req *v1.CashAdminXiaozhiMcpEntitlementsReq) (*v1.CashAdminXiaozhiMcpEntitlementsRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	page, err := cash.AdminListXiaozhiMcpEntitlements(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}
	res := &v1.CashAdminXiaozhiMcpEntitlementsRes{
		Note:  page.Note,
		Total: page.Total,
		List:  make([]v1.CashAdminXiaozhiMcpEntitlementItem, 0, len(page.List)),
	}
	for _, it := range page.List {
		res.List = append(res.List, v1.CashAdminXiaozhiMcpEntitlementItem{
			WxId:             it.WxId,
			Nickname:         it.Nickname,
			UnlockMethod:     it.UnlockMethod,
			ChannelRef:       it.ChannelRef,
			UnlockedAt:       it.UnlockedAt,
			ExpiresAt:        it.ExpiresAt,
			Active:           it.Active,
			RemainingSeconds: it.RemainingSeconds,
			Status:           it.Status,
			RevokedAt:        it.RevokedAt,
			UpdatedAt:        it.UpdatedAt,
		})
	}
	return res, nil
}

func mapXiaozhiGuideItem(it cash.XiaozhiGuideVideoSlot) v1.CashXiaozhiMcpGuideVideoItem {
	return v1.CashXiaozhiMcpGuideVideoItem{
		Slot:        it.Slot,
		Title:       it.Title,
		ObjectKey:   it.ObjectKey,
		VideoCdnUrl: it.VideoCdnURL,
		CoverCdnUrl: it.CoverCdnURL,
		Configured:  it.Configured,
		UpdatedAt:   it.UpdatedAt,
	}
}

// GuideVideos GET /cash/app/api/xiaozhi-mcp/guide-videos — 仅已配置槽（计入 usage）。
func (c *CashXiaozhiMcpController) GuideVideos(ctx context.Context, _ *v1.CashAppXiaozhiMcpGuideVideosReq) (*v1.CashAppXiaozhiMcpGuideVideosRes, error) {
	if _, err := cashWxIDFromHeader(ctx); err != nil {
		return nil, err
	}
	list, err := cash.ListXiaozhiGuideVideosApp(ctx)
	if err != nil {
		return nil, err
	}
	res := &v1.CashAppXiaozhiMcpGuideVideosRes{List: make([]v1.CashXiaozhiMcpGuideVideoItem, 0, len(list))}
	for _, it := range list {
		res.List = append(res.List, mapXiaozhiGuideItem(it))
	}
	return res, nil
}

// AdminGuideVideosGet GET /cash/admin/api/xiaozhi-mcp/guide-videos
func (c *CashXiaozhiMcpController) AdminGuideVideosGet(ctx context.Context, _ *v1.CashAdminXiaozhiMcpGuideVideosGetReq) (*v1.CashAdminXiaozhiMcpGuideVideosGetRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	list, err := cash.ListXiaozhiGuideVideosAdmin(ctx)
	if err != nil {
		return nil, err
	}
	res := &v1.CashAdminXiaozhiMcpGuideVideosGetRes{List: make([]v1.CashXiaozhiMcpGuideVideoItem, 0, len(list))}
	for _, it := range list {
		res.List = append(res.List, mapXiaozhiGuideItem(it))
	}
	return res, nil
}

// AdminGuideVideoPut PUT /cash/admin/api/xiaozhi-mcp/guide-videos/{slot}
func (c *CashXiaozhiMcpController) AdminGuideVideoPut(ctx context.Context, req *v1.CashAdminXiaozhiMcpGuideVideoPutReq) (*v1.CashAdminXiaozhiMcpGuideVideoPutRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	if err := cash.UpdateXiaozhiGuideVideoTitle(ctx, req.Slot, req.Title); err != nil {
		return nil, err
	}
	return &v1.CashAdminXiaozhiMcpGuideVideoPutRes{}, nil
}

// AdminGuideVideoUpload POST /cash/admin/api/xiaozhi-mcp/guide-videos/{slot}/upload — multipart file。
func (c *CashXiaozhiMcpController) AdminGuideVideoUpload(ctx context.Context, req *v1.CashAdminXiaozhiMcpGuideVideoUploadReq) (*v1.CashAdminXiaozhiMcpGuideVideoUploadRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "无效请求")
	}
	// 与 UCG 视频上限对齐（25MB + 表单余量）。
	if err := r.ParseMultipartForm((25 << 20) + (1 << 20)); err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "multipart 解析失败")
	}
	file, hdr, err := r.Request.FormFile("file")
	if err != nil || file == nil {
		file, hdr, err = r.Request.FormFile("video")
	}
	if err != nil || file == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 file 或 video 字段")
	}
	defer file.Close()
	filename := ""
	var size int64
	if hdr != nil {
		filename = hdr.Filename
		size = hdr.Size
	}
	slot := strings.TrimSpace(req.Slot)
	out, err := cash.UploadXiaozhiGuideVideo(ctx, slot, filename, file, size)
	if err != nil {
		return nil, err
	}
	return &v1.CashAdminXiaozhiMcpGuideVideoUploadRes{
		Slot:        out.Slot,
		Title:       out.Title,
		ObjectKey:   out.ObjectKey,
		VideoCdnUrl: out.VideoCdnURL,
		CoverCdnUrl: out.CoverCdnURL,
		UpdatedAt:   out.UpdatedAt,
	}, nil
}

// AdminGuideVideoDelete DELETE /cash/admin/api/xiaozhi-mcp/guide-videos/{slot}
func (c *CashXiaozhiMcpController) AdminGuideVideoDelete(ctx context.Context, req *v1.CashAdminXiaozhiMcpGuideVideoDeleteReq) (*v1.CashAdminXiaozhiMcpGuideVideoDeleteRes, error) {
	if err := requireCashAdmin(ctx); err != nil {
		return nil, err
	}
	if err := cash.ClearXiaozhiGuideVideo(ctx, req.Slot); err != nil {
		return nil, err
	}
	return &v1.CashAdminXiaozhiMcpGuideVideoDeleteRes{}, nil
}
