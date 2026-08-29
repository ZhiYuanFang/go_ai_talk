package cash

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// —— Admin：功能定义 / SKU ——

// AdminUpdateFeatureDef 更新已有功能定义（禁止新建任意 featureId；编号与客户端约定）。
//
// Args: defaultAllowedCount 预测类默认免费开通条数（其它功能可 0）。
func AdminUpdateFeatureDef(ctx context.Context, featureID, title, desc, unlockMethods string, durationDays, status, sortOrder, defaultAllowedCount int) error {
	featureID = strings.TrimSpace(featureID)
	if featureID == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "featureId 不能为空")
	}
	if defaultAllowedCount < 0 {
		defaultAllowedCount = 0
	}
	now := time.Now().Unix()
	r, err := g.DB().Model("feature_def").Ctx(ctx).Where("feature_id", featureID).One()
	if err != nil {
		return err
	}
	if r.IsEmpty() {
		return gerror.NewCode(gcode.CodeInvalidParameter, "功能编号不存在（须与客户端约定，禁止管理页新建）")
	}
	_, err = g.DB().Model("feature_def").Ctx(ctx).Where("feature_id", featureID).Data(g.Map{
		"title": title, "description": desc, "unlock_methods": unlockMethods,
		"duration_days": durationDays, "default_allowed_count": defaultAllowedCount,
		"status": status, "sort_order": sortOrder, "updated_at": now,
	}).Update()
	invalidateFeatureDefCache(ctx)
	return err
}

// AdminUpsertFeatureDef 兼容旧名：仅更新已存在定义。
func AdminUpsertFeatureDef(ctx context.Context, featureID, title, desc, unlockMethods string, durationDays, status, sortOrder int) error {
	return AdminUpdateFeatureDef(ctx, featureID, title, desc, unlockMethods, durationDays, status, sortOrder, 0)
}

// AdminListFeatureDefs 管理端功能列表（含停用）。
func AdminListFeatureDefs(ctx context.Context) ([]FeatureDefRow, error) {
	var raw []featureDefDB
	err := g.DB().Model("feature_def").Ctx(ctx).
		Fields("feature_id,title,description,unlock_methods,duration_days,default_allowed_count,status,sort_order").
		OrderAsc("sort_order").Scan(&raw)
	if err != nil {
		return nil, err
	}
	out := make([]FeatureDefRow, 0, len(raw))
	for _, r := range raw {
		out = append(out, FeatureDefRow{
			FeatureId: r.FeatureId, Title: r.Title, Description: r.Description,
			UnlockMethods: r.UnlockMethods, DurationDays: r.DurationDays,
			DefaultAllowedCount: r.DefaultAllowedCount,
			Status: r.Status, SortOrder: r.SortOrder,
		})
	}
	return out, nil
}

// genFeatureProductCode 生成全局唯一商品编码。
func genFeatureProductCode(featureID string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	fid := strings.TrimSpace(featureID)
	if len(fid) > 24 {
		fid = fid[:24]
	}
	if fid == "" {
		fid = "feat"
	}
	return fmt.Sprintf("fp_%s_%d_%s", fid, time.Now().Unix(), hex.EncodeToString(b))
}

