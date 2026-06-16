package ucg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hello/internal/services/contracts"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	ucgQuotaDefaultSingletonID = 1
	ucgQuotaUsageKeyPrefix     = "ai:usage:"
	ucgQuotaUsageTTLSeconds    = 90 * 24 * 3600
)

var ucgShanghaiLoc *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		ucgShanghaiLoc = time.FixedZone("CST", 8*3600)
		return
	}
	ucgShanghaiLoc = loc
}

func ucgQuotaMonthBucket() string {
	return time.Now().In(ucgShanghaiLoc).Format("200601")
}

func polishUsageRedisKey(wxID int64) string {
	return fmt.Sprintf("%s%s:%d:%s", ucgQuotaUsageKeyPrefix, contracts.AIQuotaPolish, wxID, ucgQuotaMonthBucket())
}

func validateWxIDForPolish(wxID int64) error {
	if wxID <= 0 {
		return contracts.ErrAINotLoggedIn
	}
	return nil
}

// EnsurePolishAIQuotaDefaultRow 保证 singleton 行存在（默认 5）。
func EnsurePolishAIQuotaDefaultRow(ctx context.Context) error {
	n, err := g.DB().Model("ai_quota_default").Ctx(ctx).Where("id", ucgQuotaDefaultSingletonID).Count()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	now := time.Now().Unix()
	_, err = g.DB().Model("ai_quota_default").Ctx(ctx).Data(g.Map{
		"id":                   ucgQuotaDefaultSingletonID,
		"polish_monthly_limit": 5,
		"updated_at":           now,
	}).Insert()
	return err
}

type polishQuotaDefaultRow struct {
	PolishMonthlyLimit int   `json:"polishMonthlyLimit"`
	UpdatedAt          int64 `json:"updatedAt"`
}

func loadPolishAIQuotaDefault(ctx context.Context) (polishQuotaDefaultRow, error) {
	if err := EnsurePolishAIQuotaDefaultRow(ctx); err != nil {
		return polishQuotaDefaultRow{}, err
	}
	var row polishQuotaDefaultRow
	if err := g.DB().Model("ai_quota_default").Ctx(ctx).Where("id", ucgQuotaDefaultSingletonID).Scan(&row); err != nil {
		return polishQuotaDefaultRow{}, err
	}
	if row.PolishMonthlyLimit <= 0 {
		row.PolishMonthlyLimit = 5
	}
	return row, nil
}

type polishQuotaOverrideRow struct {
	WxId               int64 `json:"wxId"`
	PolishMonthlyLimit *int  `json:"polishMonthlyLimit"`
	UpdatedAt          int64 `json:"updatedAt"`
}

func effectivePolishLimit(ctx context.Context, wxID int64) (int, error) {
	def, err := loadPolishAIQuotaDefault(ctx)
	if err != nil {
		return 0, err
	}
	limit := def.PolishMonthlyLimit
	var ov polishQuotaOverrideRow
	_ = g.DB().Model("ai_quota_user_override").Ctx(ctx).Where("wx_id", wxID).Scan(&ov)
	if ov.WxId == wxID && ov.PolishMonthlyLimit != nil && *ov.PolishMonthlyLimit > 0 {
		limit = *ov.PolishMonthlyLimit
	}
	return limit, nil
}

func readPolishUsageCount(ctx context.Context, wxID int64) (int, error) {
	key := polishUsageRedisKey(wxID)
	v, err := g.Redis().Do(ctx, "GET", key)
	if err != nil {
		return 0, err
	}
	if v.IsNil() || v.IsEmpty() {
		return 0, nil
	}
	return v.Int(), nil
}

func touchPolishUsageKeyTTL(ctx context.Context, key string) {
	_, _ = g.Redis().Do(ctx, "EXPIRE", key, ucgQuotaUsageTTLSeconds)
}

// CheckPolishAIQuota 润笔预检，不修改用量。
func CheckPolishAIQuota(ctx context.Context, wxID int64) (contracts.AIQuotaSnapshot, error) {
	if err := validateWxIDForPolish(wxID); err != nil {
		return contracts.AIQuotaSnapshot{}, err
	}
	limit, err := effectivePolishLimit(ctx, wxID)
	if err != nil {
		return contracts.AIQuotaSnapshot{}, err
	}
	used, err := readPolishUsageCount(ctx, wxID)
	if err != nil {
		return contracts.AIQuotaSnapshot{}, err
	}
	return contracts.AIQuotaSnapshot{
		Used:    used,
		Limit:   limit,
		Allowed: used < limit,
	}, nil
}

