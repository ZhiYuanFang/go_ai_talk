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
	if p.FeatureId == FeatureIDPredictionUnlock {
		p.GrantKind = GrantKindAllowedCountDelta
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
