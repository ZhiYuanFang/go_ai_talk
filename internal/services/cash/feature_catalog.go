package cash

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"hello/internal/platform/cachekit"
	"hello/internal/shared/featurelogo"

	"github.com/gogf/gf/v2/frame/g"
)

// FeatureDefRow 功能定义行。
type FeatureDefRow struct {
	FeatureId           string `json:"featureId"`
	Title               string `json:"title"`
	Description         string `json:"description"`
	UnlockMethods       string `json:"unlockMethods"`
	DurationDays        int    `json:"durationDays"`
	InviteDurationDays  int    `json:"inviteDurationDays"`
	AdDurationDays      int    `json:"adDurationDays"`
	DefaultAllowedCount int    `json:"defaultAllowedCount"`
	ActivationSubject   string `json:"activationSubject"` // device|user
	Logo                string `json:"logo"`              // 库内 objectKey；HTTP 边界再映 CDN
	Color               string `json:"color"`             // #RGB / #RRGGBB
	Status              int    `json:"status"`
	SortOrder           int    `json:"sortOrder"`
	RuleSummary         string `json:"ruleSummary,omitempty"` // 只读开通规则（运维展示）
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

// FeatureCatalogItem 合成目录项（含设备/账号开通态与可售 SKU）。
type FeatureCatalogItem struct {
	FeatureId             string                  `json:"featureId"`
	Title                 string                  `json:"title"`
	Description           string                  `json:"description"`
	UnlockMethods         string                  `json:"unlockMethods"`
	Unlocked              bool                    `json:"unlocked"`
	UnlockMethod          string                  `json:"unlockMethod,omitempty"`
	ExpiresAt             int64                   `json:"expiresAt,omitempty"`
	AllowedCount          *int                    `json:"allowedCount,omitempty"`
	DefaultCount          *int                    `json:"defaultCount,omitempty"` // 预测：定义表默认免费条数
	TotalActivatableCount *int                    `json:"totalActivatableCount,omitempty"`
	InviteDurationDays    int                     `json:"inviteDurationDays"`
	AdDurationDays        int                     `json:"adDurationDays"`
	// TrialAvailable 试用未用且当前无用户权益；客户端再与 !VIP 合成展示。
	TrialAvailable bool `json:"trialAvailable"`
	// InviteAvailable unlock_methods 含 invite 且该人尚未 InviteOncePerUser 成功。
	InviteAvailable bool                   `json:"inviteAvailable"`
	Logo            string                 `json:"logo"`  // CDN URL；无则空串
	Color           string                 `json:"color"` // 主色 hex
	Products        []FeatureCatalogProduct `json:"products"`
}

type featureDefDB struct {
	FeatureId           string `json:"feature_id"`
	Title               string `json:"title"`
	Description         string `json:"description"`
	UnlockMethods       string `json:"unlock_methods"`
	DurationDays        int    `json:"duration_days"`
	InviteDurationDays  int    `json:"invite_duration_days"`
	AdDurationDays      int    `json:"ad_duration_days"`
	DefaultAllowedCount int    `json:"default_allowed_count"`
	ActivationSubject   string `json:"activation_subject"`
	Logo                string `json:"logo"`
	Color               string `json:"color"`
	Status              int    `json:"status"`
	SortOrder           int    `json:"sort_order"`
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
		Fields("feature_id,title,description,unlock_methods,duration_days,invite_duration_days,ad_duration_days,default_allowed_count,activation_subject,logo,color,status,sort_order").
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
			InviteDurationDays: r.InviteDurationDays, AdDurationDays: r.AdDurationDays,
			DefaultAllowedCount: r.DefaultAllowedCount,
			ActivationSubject:    NormalizeActivationSubject(r.ActivationSubject),
			Logo: featurelogo.NormalizeObjectKey(r.Logo), Color: strings.TrimSpace(r.Color),
			Status: r.Status, SortOrder: r.SortOrder,
		})
	}
	if b, mErr := json.Marshal(rows); mErr == nil {
		_ = c.SetEX(ctx, key, string(b), 10*time.Minute)
	}
	return rows, nil
}

