package cash

import (
	"context"
	"strings"
	"time"

	"hello/internal/shared/featurelogo"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// —— Admin：功能定义 / SKU ——

// AdminUpdateFeatureDef 更新已有功能定义（禁止新建任意 featureId；编号与客户端约定）。
//
// Args: durationDays、邀请天数、广告天数均须≥1，不支持永久。
//
//	defaultAllowedCount 预测类默认免费开通条数（其它功能可 0）。
//	activationSubject 可选：nil 保持原值；非 nil 须为 device|user；预测类禁止改为 user。
//	logoObjectKey 可选：非空则更新 logo（存 objectKey）；空串表示保留原 logo。
//	color 主色 hex；空串允许写入（防御）；非法格式拒绝。
func AdminUpdateFeatureDef(ctx context.Context, featureID, title, desc, unlockMethods string, durationDays, inviteDurationDays, adDurationDays, status, sortOrder, defaultAllowedCount int, activationSubject *string, logoObjectKey, color string) error {
	featureID = strings.TrimSpace(featureID)
	if featureID == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "featureId 不能为空")
	}
	if featureID == FeatureIDPredictionUnlock {
		glog.Warningf(ctx, "[cash] prediction_unlock 已下线，拒绝改定义 featureId=%s", featureID)
		return gerror.NewCode(gcode.CodeInvalidOperation, "预测事项开通数量已下线")
	}
	if durationDays < 1 || inviteDurationDays < 1 || adDurationDays < 1 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "授予天数须≥1，不支持永久")
	}
	if defaultAllowedCount < 0 {
		defaultAllowedCount = 0
	}
	color = strings.TrimSpace(color)
	if err := ValidateFeatureColor(color); err != nil {
		return err
	}
	now := time.Now().Unix()
	r, err := g.DB().Model("feature_def").Ctx(ctx).Where("feature_id", featureID).One()
	if err != nil {
		return err
	}
	if r.IsEmpty() {
		return gerror.NewCode(gcode.CodeInvalidParameter, "功能编号不存在（须与客户端约定，禁止管理页新建）")
	}
	data := g.Map{
		"title": title, "description": desc, "unlock_methods": unlockMethods,
		"duration_days":        durationDays,
		"invite_duration_days": inviteDurationDays, "ad_duration_days": adDurationDays,
		"default_allowed_count": defaultAllowedCount,
		"color":                 color,
		"status":                status, "sort_order": sortOrder, "updated_at": now,
	}
	// logo：非空才更新，避免未传文件时清空已有图
	if key := featurelogo.StoredObjectKey(ctx, logoObjectKey); key != "" {
		data["logo"] = key
	}
	if activationSubject != nil {
		raw := strings.TrimSpace(*activationSubject)
		if raw != "" && raw != ActivationSubjectDevice && raw != ActivationSubjectUser {
			return gerror.NewCode(gcode.CodeInvalidParameter, "activationSubject 须为 device 或 user")
		}
		subj := NormalizeActivationSubject(raw)
		data["activation_subject"] = subj
	}
	_, err = g.DB().Model("feature_def").Ctx(ctx).Where("feature_id", featureID).Data(data).Update()
	invalidateFeatureDefCache(ctx)
	return err
}

// AdminUpsertFeatureDef 兼容旧名：仅更新已存在定义（邀请/广告天数与 durationDays 双写）。
func AdminUpsertFeatureDef(ctx context.Context, featureID, title, desc, unlockMethods string, durationDays, status, sortOrder int) error {
	return AdminUpdateFeatureDef(ctx, featureID, title, desc, unlockMethods, durationDays, durationDays, durationDays, status, sortOrder, 0, nil, "", "")
}

// AdminListFeatureDefs 管理端功能列表（含停用）；Logo 字段为 CDN URL。
func AdminListFeatureDefs(ctx context.Context) ([]FeatureDefRow, error) {
	var raw []featureDefDB
	err := g.DB().Model("feature_def").Ctx(ctx).
		Fields("feature_id,title,description,unlock_methods,duration_days,invite_duration_days,ad_duration_days,default_allowed_count,activation_subject,logo,color,status,sort_order").
		OrderAsc("sort_order").Scan(&raw)
	if err != nil {
		return nil, err
	}
	out := make([]FeatureDefRow, 0, len(raw))
	for _, r := range raw {
		out = append(out, FeatureDefRow{
			FeatureId: r.FeatureId, Title: r.Title, Description: r.Description,
			UnlockMethods: r.UnlockMethods, DurationDays: r.DurationDays,
			InviteDurationDays: r.InviteDurationDays, AdDurationDays: r.AdDurationDays,
			DefaultAllowedCount: r.DefaultAllowedCount,
			ActivationSubject:   NormalizeActivationSubject(r.ActivationSubject),
			Logo:                featurelogo.CdnURL(ctx, r.Logo), Color: strings.TrimSpace(r.Color),
			Status: r.Status, SortOrder: r.SortOrder,
			RuleSummary: FeatureRuleSummary(r.FeatureId),
		})
	}
	return out, nil
}

// AdminUpsertFeatureProduct 只更新已有功能 SKU，不新建商品编码。
//
// 业务：空编码或库中不存在的编码直接拒绝。已有行更新价格、天数、Apple ID、上下架等，不改 product_code。
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
	if p.FeatureId == FeatureIDPredictionUnlock || p.GrantKind == GrantKindAllowedCountDelta {
		glog.Warningf(ctx, "[cash] prediction_unlock 已下线，拒绝 SKU featureId=%s kind=%s", p.FeatureId, p.GrantKind)
		return gerror.NewCode(gcode.CodeInvalidOperation, "预测事项开通数量已下线")
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
	if p.DurationDays < 1 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "有效天数须≥1，不支持永久")
	}
	if p.ProductCode == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "须选择已有套餐，不能新建")
	}

	r, err := g.DB().Model("feature_product").Ctx(ctx).Where("product_code", p.ProductCode).One()
	if err != nil {
		return err
	}
	if r.IsEmpty() {
		return gerror.NewCode(gcode.CodeInvalidParameter, "商品编码不存在，不能新建")
	}
	_, err = g.DB().Model("feature_product").Ctx(ctx).Where("product_code", p.ProductCode).Data(g.Map{
		"feature_id": p.FeatureId, "grant_kind": p.GrantKind,
		"grant_quantity": p.GrantQuantity, "price_fen": p.PriceFen, "original_price_fen": p.OriginalPriceFen,
		"duration_days": p.DurationDays, "apple_product_id": p.AppleProductId, "status": p.Status, "updated_at": now,
	}).Update()
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
