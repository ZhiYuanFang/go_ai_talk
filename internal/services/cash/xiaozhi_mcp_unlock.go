package cash

// xiaozhi_mcp_unlock.go：小智 MCP 永久能力（SKU / 权益 / 建单 / 履约 / 手工授）。
//
// 业务：
//   - wx 维一次开通永久「可添加小智绑定」能力；
//   - SKU 专用表，不进 feature_product / 开通功能管理；
//   - 订单复用 feature_order，履约不走 ActivateFeature；
//   - Admin 手工授写 0 元 paid 订单 + 永久权益。

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// XiaozhiMcpProduct 小智 MCP SKU 视图。
type XiaozhiMcpProduct struct {
	ProductCode      string `json:"productCode"`
	Title            string `json:"title"`
	PriceFen         int    `json:"priceFen"`
	OriginalPriceFen int    `json:"originalPriceFen"`
	AppleProductId   string `json:"appleProductId"`
	Status           int    `json:"status"`
	UpdatedAt        int64  `json:"updatedAt"`
}

// XiaozhiMcpUnlockView App 开通态 + 可售 SKU。
type XiaozhiMcpUnlockView struct {
	Unlocked       bool               `json:"unlocked"`
	TrialAvailable bool               `json:"trialAvailable"`       // 从未 claim 且当前未开通
	ExpiresAt      int64              `json:"expiresAt,omitempty"` // 0=永久（已开通时）；试用为截止秒
	Product        *XiaozhiMcpProduct `json:"product,omitempty"`   // 上架时可购；已开通也可带回便于展示
}

// GetXiaozhiMcpProduct 按编码读 SKU（含下架；空行业务 NotFound）。
func GetXiaozhiMcpProduct(ctx context.Context, productCode string) (*XiaozhiMcpProduct, error) {
	productCode = strings.TrimSpace(productCode)
	if productCode == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "productCode 不能为空")
	}
	one, err := g.DB().Model("xiaozhi_mcp_product").Ctx(ctx).Where("product_code", productCode).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotFound, "小智 MCP 商品不存在")
	}
	return mapXiaozhiMcpProduct(one), nil
}

// GetActiveXiaozhiMcpProduct 仅上架 SKU。
func GetActiveXiaozhiMcpProduct(ctx context.Context, productCode string) (*XiaozhiMcpProduct, error) {
	p, err := GetXiaozhiMcpProduct(ctx, productCode)
	if err != nil {
		return nil, err
	}
	if p.Status != XiaozhiMcpProductOn {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "小智 MCP 商品已下架")
	}
	return p, nil
}

// GetXiaozhiMcpProductByAppleID 按 Apple productId 查找上架 MCP SKU；未命中返回 (nil, nil)。
func GetXiaozhiMcpProductByAppleID(ctx context.Context, applePID string) (*XiaozhiMcpProduct, error) {
	applePID = strings.TrimSpace(applePID)
	if applePID == "" {
		return nil, nil
	}
	one, err := g.DB().Model("xiaozhi_mcp_product").Ctx(ctx).
		Where("apple_product_id", applePID).Where("status", XiaozhiMcpProductOn).Limit(1).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return nil, nil
	}
	return mapXiaozhiMcpProduct(one), nil
}

func mapXiaozhiMcpProduct(one gdb.Record) *XiaozhiMcpProduct {
	return &XiaozhiMcpProduct{
		ProductCode:      one["product_code"].String(),
		Title:            one["title"].String(),
		PriceFen:         one["price_fen"].Int(),
		OriginalPriceFen: one["original_price_fen"].Int(),
		AppleProductId:   one["apple_product_id"].String(),
		Status:           one["status"].Int(),
		UpdatedAt:        one["updated_at"].Int64(),
	}
}

// AdminGetXiaozhiMcpProduct 管理端读取种子 SKU。
func AdminGetXiaozhiMcpProduct(ctx context.Context) (*XiaozhiMcpProduct, error) {
	return GetXiaozhiMcpProduct(ctx, XiaozhiMcpProductCodePerm)
}

// AdminUpdateXiaozhiMcpProductInput 更新 SKU（不可改 product_code）。
type AdminUpdateXiaozhiMcpProductInput struct {
	Title            string
	PriceFen         int
	OriginalPriceFen int
	AppleProductId   string
	Status           int
}

