package v1

import "github.com/gogf/gf/v2/frame/g"

// —— App：UCG 资格 / 功能目录 / 建单 / 邀请码 / 广告 ——

// CashUCGEligibilityReq GET UCG 入场资格。
type CashUCGEligibilityReq struct {
	g.Meta `path:"/cash/app/api/ucg/eligibility" method:"get" tags:"cash" summary:"UCG 入场资格（连续有效喂养日）"`
}

// CashUCGEligibilityRes 资格 data。
type CashUCGEligibilityRes struct {
	Qualified     bool   `json:"qualified"`
	RequiredDays  int    `json:"requiredDays"`
	EffectiveDays int    `json:"effectiveDays"`
	RemainingDays int    `json:"remainingDays"`
	Message       string `json:"message,omitempty"`
}

// CashCareAlertEligibilityReq GET 值得留意喂养资格（cash 合成）。
type CashCareAlertEligibilityReq struct {
	g.Meta `path:"/cash/app/api/care-alert/eligibility" method:"get" tags:"cash" summary:"值得留意喂养资格（连续有效喂养日）"`
}

// CashCareAlertEligibilityRes 与 UCG 资格字段同构。
type CashCareAlertEligibilityRes struct {
	Qualified     bool   `json:"qualified"`
	RequiredDays  int    `json:"requiredDays"`
	EffectiveDays int    `json:"effectiveDays"`
	RemainingDays int    `json:"remainingDays"`
	Message       string `json:"message,omitempty"`
}

// CashFeatureCatalogReq GET 合成功能目录。
type CashFeatureCatalogReq struct {
	g.Meta `path:"/cash/app/api/feature/catalog" method:"get" tags:"cash" summary:"功能目录（含是否已开通与可售 SKU）"`
}

// CashFeatureCatalogProductItem 目录项内嵌可售 SKU（供展示标价与建单 productCode）。
type CashFeatureCatalogProductItem struct {
	ProductCode      string `json:"productCode"`
	PriceFen         int    `json:"priceFen"`
	OriginalPriceFen int    `json:"originalPriceFen"`
	DurationDays     int    `json:"durationDays"`
	GrantKind        string `json:"grantKind"`
	GrantQuantity    int    `json:"grantQuantity"`
	AppleProductId   string `json:"appleProductId,omitempty"`
}

// CashFeatureCatalogItem 目录项（开通态 + 可售 products）。
// AllowedCount：预测类有效可看条数；-1 表示临时/永久全开（哨兵，客户端须识别）。
type CashFeatureCatalogItem struct {
	FeatureId     string                         `json:"featureId"`
	Title         string                         `json:"title"`
	Description   string                         `json:"description"`
	UnlockMethods string                         `json:"unlockMethods"`
	Unlocked      bool                           `json:"unlocked"`
	UnlockMethod  string                         `json:"unlockMethod,omitempty"`
	ExpiresAt     int64                          `json:"expiresAt,omitempty"`
	AllowedCount  *int                           `json:"allowedCount,omitempty"`
	Products      []CashFeatureCatalogProductItem `json:"products"`
}

// CashFeatureCatalogRes 目录 data。
type CashFeatureCatalogRes struct {
	List []CashFeatureCatalogItem `json:"list"`
}

// CashFeatureCreateOrderReq POST 功能建单。
type CashFeatureCreateOrderReq struct {
	g.Meta      `path:"/cash/app/api/feature/orders" method:"post" tags:"cash" summary:"创建功能开通订单"`
	ProductCode string `json:"productCode" v:"required"`
	Channel     string `json:"channel" v:"required" dc:"alipay|apple_iap"`
}

// CashFeatureCreateOrderRes 建单 data（字段对齐 VIP 建单）。
type CashFeatureCreateOrderRes struct {
	OrderNo        string `json:"orderNo"`
	ProductCode    string `json:"productCode"`
	Channel        string `json:"channel"`
	AmountFen      int    `json:"amountFen"`
	AppleProductId string `json:"appleProductId,omitempty"`
	AlipayOrderStr string `json:"alipayOrderStr,omitempty"`
	PayTip         string `json:"payTip,omitempty"`
}

// CashFeatureInviteRedeemReq POST 邀请码兑换单功能。
type CashFeatureInviteRedeemReq struct {
	g.Meta    `path:"/cash/app/api/feature/invite-codes/redeem" method:"post" tags:"cash" summary:"邀请码兑换开通单个功能"`
	Code      string `json:"code" v:"required"`
	FeatureId string `json:"featureId" v:"required"`
}