// AdminUpsertFeatureProduct 创建或更新功能 SKU。
//
// 业务：新建时 productCode 可空（服务端自动生成）；更新必须带已有编码且不得改所属功能外的主键语义。
// 所属 featureId MUST 已存在于 feature_def。
func AdminUpsertFeatureProduct(ctx context.Context, p *FeatureProduct) error {
	if p == nil {
		return gerror.NewCode(gcode.CodeInvalidParameter, "参数无效")
	}
	p.FeatureId = strings.TrimSpace(p.FeatureId)
	p.ProductCode = strings.TrimSpace(p.ProductCode)
	if p.FeatureId == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "featureId 不能为空")
	}
	def, err := g.DB().Model("feature_def").Ctx(ctx).Where("feature_id", p.FeatureId).One()
	if err != nil {
		return err
	}
	if def.IsEmpty() {
		return gerror.NewCode(gcode.CodeInvalidParameter, "所属功能不存在")
	}
	now := time.Now().Unix()
	if p.GrantQuantity <= 0 {
		p.GrantQuantity = 1
	}
	if p.GrantKind == "" {
		p.GrantKind = GrantKindEntitlement
	}

	if p.ProductCode == "" {
		// 新建：自动生成编码。
		for i := 0; i < 5; i++ {
			code := genFeatureProductCode(p.FeatureId)
			exist, e := g.DB().Model("feature_product").Ctx(ctx).Where("product_code", code).One()
			if e != nil {
				return e
			}
			if exist.IsEmpty() {
				p.ProductCode = code
				break
			}
		}
		if p.ProductCode == "" {
			return gerror.NewCode(gcode.CodeInternalError, "生成商品编码失败")
		}
		_, err = g.DB().Model("feature_product").Ctx(ctx).Data(g.Map{
			"product_code": p.ProductCode, "feature_id": p.FeatureId, "grant_kind": p.GrantKind,
			"grant_quantity": p.GrantQuantity, "price_fen": p.PriceFen, "original_price_fen": p.OriginalPriceFen,
			"duration_days": p.DurationDays, "apple_product_id": p.AppleProductId, "status": p.Status, "updated_at": now,
		}).Insert()
		invalidateFeatureDefCache(ctx)
		return err
	}

	r, err := g.DB().Model("feature_product").Ctx(ctx).Where("product_code", p.ProductCode).One()
	if err != nil {
		return err
	}
	data := g.Map{
		"feature_id": p.FeatureId, "grant_kind": p.GrantKind,
		"grant_quantity": p.GrantQuantity, "price_fen": p.PriceFen, "original_price_fen": p.OriginalPriceFen,
		"duration_days": p.DurationDays, "apple_product_id": p.AppleProductId, "status": p.Status, "updated_at": now,
	}
	if r.IsEmpty() {
		// 运维指定编码新建。
		data["product_code"] = p.ProductCode
		_, err = g.DB().Model("feature_product").Ctx(ctx).Data(data).Insert()
	} else {
		_, err = g.DB().Model("feature_product").Ctx(ctx).Where("product_code", p.ProductCode).Data(data).Update()
	}
	invalidateFeatureDefCache(ctx)
	return err
}

// AdminListFeatureProducts 功能 SKU 列表。
func AdminListFeatureProducts(ctx context.Context) ([]FeatureProduct, error) {
	var raw []featureProductDB
	err := g.DB().Model("feature_product").Ctx(ctx).Scan(&raw)
	if err != nil {
		return nil, err
	}
	out := make([]FeatureProduct, 0, len(raw))
	for _, r := range raw {
		out = append(out, *mapFeatureProduct(r))
	}
	return out, nil
}

// —— Admin：邀请码 ——

// InviteCodeAdminRow 邀请码管理行。
type InviteCodeAdminRow struct {
	Code              string   `json:"code"`
	OwnerWxId         int64    `json:"ownerWxId"`
	ExpiresAt         int64    `json:"expiresAt"`
	MaxRedemptions    int      `json:"maxRedemptions"`
	RedeemedCount     int      `json:"redeemedCount"`
	GrantDurationDays int      `json:"grantDurationDays"`
	Status            int      `json:"status"`
	FeatureIds        []string `json:"featureIds"`
	CreatedAt         int64    `json:"createdAt"`
}

// InviteRedemptionRow 兑换明细。
type InviteRedemptionRow struct {
	Code         string `json:"code"`
	OwnerWxId    int64  `json:"ownerWxId"`
	RedeemerWxId int64  `json:"redeemerWxId"`
	DeviceNo     string `json:"deviceNo"`
	FeatureId    string `json:"featureId"`
	RedeemedAt   int64  `json:"redeemedAt"`
}