// ConsumePolishAIQuota 润笔成功扣减；超额回滚 INCR。
func ConsumePolishAIQuota(ctx context.Context, wxID int64) (contracts.AIQuotaSnapshot, error) {
	if err := validateWxIDForPolish(wxID); err != nil {
		return contracts.AIQuotaSnapshot{}, err
	}
	limit, err := effectivePolishLimit(ctx, wxID)
	if err != nil {
		return contracts.AIQuotaSnapshot{}, err
	}
	key := polishUsageRedisKey(wxID)
	n, err := g.Redis().Do(ctx, "INCR", key)
	if err != nil {
		return contracts.AIQuotaSnapshot{}, err
	}
	touchPolishUsageKeyTTL(ctx, key)
	used := n.Int()
	if used > limit {
		_, _ = g.Redis().Do(ctx, "DECR", key)
		return contracts.AIQuotaSnapshot{Used: limit, Limit: limit, Allowed: false}, contracts.ErrAIQuotaExhausted
	}
	return contracts.AIQuotaSnapshot{Used: used, Limit: limit, Allowed: true}, nil
}

// GetPolishAIQuotaAppStatus App 读 API polish 快照。
func GetPolishAIQuotaAppStatus(ctx context.Context, wxID int64) (contracts.PolishAIQuotaAppStatus, error) {
	snap, err := CheckPolishAIQuota(ctx, wxID)
	if err != nil {
		return contracts.PolishAIQuotaAppStatus{}, err
	}
	return contracts.PolishAIQuotaAppStatus{Polish: snap}, nil
}

// GetPolishAIQuotaDefaultForAdmin 读取全局润笔默认。
func GetPolishAIQuotaDefaultForAdmin(ctx context.Context) (contracts.PolishAIQuotaDefaultDTO, error) {
	row, err := loadPolishAIQuotaDefault(ctx)
	if err != nil {
		return contracts.PolishAIQuotaDefaultDTO{}, err
	}
	return contracts.PolishAIQuotaDefaultDTO{
		PolishMonthlyLimit: row.PolishMonthlyLimit,
		UpdatedAt:          row.UpdatedAt,
	}, nil
}

// UpdatePolishAIQuotaDefaultForAdmin 更新全局润笔默认。
func UpdatePolishAIQuotaDefaultForAdmin(ctx context.Context, polishLimit int) (contracts.PolishAIQuotaDefaultDTO, error) {
	if polishLimit <= 0 {
		return contracts.PolishAIQuotaDefaultDTO{}, errors.New("polishMonthlyLimit 须为正整数")
	}
	if err := EnsurePolishAIQuotaDefaultRow(ctx); err != nil {
		return contracts.PolishAIQuotaDefaultDTO{}, err
	}
	now := time.Now().Unix()
	_, err := g.DB().Model("ai_quota_default").Ctx(ctx).Where("id", ucgQuotaDefaultSingletonID).Data(g.Map{
		"polish_monthly_limit": polishLimit,
		"updated_at":           now,
	}).Update()
	if err != nil {
		return contracts.PolishAIQuotaDefaultDTO{}, err
	}
	return contracts.PolishAIQuotaDefaultDTO{PolishMonthlyLimit: polishLimit, UpdatedAt: now}, nil
}

// GetPolishAIQuotaUserOverrideForAdmin 读取 wxId 润笔 override。
func GetPolishAIQuotaUserOverrideForAdmin(ctx context.Context, wxID int64) (contracts.PolishAIQuotaUserOverrideDTO, error) {
	if wxID <= 0 {
		return contracts.PolishAIQuotaUserOverrideDTO{}, errors.New("wxId 无效")
	}
	var row polishQuotaOverrideRow
	err := g.DB().Model("ai_quota_user_override").Ctx(ctx).Where("wx_id", wxID).Scan(&row)
	if err != nil {
		return contracts.PolishAIQuotaUserOverrideDTO{}, err
	}
	if row.WxId != wxID {
		return contracts.PolishAIQuotaUserOverrideDTO{WxId: wxID}, nil
	}
	return contracts.PolishAIQuotaUserOverrideDTO{
		WxId:               row.WxId,
		PolishMonthlyLimit: row.PolishMonthlyLimit,
		UpdatedAt:          row.UpdatedAt,
	}, nil
}

// UpdatePolishAIQuotaUserOverrideForAdmin 写入或清除润笔 override。
func UpdatePolishAIQuotaUserOverrideForAdmin(ctx context.Context, wxID int64, polishLimit *int) (contracts.PolishAIQuotaUserOverrideDTO, error) {
	if wxID <= 0 {
		return contracts.PolishAIQuotaUserOverrideDTO{}, errors.New("wxId 无效")
	}
	if polishLimit != nil && *polishLimit <= 0 {
		return contracts.PolishAIQuotaUserOverrideDTO{}, errors.New("polishMonthlyLimit 须为正整数")
	}
	now := time.Now().Unix()
	data := g.Map{
		"wx_id":      wxID,
		"updated_at": now,
	}
	if polishLimit == nil {
		data["polish_monthly_limit"] = nil
	} else {
		data["polish_monthly_limit"] = *polishLimit
	}
	_, err := g.DB().Model("ai_quota_user_override").Ctx(ctx).Data(data).Save()
	if err != nil {
		return contracts.PolishAIQuotaUserOverrideDTO{}, err
	}
	return GetPolishAIQuotaUserOverrideForAdmin(ctx, wxID)
}