// AdminUpdateXiaozhiMcpProduct 更新种子 SKU 价格/Apple ID/上下架。
func AdminUpdateXiaozhiMcpProduct(ctx context.Context, in AdminUpdateXiaozhiMcpProductInput) error {
	if in.PriceFen < 0 || in.OriginalPriceFen < 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "价格不能为负")
	}
	if in.Status != XiaozhiMcpProductOn && in.Status != XiaozhiMcpProductOff {
		return gerror.NewCode(gcode.CodeInvalidParameter, "status 须为 0 或 1")
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = "小智 MCP 永久开通"
	}
	now := time.Now().Unix()
	res, err := g.DB().Model("xiaozhi_mcp_product").Ctx(ctx).
		Where("product_code", XiaozhiMcpProductCodePerm).
		Data(g.Map{
			"title":              title,
			"price_fen":          in.PriceFen,
			"original_price_fen": in.OriginalPriceFen,
			"apple_product_id":   strings.TrimSpace(in.AppleProductId),
			"status":             in.Status,
			"updated_at":         now,
		}).Update()
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		// 行不存在时补插（EnsureSchema 未跑到的极端情况）。
		_, err = g.DB().Model("xiaozhi_mcp_product").Ctx(ctx).Data(g.Map{
			"product_code":       XiaozhiMcpProductCodePerm,
			"title":              title,
			"price_fen":          in.PriceFen,
			"original_price_fen": in.OriginalPriceFen,
			"apple_product_id":   strings.TrimSpace(in.AppleProductId),
			"status":             in.Status,
			"updated_at":         now,
		}).Insert()
		return err
	}
	return nil
}

// HasActiveXiaozhiMcpEntitlement 账号是否具备有效能力（永久或未过期试用）。
func HasActiveXiaozhiMcpEntitlement(ctx context.Context, wxID int64) (bool, error) {
	active, _, err := loadXiaozhiMcpEntitlementActive(ctx, wxID)
	return active, err
}

// HasPermanentXiaozhiMcpEntitlement 是否已永久开通（expires_at=0 且有效）。
func HasPermanentXiaozhiMcpEntitlement(ctx context.Context, wxID int64) (bool, error) {
	active, exp, err := loadXiaozhiMcpEntitlementActive(ctx, wxID)
	if err != nil || !active {
		return false, err
	}
	return exp == 0, nil
}

// loadXiaozhiMcpEntitlementActive 返回是否有效及 expires_at（0=永久）。
func loadXiaozhiMcpEntitlementActive(ctx context.Context, wxID int64) (active bool, expiresAt int64, err error) {
	if wxID <= 0 {
		return false, 0, nil
	}
	one, err := g.DB().Model("xiaozhi_mcp_entitlement").Ctx(ctx).
		Fields("status,expires_at").Where("wx_id", wxID).One()
	if err != nil {
		return false, 0, err
	}
	if one.IsEmpty() || one["status"].Int() != XiaozhiMcpEntitlementActive {
		return false, 0, nil
	}
	exp := one["expires_at"].Int64()
	if exp == 0 {
		return true, 0, nil
	}
	if time.Now().Unix() < exp {
		return true, exp, nil
	}
	return false, exp, nil
}

// GetXiaozhiMcpUnlockView App 开通态 + 可售信息。
func GetXiaozhiMcpUnlockView(ctx context.Context, wxID int64) (*XiaozhiMcpUnlockView, error) {
	unlocked, exp, err := loadXiaozhiMcpEntitlementActive(ctx, wxID)
	if err != nil {
		return nil, err
	}
	trialAvail, tErr := IsXiaozhiMcpTrialUnused(ctx, wxID)
	if tErr != nil {
		return nil, tErr
	}
	out := &XiaozhiMcpUnlockView{
		Unlocked:       unlocked,
		TrialAvailable: trialAvail && !unlocked,
	}
	if unlocked {
		out.ExpiresAt = exp // 0=永久
	}
	prod, pErr := GetXiaozhiMcpProduct(ctx, XiaozhiMcpProductCodePerm)
	if pErr != nil {
		glog.Warningf(ctx, "[cash] xiaozhi-mcp product missing err=%v", pErr)
		return out, nil
	}
	if unlocked || prod.Status == XiaozhiMcpProductOn {
		out.Product = prod
	}
	return out, nil
}

