package cash

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/frame/g"
)

// FeatureDefRow 功能定义行。
type FeatureDefRow struct {
	FeatureId     string `json:"featureId"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	UnlockMethods string `json:"unlockMethods"`
	DurationDays  int    `json:"durationDays"`
	Status        int    `json:"status"`
	SortOrder     int    `json:"sortOrder"`
}

// FeatureCatalogProduct 目录项内嵌可售 SKU（仅 status=1）。
type FeatureCatalogProduct struct {
	ProductCode      string `json:"productCode"`
	PriceFen         int    `json:"priceFen"`
	OriginalPriceFen int    `json:"originalPriceFen"`
	DurationDays     int    `json:"durationDays"`
	GrantKind        string `json:"grantKind"`
	GrantQuantity    int    `json:"grantQuantity"`
	AppleProductId   string `json:"appleProductId,omitempty"`
}

// FeatureCatalogItem 合成目录项（含设备开通态与可售 SKU）。
type FeatureCatalogItem struct {
	FeatureId     string                  `json:"featureId"`
	Title         string                  `json:"title"`
	Description   string                  `json:"description"`
	UnlockMethods string                  `json:"unlockMethods"`
	Unlocked      bool                    `json:"unlocked"`
	UnlockMethod  string                  `json:"unlockMethod,omitempty"`
	ExpiresAt     int64                   `json:"expiresAt,omitempty"`
	AllowedCount  *int                    `json:"allowedCount,omitempty"`
	Products      []FeatureCatalogProduct `json:"products"`
}

type featureDefDB struct {
	FeatureId     string `json:"feature_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	UnlockMethods string `json:"unlock_methods"`
	DurationDays  int    `json:"duration_days"`
	Status        int    `json:"status"`
	SortOrder     int    `json:"sort_order"`
}

// ListActiveFeatureDefs 启用功能定义（全站 Redis 热读）。
func ListActiveFeatureDefs(ctx context.Context) ([]FeatureDefRow, error) {
	c := cachekit.Default()
	key := cachekit.CashFeatureDefCatalogKey()
	if raw, ok, err := c.Get(ctx, key); err == nil && ok && raw != "" {
		var cached []FeatureDefRow
		if json.Unmarshal([]byte(raw), &cached) == nil {
			return cached, nil
		}
	}
	var rawRows []featureDefDB
	err := g.DB().Model("feature_def").Ctx(ctx).
		Fields("feature_id,title,description,unlock_methods,duration_days,status,sort_order").
		Where("status", 1).
		OrderAsc("sort_order").OrderAsc("feature_id").
		Scan(&rawRows)
	if err != nil {
		return nil, err
	}
	rows := make([]FeatureDefRow, 0, len(rawRows))
	for _, r := range rawRows {
		rows = append(rows, FeatureDefRow{
			FeatureId: r.FeatureId, Title: r.Title, Description: r.Description,
			UnlockMethods: r.UnlockMethods, DurationDays: r.DurationDays,
			Status: r.Status, SortOrder: r.SortOrder,
		})
	}
	if b, mErr := json.Marshal(rows); mErr == nil {
		_ = c.SetEX(ctx, key, string(b), 10*time.Minute)
	}
	return rows, nil
}

// GetFeatureCatalog 合成目录：定义 ⊕ 设备权益 ⊕ allowedCount ⊕ 可售 SKU；不含 UCG。
//
// 业务：客户端一次读齐开通态与支付展示；建单仍只信服务端 productCode 对应价。
func GetFeatureCatalog(ctx context.Context, deviceNo string) ([]FeatureCatalogItem, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	defs, err := ListActiveFeatureDefs(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	type entRow struct {
		FeatureId    string `json:"feature_id"`
		UnlockMethod string `json:"unlock_method"`
		ExpiresAt    int64  `json:"expires_at"`
	}
	var ents []entRow
	_ = g.DB().Model("feature_entitlement").Ctx(ctx).
		Fields("feature_id,unlock_method,expires_at").
		Where("device_no", deviceNo).
		Scan(&ents)
	entMap := map[string]entRow{}
	for _, e := range ents {
		if e.ExpiresAt > 0 && e.ExpiresAt < now {
			continue
		}
		entMap[e.FeatureId] = e
	}
	allowed, _ := GetAllowedCount(ctx, deviceNo)

	// 一次拉取全部启用 SKU，按 feature_id 分组（字典小，避免 N+1）。
	prodByFeature, err := listActiveProductsByFeature(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]FeatureCatalogItem, 0, len(defs))
	for _, d := range defs {
		item := FeatureCatalogItem{
			FeatureId: d.FeatureId, Title: d.Title, Description: d.Description,
			UnlockMethods: d.UnlockMethods,
			Products:      prodByFeature[d.FeatureId],
		}
		if item.Products == nil {
			item.Products = []FeatureCatalogProduct{}
		}
		if d.FeatureId == FeatureIDPredictionUnlock {
			ac := allowed
			item.AllowedCount = &ac
			item.Unlocked = ac > 0
		}
		if e, ok := entMap[d.FeatureId]; ok {
			item.Unlocked = true
			item.UnlockMethod = e.UnlockMethod
			item.ExpiresAt = e.ExpiresAt
		}
		out = append(out, item)
	}
	return out, nil
}

// listActiveProductsByFeature 返回 feature_id → 启用中 SKU 列表。
func listActiveProductsByFeature(ctx context.Context) (map[string][]FeatureCatalogProduct, error) {
	var raw []featureProductDB
	err := g.DB().Model("feature_product").Ctx(ctx).
		Where("status", 1).
		OrderAsc("product_code").
		Scan(&raw)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]FeatureCatalogProduct, len(raw))
	for _, r := range raw {
		p := FeatureCatalogProduct{
			ProductCode: r.ProductCode, PriceFen: r.PriceFen, OriginalPriceFen: r.OriginalPriceFen,
			DurationDays: r.DurationDays, GrantKind: r.GrantKind, GrantQuantity: r.GrantQuantity,
			AppleProductId: r.AppleProductId,
		}
		out[r.FeatureId] = append(out[r.FeatureId], p)
	}
	return out, nil
}