// CashFeatureInviteRedeemRes 兑换结果。
type CashFeatureInviteRedeemRes struct{}

// CashFeatureAdCompleteReq POST 广告完成开通。
type CashFeatureAdCompleteReq struct {
	g.Meta       `path:"/cash/app/api/feature/ad/complete" method:"post" tags:"cash" summary:"广告完成开通功能（MVP）"`
	FeatureId    string `json:"featureId" v:"required"`
	IdempotencyKey string `json:"idempotencyKey" dc:"短窗幂等键"`
}

// CashFeatureAdCompleteRes 广告开通结果。
type CashFeatureAdCompleteRes struct{}

// —— Admin：功能 ——

// CashAdminFeatureDefsListReq GET 功能定义列表。
type CashAdminFeatureDefsListReq struct {
	g.Meta `path:"/cash/admin/api/feature/defs" method:"get" tags:"cash-admin" summary:"管理端功能定义列表"`
}

// CashAdminFeatureDefsListRes 功能定义列表。
type CashAdminFeatureDefsListRes struct {
	List []CashAdminFeatureDefItem `json:"list"`
}

// CashAdminFeatureDefItem 功能定义项。
type CashAdminFeatureDefItem struct {
	FeatureId           string `json:"featureId"`
	Title               string `json:"title"`
	Description         string `json:"description"`
	UnlockMethods       string `json:"unlockMethods"`
	DurationDays        int    `json:"durationDays"`
	DefaultAllowedCount int    `json:"defaultAllowedCount"`
	Status              int    `json:"status"`
	SortOrder           int    `json:"sortOrder"`
}

// CashAdminFeatureDefUpsertReq POST 更新功能定义（禁止新建未知 featureId）。
type CashAdminFeatureDefUpsertReq struct {
	g.Meta              `path:"/cash/admin/api/feature/defs" method:"post" tags:"cash-admin" summary:"管理端更新功能定义（编号只读）"`
	FeatureId           string `json:"featureId" v:"required"`
	Title               string `json:"title"`
	Description         string `json:"description"`
	UnlockMethods       string `json:"unlockMethods"`
	DurationDays        int    `json:"durationDays"`
	DefaultAllowedCount int    `json:"defaultAllowedCount"`
	Status              int    `json:"status" d:"1"`
	SortOrder           int    `json:"sortOrder"`
}

// CashAdminFeatureDefUpsertRes 空。
type CashAdminFeatureDefUpsertRes struct{}

// CashAdminFeatureProductsListReq GET 功能 SKU。
type CashAdminFeatureProductsListReq struct {
	g.Meta `path:"/cash/admin/api/feature/products" method:"get" tags:"cash-admin" summary:"管理端功能 SKU 列表"`
}

// CashAdminFeatureProductItem SKU 项。
type CashAdminFeatureProductItem struct {
	ProductCode      string `json:"productCode"`
	FeatureId        string `json:"featureId"`
	GrantKind        string `json:"grantKind"`
	GrantQuantity    int    `json:"grantQuantity"`
	PriceFen         int    `json:"priceFen"`
	OriginalPriceFen int    `json:"originalPriceFen"`
	DurationDays     int    `json:"durationDays"`
	AppleProductId   string `json:"appleProductId"`
	Status           int    `json:"status"`
}

// CashAdminFeatureProductsListRes SKU 列表。
type CashAdminFeatureProductsListRes struct {
	List []CashAdminFeatureProductItem `json:"list"`
}

// CashAdminFeatureProductUpsertReq POST SKU（新建 productCode 可空，服务端自动生成）。
type CashAdminFeatureProductUpsertReq struct {
	g.Meta           `path:"/cash/admin/api/feature/products" method:"post" tags:"cash-admin" summary:"管理端创建或更新功能 SKU"`
	ProductCode      string `json:"productCode" dc:"空则自动生成；更新时必填已有编码"`
	FeatureId        string `json:"featureId" v:"required"`
	GrantKind        string `json:"grantKind"`
	GrantQuantity    int    `json:"grantQuantity" d:"1"`
	PriceFen         int    `json:"priceFen"`
	OriginalPriceFen int    `json:"originalPriceFen"`
	DurationDays     int    `json:"durationDays"`
	AppleProductId   string `json:"appleProductId"`
	Status           int    `json:"status" d:"1"`
}

// CashAdminFeatureProductUpsertRes 创建/更新结果（含最终商品编码）。
type CashAdminFeatureProductUpsertRes struct {
	ProductCode string `json:"productCode"`
}

// —— Admin：邀请码 ——