// GrantXiaozhiMcpPermanent 写入/恢复 wx 永久能力（expires_at=0）。
//
// Args: unlockMethod payment|admin；channelRef 订单号。
// Side Effects: upsert xiaozhi_mcp_entitlement。
func GrantXiaozhiMcpPermanent(ctx context.Context, wxID int64, unlockMethod, channelRef string) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	unlockMethod = strings.TrimSpace(unlockMethod)
	if unlockMethod != XiaozhiMcpUnlockPayment && unlockMethod != XiaozhiMcpUnlockAdmin {
		return gerror.NewCode(gcode.CodeInvalidParameter, "unlockMethod 无效")
	}
	channelRef = strings.TrimSpace(channelRef)
	now := time.Now().Unix()
	one, err := g.DB().Model("xiaozhi_mcp_entitlement").Ctx(ctx).Where("wx_id", wxID).One()
	if err != nil {
		return err
	}
	data := g.Map{
		"status":        XiaozhiMcpEntitlementActive,
		"unlock_method": unlockMethod,
		"channel_ref":   channelRef,
		"expires_at":    0,
		"revoked_at":    0,
		"updated_at":    now,
	}
	if one.IsEmpty() {
		data["wx_id"] = wxID
		data["unlocked_at"] = now
		_, err = g.DB().Model("xiaozhi_mcp_entitlement").Ctx(ctx).Data(data).Insert()
		return err
	}
	// 支付/手工授可覆盖试用为永久。
	data["unlocked_at"] = now
	_, err = g.DB().Model("xiaozhi_mcp_entitlement").Ctx(ctx).Where("wx_id", wxID).Data(data).Update()
	return err
}

// RevokeXiaozhiMcpEntitlement 撤销账号能力（退款/手工撤共用）。
func RevokeXiaozhiMcpEntitlement(ctx context.Context, wxID int64) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	now := time.Now().Unix()
	_, err := g.DB().Model("xiaozhi_mcp_entitlement").Ctx(ctx).
		Where("wx_id", wxID).Where("status", XiaozhiMcpEntitlementActive).
		Data(g.Map{
			"status":     XiaozhiMcpEntitlementRevoked,
			"revoked_at": now,
			"updated_at": now,
		}).Update()
	return err
}

// CreateXiaozhiMcpOrder 创建小智 MCP 永久买断订单（写入 feature_order）。
//
// 已永久开通则拒绝；试用中允许建单以升级为永久。deviceNo 可空串。
func CreateXiaozhiMcpOrder(ctx context.Context, deviceNo string, wxID int64, channel string) (*CreateOrderResult, error) {
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少登录用户")
	}
	channel = strings.TrimSpace(channel)
	if channel != ChannelAlipay && channel != ChannelAppleIAP {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "channel 必须为 alipay 或 apple_iap")
	}
	perm, err := HasPermanentXiaozhiMcpEntitlement(ctx, wxID)
	if err != nil {
		return nil, err
	}
	if perm {
		return nil, gerror.NewCode(gcode.CodeInvalidOperation, "已开通小智 MCP，无需重复购买")
	}
	prod, err := GetActiveXiaozhiMcpProduct(ctx, XiaozhiMcpProductCodePerm)
	if err != nil {
		return nil, err
	}
	orderNo, err := newOrderNo()
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	row := g.Map{
		"order_no":     orderNo,
		"device_no":    strings.TrimSpace(deviceNo),
		"wx_id":        wxID,
		"product_code": prod.ProductCode,
		"channel":      channel,
		"amount_fen":   prod.PriceFen,
		"currency":     "CNY",
		"status":       OrderCreated,
		"created_at":   now,
	}
	var appToken string
	if channel == ChannelAppleIAP {
		appToken, err = newAppAccountToken()
		if err != nil {
			return nil, err
		}
		row["app_account_token"] = appToken
	}
	if _, err = g.DB().Model("feature_order").Ctx(ctx).Data(row).Insert(); err != nil {
		return nil, err
	}
	out := &CreateOrderResult{
		OrderNo: orderNo, ProductCode: prod.ProductCode, Channel: channel, AmountFen: prod.PriceFen,
	}
	if channel == ChannelAppleIAP {
		out.AppleProductId = prod.AppleProductId
		out.AppAccountToken = appToken
		return out, nil
	}
	vipShape := &Product{
		ProductCode: prod.ProductCode, Title: prod.Title, PriceFen: prod.PriceFen,
		DurationDays: 0, AppleProductId: prod.AppleProductId,
	}
	orderStr, tip, aErr := BuildAlipayAppPayOrderStr(ctx, orderNo, vipShape)
	if aErr != nil {
		return nil, aErr
	}
	out.AlipayOrderStr = orderStr
	out.PayTip = tip
	return out, nil
}