// AdminCreateInviteCode 创建邀请码（一期一 owner 一码）。
func AdminCreateInviteCode(ctx context.Context, ownerWxID int64, code string, expiresAt int64, maxRedeem, grantDays, status int, featureIDs []string) (string, error) {
	if ownerWxID <= 0 {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "ownerWxId 无效")
	}
	if status == 0 {
		status = 1
	}
	code = strings.TrimSpace(code)
	if code == "" {
		b := make([]byte, 8)
		_, _ = rand.Read(b)
		code = strings.ToUpper(hex.EncodeToString(b))
	}
	exist, err := g.DB().Model("feature_invite_code").Ctx(ctx).Where("owner_wx_id", ownerWxID).One()
	if err != nil {
		return "", err
	}
	if !exist.IsEmpty() {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "该用户已有邀请码（一期一人一码）")
	}
	now := time.Now().Unix()
	_, err = g.DB().Model("feature_invite_code").Ctx(ctx).Data(g.Map{
		"code": code, "owner_wx_id": ownerWxID, "expires_at": expiresAt,
		"max_redemptions": maxRedeem, "redeemed_count": 0, "grant_duration_days": grantDays,
		"status": status, "created_at": now, "updated_at": now,
	}).Insert()
	if err != nil {
		return "", err
	}
	for _, fid := range featureIDs {
		fid = strings.TrimSpace(fid)
		if fid == "" {
			continue
		}
		_, _ = g.DB().Model("feature_invite_code_feature").Ctx(ctx).Data(g.Map{
			"code": code, "feature_id": fid, "grant_quantity": 1,
		}).Insert()
	}
	return code, nil
}

// AdminListInviteCodes 邀请码列表。
func AdminListInviteCodes(ctx context.Context) ([]InviteCodeAdminRow, error) {
	type row struct {
		Code              string `json:"code"`
		OwnerWxId         int64  `json:"owner_wx_id"`
		ExpiresAt         int64  `json:"expires_at"`
		MaxRedemptions    int    `json:"max_redemptions"`
		RedeemedCount     int    `json:"redeemed_count"`
		GrantDurationDays int    `json:"grant_duration_days"`
		Status            int    `json:"status"`
		CreatedAt         int64  `json:"created_at"`
	}
	var rows []row
	err := g.DB().Model("feature_invite_code").Ctx(ctx).OrderDesc("created_at").Scan(&rows)
	if err != nil {
		return nil, err
	}
	out := make([]InviteCodeAdminRow, 0, len(rows))
	for _, r := range rows {
		item := InviteCodeAdminRow{
			Code: r.Code, OwnerWxId: r.OwnerWxId, ExpiresAt: r.ExpiresAt,
			MaxRedemptions: r.MaxRedemptions, RedeemedCount: r.RedeemedCount,
			GrantDurationDays: r.GrantDurationDays, Status: r.Status, CreatedAt: r.CreatedAt,
		}
		var fids []struct {
			FeatureId string `json:"feature_id"`
		}
		_ = g.DB().Model("feature_invite_code_feature").Ctx(ctx).Where("code", r.Code).Scan(&fids)
		for _, f := range fids {
			item.FeatureIds = append(item.FeatureIds, f.FeatureId)
		}
		out = append(out, item)
	}
	return out, nil
}

// AdminUpdateInviteCodeStatus 停用/启用邀请码。
func AdminUpdateInviteCodeStatus(ctx context.Context, code string, status int) error {
	code = strings.TrimSpace(code)
	_, err := g.DB().Model("feature_invite_code").Ctx(ctx).Where("code", code).Data(g.Map{
		"status": status, "updated_at": time.Now().Unix(),
	}).Update()
	return err
}

// AdminListInviteRedemptions 按码查兑换明细。
func AdminListInviteRedemptions(ctx context.Context, code string, limit int) ([]InviteRedemptionRow, error) {
	code = strings.TrimSpace(code)
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	type row struct {
		Code         string `json:"code"`
		OwnerWxId    int64  `json:"owner_wx_id"`
		RedeemerWxId int64  `json:"redeemer_wx_id"`
		DeviceNo     string `json:"device_no"`
		FeatureId    string `json:"feature_id"`
		RedeemedAt   int64  `json:"redeemed_at"`
	}
	m := g.DB().Model("feature_invite_redemption").Ctx(ctx).OrderDesc("redeemed_at").Limit(limit)
	if code != "" {
		m = m.Where("code", code)
	}
	var rows []row
	if err := m.Scan(&rows); err != nil {
		return nil, err
	}
	out := make([]InviteRedemptionRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, InviteRedemptionRow{
			Code: r.Code, OwnerWxId: r.OwnerWxId, RedeemerWxId: r.RedeemerWxId,
			DeviceNo: r.DeviceNo, FeatureId: r.FeatureId, RedeemedAt: r.RedeemedAt,
		})
	}
	return out, nil
}