// CashAdminInviteCodesListReq GET 邀请码列表。
type CashAdminInviteCodesListReq struct {
	g.Meta `path:"/cash/admin/api/invite-code/list" method:"get" tags:"cash-admin" summary:"管理端邀请码列表"`
}

// CashAdminInviteCodeItem 邀请码项。
type CashAdminInviteCodeItem struct {
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

// CashAdminInviteCodesListRes 列表。
type CashAdminInviteCodesListRes struct {
	List []CashAdminInviteCodeItem `json:"list"`
}

// CashAdminInviteCodeCreateReq POST 创建邀请码。
type CashAdminInviteCodeCreateReq struct {
	g.Meta            `path:"/cash/admin/api/invite-code" method:"post" tags:"cash-admin" summary:"管理端创建邀请码"`
	OwnerWxId         int64    `json:"ownerWxId" v:"required"`
	Code              string   `json:"code" dc:"空则自动生成"`
	ExpiresAt         int64    `json:"expiresAt"`
	MaxRedemptions    int      `json:"maxRedemptions"`
	GrantDurationDays int      `json:"grantDurationDays"`
	FeatureIds        []string `json:"featureIds"`
	Status            int      `json:"status" d:"1"`
}

// CashAdminInviteCodeCreateRes 创建结果。
type CashAdminInviteCodeCreateRes struct {
	Code string `json:"code"`
}

// CashAdminInviteCodeStatusReq POST 停用/启用。
type CashAdminInviteCodeStatusReq struct {
	g.Meta `path:"/cash/admin/api/invite-code/status" method:"post" tags:"cash-admin" summary:"管理端停用或启用邀请码"`
	Code   string `json:"code" v:"required"`
	Status int    `json:"status" v:"required"`
}

// CashAdminInviteCodeStatusRes 空。
type CashAdminInviteCodeStatusRes struct{}

// CashAdminInviteRedemptionsReq GET 兑换明细。
type CashAdminInviteRedemptionsReq struct {
	g.Meta `path:"/cash/admin/api/invite-code/redemptions" method:"get" tags:"cash-admin" summary:"管理端邀请码兑换明细"`
	Code   string `json:"code" in:"query"`
	Limit  int    `json:"limit" in:"query" d:"100"`
}

// CashAdminInviteRedemptionItem 明细行。
type CashAdminInviteRedemptionItem struct {
	Code         string `json:"code"`
	OwnerWxId    int64  `json:"ownerWxId"`
	RedeemerWxId int64  `json:"redeemerWxId"`
	DeviceNo     string `json:"deviceNo"`
	FeatureId    string `json:"featureId"`
	RedeemedAt   int64  `json:"redeemedAt"`
}

// CashAdminInviteRedemptionsRes 明细。
type CashAdminInviteRedemptionsRes struct {
	List []CashAdminInviteRedemptionItem `json:"list"`
}

// —— Admin：喂养资格场景 ——

// CashAdminFeedingEligibilityScenesListReq GET 场景阈值列表。
type CashAdminFeedingEligibilityScenesListReq struct {
	g.Meta `path:"/cash/admin/api/feeding-eligibility/scenes" method:"get" tags:"cash-admin" summary:"管理端喂养资格场景列表"`
}

// CashAdminFeedingEligibilitySceneItem 场景项。
type CashAdminFeedingEligibilitySceneItem struct {
	SceneKey         string `json:"sceneKey"`
	RequiredDays     int    `json:"requiredDays"`
	MinRecordsPerDay int    `json:"minRecordsPerDay"`
	UpdatedAt        int64  `json:"updatedAt"`
}

// CashAdminFeedingEligibilityScenesListRes 列表。
type CashAdminFeedingEligibilityScenesListRes struct {
	List []CashAdminFeedingEligibilitySceneItem `json:"list"`
}

// CashAdminFeedingEligibilitySceneUpdateReq POST 更新已有场景。
type CashAdminFeedingEligibilitySceneUpdateReq struct {
	g.Meta           `path:"/cash/admin/api/feeding-eligibility/scenes" method:"post" tags:"cash-admin" summary:"管理端更新喂养资格场景阈值"`
	SceneKey         string `json:"sceneKey" v:"required"`
	RequiredDays     int    `json:"requiredDays" v:"required|min:1"`
	MinRecordsPerDay int    `json:"minRecordsPerDay" v:"required|min:1"`
}

// CashAdminFeedingEligibilitySceneUpdateRes 空。
type CashAdminFeedingEligibilitySceneUpdateRes struct{}