// FulfillXiaozhiMcpPaid MCP 订单履约：标 paid + 永久授（幂等）。
func FulfillXiaozhiMcpPaid(ctx context.Context, orderNo, channel, channelTxnID string, amountFen int) error {
	orderNo = strings.TrimSpace(orderNo)
	if orderNo == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "orderNo 不能为空")
	}
	if channelTxnID != "" {
		if existed, err := loadFeatureOrderByChannelTxn(ctx, channel, channelTxnID); err != nil {
			return err
		} else if existed != nil && existed.Status == OrderPaid {
			return nil
		}
	}
	order, err := loadFeatureOrderByNo(ctx, orderNo)
	if err != nil {
		return err
	}
	if !IsXiaozhiMcpProductCode(order.ProductCode) {
		return gerror.NewCode(gcode.CodeInvalidParameter, "非小智 MCP 订单")
	}
	if order.Channel != channel && channel != "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "支付渠道与订单不一致")
	}
	if amountFen > 0 && order.AmountFen != amountFen {
		return gerror.NewCode(gcode.CodeInvalidParameter, "支付金额与订单不一致")
	}
	if order.Status == OrderPaid {
		return nil
	}
	if order.Status != OrderCreated {
		return gerror.NewCode(gcode.CodeInvalidOperation, "订单状态不可支付")
	}
	now := time.Now().Unix()
	res, err := g.DB().Model("feature_order").Ctx(ctx).
		Where("id", order.Id).Where("status", OrderCreated).
		Data(g.Map{
			"status":         OrderPaid,
			"channel_txn_id": channelTxnID,
			"paid_at":        now,
		}).Update()
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil
	}
	if order.WxId <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "订单缺少 wxId")
	}
	return GrantXiaozhiMcpPermanent(ctx, order.WxId, XiaozhiMcpUnlockPayment, order.OrderNo)
}

// RevokeXiaozhiMcpGrantForOrder 支付退款：撤销订单对应 wx 的 MCP 能力。
func RevokeXiaozhiMcpGrantForOrder(ctx context.Context, order *FeatureOrder) error {
	if order == nil || !IsXiaozhiMcpProductCode(order.ProductCode) {
		return nil
	}
	if order.WxId <= 0 {
		return nil
	}
	if err := RevokeXiaozhiMcpEntitlement(ctx, order.WxId); err != nil {
		return err
	}
	glog.Infof(ctx, "[cash] xiaozhi-mcp refund revoke wxId=%d orderNo=%s", order.WxId, order.OrderNo)
	return nil
}

// AdminGrantXiaozhiMcp 手工授永久能力（0 元 admin feature_order + 权益）。
// 试用中也可授，升级为永久。
func AdminGrantXiaozhiMcp(ctx context.Context, wxID int64, reason string) (orderNo string, err error) {
	if wxID <= 0 {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	reason, err = normalizeGrantReason(reason)
	if err != nil {
		return "", err
	}
	perm, err := HasPermanentXiaozhiMcpEntitlement(ctx, wxID)
	if err != nil {
		return "", err
	}
	if perm {
		return "", gerror.NewCode(gcode.CodeInvalidOperation, "该账号已永久开通小智 MCP")
	}
	orderNo, err = newAdminOrderNo("XMP")
	if err != nil {
		return "", err
	}
	now := time.Now().Unix()
	txnID := "admin-" + orderNo
	_, err = g.DB().Model("feature_order").Ctx(ctx).Data(g.Map{
		"order_no":       orderNo,
		"device_no":      "",
		"wx_id":          wxID,
		"product_code":   XiaozhiMcpProductCodePerm,
		"channel":        ChannelAdmin,
		"amount_fen":     0,
		"currency":       "CNY",
		"status":         OrderPaid,
		"channel_txn_id": txnID,
		"grant_reason":   reason,
		"created_at":     now,
		"paid_at":        now,
	}).Insert()
	if err != nil {
		return "", err
	}
	if err = GrantXiaozhiMcpPermanent(ctx, wxID, XiaozhiMcpUnlockAdmin, orderNo); err != nil {
		return "", err
	}
	return orderNo, nil
}

// AdminRevokeXiaozhiMcp 撤销最近一笔仍有效的手工授。
func AdminRevokeXiaozhiMcp(ctx context.Context, wxID int64) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	one, err := g.DB().Model("xiaozhi_mcp_entitlement").Ctx(ctx).Where("wx_id", wxID).One()
	if err != nil {
		return err
	}
	if one.IsEmpty() || one["status"].Int() != XiaozhiMcpEntitlementActive {
		return gerror.NewCode(gcode.CodeInvalidOperation, "该账号无有效小智 MCP 开通")
	}
	if one["unlock_method"].String() != XiaozhiMcpUnlockAdmin {
		return gerror.NewCode(gcode.CodeInvalidOperation, "仅最近一笔手工授可撤销")
	}
	channelRef := strings.TrimSpace(one["channel_ref"].String())
	if channelRef != "" {
		_, _ = g.DB().Model("feature_order").Ctx(ctx).
			Where("order_no", channelRef).Where("channel", ChannelAdmin).
			Where("product_code", XiaozhiMcpProductCodePerm).
			Data(g.Map{"status": OrderRevoked}).Update()
	}
	return RevokeXiaozhiMcpEntitlement(ctx, wxID)
}