// FeatureCatalogResult App 目录合成结果（含页级群二维码 URL）。
type FeatureCatalogResult struct {
	List              []FeatureCatalogItem `json:"list"`
	InviteGroupQrUrl  string               `json:"inviteGroupQrUrl,omitempty"`
}

// GetFeatureCatalog 合成目录：定义 ⊕ 设备/账号权益 ⊕ allowedCount ⊕ 可售 SKU；不含 UCG。
//
// 业务：客户端一次读齐开通态与支付展示；建单仍只信服务端 productCode 对应价。
// wxID：user 主体功能用账号权益；<=0 时 user 类视为未解锁（不得用设备权益冒充）。
func GetFeatureCatalog(ctx context.Context, deviceNo string, wxID int64) (*FeatureCatalogResult, error) {
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
	userEntMap := map[string]entRow{}
	if wxID > 0 {
		var uents []entRow
		_ = g.DB().Model("feature_user_entitlement").Ctx(ctx).
			Fields("feature_id,unlock_method,expires_at").
			Where("wx_id", wxID).
			Scan(&uents)
		for _, e := range uents {
			if e.ExpiresAt > 0 && e.ExpiresAt < now {
				continue
			}
			userEntMap[e.FeatureId] = e
		}
	}

	// 一次拉取全部启用 SKU，按 feature_id 分组（字典小，避免 N+1）。
	prodByFeature, err := listActiveProductsByFeature(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]FeatureCatalogItem, 0, len(defs))
	for _, d := range defs {
		// prediction_unlock 已停用；防御性跳过（ListActive 已滤 status=1）。
		if d.FeatureId == FeatureIDPredictionUnlock {
			continue
		}
		item := FeatureCatalogItem{
			FeatureId: d.FeatureId, Title: d.Title, Description: d.Description,
			UnlockMethods:      d.UnlockMethods,
			InviteDurationDays: d.InviteDurationDays,
			AdDurationDays:     d.AdDurationDays,
			Logo:               featurelogo.CdnURL(ctx, d.Logo),
			Color:              strings.TrimSpace(d.Color),
			Products:           prodByFeature[d.FeatureId],
		}
		if item.Products == nil {
			item.Products = []FeatureCatalogProduct{}
		}
		subj := NormalizeActivationSubject(d.ActivationSubject)
		if subj == ActivationSubjectUser {
			if e, ok := userEntMap[d.FeatureId]; ok {
				item.Unlocked = true
				item.UnlockMethod = e.UnlockMethod
				item.ExpiresAt = e.ExpiresAt
			}
		} else if e, ok := entMap[d.FeatureId]; ok {
			item.Unlocked = true
			item.UnlockMethod = e.UnlockMethod
			item.ExpiresAt = e.ExpiresAt
		}
		// trialAvailable：试用未用且无用户权益（VIP 由客户端合成隐藏）。
		if wxID > 0 {
			if unused, tErr := IsTrialUnused(ctx, wxID, d.FeatureId); tErr != nil {
				g.Log().Warningf(ctx, "[cash-catalog] trial 查询失败 feature=%s err=%v", d.FeatureId, tErr)
			} else {
				item.TrialAvailable = unused && !item.Unlocked
			}
			if strings.Contains(d.UnlockMethods, UnlockMethodInviteCode) {
				if InviteOncePerUser(d.FeatureId) {
					has, iErr := HasInviteFeatureGrantAnyCode(ctx, wxID, d.FeatureId)
					if iErr != nil {
						g.Log().Warningf(ctx, "[cash-catalog] invite grant 查询失败 feature=%s err=%v", d.FeatureId, iErr)
					} else {
						item.InviteAvailable = !has
					}
				} else {
					item.InviteAvailable = true
				}
			}
		}
		out = append(out, item)
	}
	qrURL, _ := ResolveInviteGroupQrURLForApp(ctx)
	return &FeatureCatalogResult{List: out, InviteGroupQrUrl: qrURL}, nil
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
