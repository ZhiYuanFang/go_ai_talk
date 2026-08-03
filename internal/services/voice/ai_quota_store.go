package voice

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hello/internal/platform/cachekit"
	"hello/internal/services/contracts"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	aiQuotaDefaultSingletonID = 1
	aiQuotaUsageTTLSeconds    = 90 * 24 * 3600
)

var voiceQuotaCache = cachekit.Default()

var shanghaiLoc *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		shanghaiLoc = time.FixedZone("CST", 8*3600)
		return
	}
	shanghaiLoc = loc
}

func aiQuotaMonthBucket() string {
	return time.Now().In(shanghaiLoc).Format("200601")
}

func aiQuotaUsageRedisKey(feature contracts.AIQuotaFeature, wxID int64) string {
	return cachekit.AIQuotaUsageKey(string(feature), wxID, aiQuotaMonthBucket())
}

func validateVoiceQuotaFeature(feature contracts.AIQuotaFeature) error {
	switch feature {
	case contracts.AIQuotaVoiceAI, contracts.AIQuotaClinicAI:
		return nil
	default:
		return fmt.Errorf("未知 feature: %s", feature)
	}
}

func validateWxIDForAI(wxID int64) error {
	if wxID <= 0 {
		return contracts.ErrAINotLoggedIn
	}
	return nil
}

// EnsureVoiceAIQuotaDefaultRow 保证 singleton 行存在（默认 5/30）。
func EnsureVoiceAIQuotaDefaultRow(ctx context.Context) error {
	n, err := g.DB().Model("ai_quota_default").Ctx(ctx).Where("id", aiQuotaDefaultSingletonID).Count()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	now := time.Now().Unix()
	_, err = g.DB().Model("ai_quota_default").Ctx(ctx).Data(g.Map{
		"id":                      aiQuotaDefaultSingletonID,
		"voice_ai_monthly_limit":  5,
		"clinic_ai_monthly_limit": 30,
		"updated_at":              now,
	}).Insert()
	return err
}

type voiceQuotaDefaultRow struct {
	VoiceAiMonthlyLimit  int   `json:"voiceAiMonthlyLimit"`
	ClinicAiMonthlyLimit int   `json:"clinicAiMonthlyLimit"`
	UpdatedAt            int64 `json:"updatedAt"`
}

func loadVoiceAIQuotaDefault(ctx context.Context) (voiceQuotaDefaultRow, error) {
	if err := EnsureVoiceAIQuotaDefaultRow(ctx); err != nil {
		return voiceQuotaDefaultRow{}, err
	}
	var row voiceQuotaDefaultRow
	if err := g.DB().Model("ai_quota_default").Ctx(ctx).Where("id", aiQuotaDefaultSingletonID).Scan(&row); err != nil {
		return voiceQuotaDefaultRow{}, err
	}
	if row.VoiceAiMonthlyLimit <= 0 {
		row.VoiceAiMonthlyLimit = 5
	}
	if row.ClinicAiMonthlyLimit <= 0 {
		row.ClinicAiMonthlyLimit = 30
	}
	return row, nil
}

type voiceQuotaOverrideRow struct {
	WxId                 int64 `json:"wxId"`
	VoiceAiMonthlyLimit  *int  `json:"voiceAiMonthlyLimit"`
	ClinicAiMonthlyLimit *int  `json:"clinicAiMonthlyLimit"`
	UpdatedAt            int64 `json:"updatedAt"`
}

func effectiveVoiceLimitForFeature(ctx context.Context, wxID int64, feature contracts.AIQuotaFeature) (int, error) {
	def, err := loadVoiceAIQuotaDefault(ctx)
	if err != nil {
		return 0, err
	}
	limit := def.VoiceAiMonthlyLimit
	if feature == contracts.AIQuotaClinicAI {
		limit = def.ClinicAiMonthlyLimit
	}
	var ov voiceQuotaOverrideRow
	_ = g.DB().Model("ai_quota_user_override").Ctx(ctx).Where("wx_id", wxID).Scan(&ov)
	if ov.WxId == wxID {
		if feature == contracts.AIQuotaVoiceAI && ov.VoiceAiMonthlyLimit != nil && *ov.VoiceAiMonthlyLimit > 0 {
			limit = *ov.VoiceAiMonthlyLimit
		}
		if feature == contracts.AIQuotaClinicAI && ov.ClinicAiMonthlyLimit != nil && *ov.ClinicAiMonthlyLimit > 0 {
			limit = *ov.ClinicAiMonthlyLimit
		}
	}
	return limit, nil
}

func readVoiceUsageCount(ctx context.Context, feature contracts.AIQuotaFeature, wxID int64) (int, error) {
	key := aiQuotaUsageRedisKey(feature, wxID)
	v, ok, err := voiceQuotaCache.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, nil
	}
	n, _ := strconv.Atoi(strings.TrimSpace(v))
	return n, nil
}

func touchVoiceUsageKeyTTL(ctx context.Context, key string) {
	_ = voiceQuotaCache.Expire(ctx, key, aiQuotaUsageTTLSeconds*time.Second)
}

// CheckVoiceAIQuotaStore 只读预检，不修改用量。
func CheckVoiceAIQuotaStore(ctx context.Context, wxID int64, feature contracts.AIQuotaFeature) (contracts.AIQuotaSnapshot, error) {
	if err := validateWxIDForAI(wxID); err != nil {
		return contracts.AIQuotaSnapshot{}, err
	}
	if err := validateVoiceQuotaFeature(feature); err != nil {
		return contracts.AIQuotaSnapshot{}, err
	}
	limit, err := effectiveVoiceLimitForFeature(ctx, wxID, feature)
	if err != nil {
		return contracts.AIQuotaSnapshot{}, err
	}
	used, err := readVoiceUsageCount(ctx, feature, wxID)
	if err != nil {
		return contracts.AIQuotaSnapshot{}, err
	}
	allowed := used < limit
	// clinic_ai / voice_ai 用尽时均走降速 fallback（Degraded=true），由调用方强制种子模型且不计次。
	degraded := !allowed && (feature == contracts.AIQuotaClinicAI || feature == contracts.AIQuotaVoiceAI)
	return contracts.AIQuotaSnapshot{
		Used:     used,
		Limit:    limit,
		Allowed:  allowed,
		Degraded: degraded,
	}, nil
}

// ConsumeVoiceAIQuotaStore AI 成功返回后扣减；超额时回滚 INCR。
func ConsumeVoiceAIQuotaStore(ctx context.Context, wxID int64, feature contracts.AIQuotaFeature) (contracts.AIQuotaSnapshot, error) {
	if err := validateWxIDForAI(wxID); err != nil {
		return contracts.AIQuotaSnapshot{}, err
	}
	if err := validateVoiceQuotaFeature(feature); err != nil {
		return contracts.AIQuotaSnapshot{}, err
	}
	limit, err := effectiveVoiceLimitForFeature(ctx, wxID, feature)
	if err != nil {
		return contracts.AIQuotaSnapshot{}, err
	}
	key := aiQuotaUsageRedisKey(feature, wxID)
	n, err := voiceQuotaCache.Incr(ctx, key)
	if err != nil {
		return contracts.AIQuotaSnapshot{}, err
	}
	touchVoiceUsageKeyTTL(ctx, key)
	used := int(n)
	if used > limit {
		_, _ = voiceQuotaCache.Decr(ctx, key)
		return contracts.AIQuotaSnapshot{Used: limit, Limit: limit, Allowed: false}, contracts.ErrAIQuotaExhausted
	}
	return contracts.AIQuotaSnapshot{Used: used, Limit: limit, Allowed: true}, nil
}

// GetVoiceAIQuotaAppStatus 返回 voiceAi + clinicAi 快照。
func GetVoiceAIQuotaAppStatus(ctx context.Context, wxID int64) (contracts.VoiceAIQuotaAppStatus, error) {
	if err := validateWxIDForAI(wxID); err != nil {
		return contracts.VoiceAIQuotaAppStatus{}, err
	}
	voice, err := CheckVoiceAIQuotaStore(ctx, wxID, contracts.AIQuotaVoiceAI)
	if err != nil {
		return contracts.VoiceAIQuotaAppStatus{}, err
	}
	clinic, err := CheckVoiceAIQuotaStore(ctx, wxID, contracts.AIQuotaClinicAI)
	if err != nil {
		return contracts.VoiceAIQuotaAppStatus{}, err
	}
	return contracts.VoiceAIQuotaAppStatus{VoiceAi: voice, ClinicAi: clinic}, nil
}

// GetVoiceAIQuotaDefaultForAdmin 读取全局默认。
func GetVoiceAIQuotaDefaultForAdmin(ctx context.Context) (contracts.VoiceAIQuotaDefaultDTO, error) {
	row, err := loadVoiceAIQuotaDefault(ctx)
	if err != nil {
		return contracts.VoiceAIQuotaDefaultDTO{}, err
	}
	return contracts.VoiceAIQuotaDefaultDTO{
		VoiceAiMonthlyLimit:  row.VoiceAiMonthlyLimit,
		ClinicAiMonthlyLimit: row.ClinicAiMonthlyLimit,
		UpdatedAt:            row.UpdatedAt,
	}, nil
}

// UpdateVoiceAIQuotaDefaultForAdmin 更新全局默认。
func UpdateVoiceAIQuotaDefaultForAdmin(ctx context.Context, voiceLimit, clinicLimit int) (contracts.VoiceAIQuotaDefaultDTO, error) {
	if voiceLimit <= 0 || clinicLimit <= 0 {
		return contracts.VoiceAIQuotaDefaultDTO{}, errors.New("额度须为正整数")
	}
	if err := EnsureVoiceAIQuotaDefaultRow(ctx); err != nil {
		return contracts.VoiceAIQuotaDefaultDTO{}, err
	}
	now := time.Now().Unix()
	_, err := g.DB().Model("ai_quota_default").Ctx(ctx).Where("id", aiQuotaDefaultSingletonID).Data(g.Map{
		"voice_ai_monthly_limit":  voiceLimit,
		"clinic_ai_monthly_limit": clinicLimit,
		"updated_at":              now,
	}).Update()
	if err != nil {
		return contracts.VoiceAIQuotaDefaultDTO{}, err
	}
	return contracts.VoiceAIQuotaDefaultDTO{
		VoiceAiMonthlyLimit:  voiceLimit,
		ClinicAiMonthlyLimit: clinicLimit,
		UpdatedAt:            now,
	}, nil
}

// GetVoiceAIQuotaUserOverrideForAdmin 读取单人 override。
func GetVoiceAIQuotaUserOverrideForAdmin(ctx context.Context, wxID int64) (contracts.VoiceAIQuotaUserOverrideDTO, error) {
	if wxID <= 0 {
		return contracts.VoiceAIQuotaUserOverrideDTO{}, errors.New("wxId 无效")
	}
	var row voiceQuotaOverrideRow
	err := g.DB().Model("ai_quota_user_override").Ctx(ctx).Where("wx_id", wxID).Scan(&row)
	if err != nil {
		return contracts.VoiceAIQuotaUserOverrideDTO{}, err
	}
	if row.WxId != wxID {
		return contracts.VoiceAIQuotaUserOverrideDTO{WxId: wxID}, nil
	}
	return contracts.VoiceAIQuotaUserOverrideDTO{
		WxId:                 row.WxId,
		VoiceAiMonthlyLimit:  row.VoiceAiMonthlyLimit,
		ClinicAiMonthlyLimit: row.ClinicAiMonthlyLimit,
		UpdatedAt:            row.UpdatedAt,
	}, nil
}

// UpdateVoiceAIQuotaUserOverrideForAdmin 写入或清除单人 override（nil 表示清除）。
// 业务逻辑：若提交上限等于当前全局默认，则视为清除该 feature override，便于后续跟随全局。
func UpdateVoiceAIQuotaUserOverrideForAdmin(ctx context.Context, wxID int64, voiceLimit, clinicLimit *int) (contracts.VoiceAIQuotaUserOverrideDTO, error) {
	if wxID <= 0 {
		return contracts.VoiceAIQuotaUserOverrideDTO{}, errors.New("wxId 无效")
	}
	if voiceLimit != nil && *voiceLimit <= 0 {
		return contracts.VoiceAIQuotaUserOverrideDTO{}, errors.New("voiceAiMonthlyLimit 须为正整数")
	}
	if clinicLimit != nil && *clinicLimit <= 0 {
		return contracts.VoiceAIQuotaUserOverrideDTO{}, errors.New("clinicAiMonthlyLimit 须为正整数")
	}
	// 等于全局默认 → 清 override，避免无意义钉死个人值。
	def, err := loadVoiceAIQuotaDefault(ctx)
	if err != nil {
		return contracts.VoiceAIQuotaUserOverrideDTO{}, err
	}
	if voiceLimit != nil && *voiceLimit == def.VoiceAiMonthlyLimit {
		voiceLimit = nil
	}
	if clinicLimit != nil && *clinicLimit == def.ClinicAiMonthlyLimit {
		clinicLimit = nil
	}
	now := time.Now().Unix()
	data := g.Map{
		"wx_id":      wxID,
		"updated_at": now,
	}
	if voiceLimit == nil {
		data["voice_ai_monthly_limit"] = nil
	} else {
		data["voice_ai_monthly_limit"] = *voiceLimit
	}
	if clinicLimit == nil {
		data["clinic_ai_monthly_limit"] = nil
	} else {
		data["clinic_ai_monthly_limit"] = *clinicLimit
	}
	_, err = g.DB().Model("ai_quota_user_override").Ctx(ctx).Data(data).Save()
	if err != nil {
		return contracts.VoiceAIQuotaUserOverrideDTO{}, err
	}
	return GetVoiceAIQuotaUserOverrideForAdmin(ctx, wxID)
}

// ListVoiceAIQuotaUsersForAdmin 分页聚合全部真实 wx 的有效额度与身份。
// Args: page/pageSize 分页；deviceNo 非空时作为 device wx 列表 q（模糊匹配设备号等）。
// Returns: 含 deviceNo/wxId/account/babyName 与 voiceAi/clinicAi 的 used/limit。
// Side Effects: 经 DeviceAdmin HTTP 拉 wx；读本域 override 表与 Redis usage 键；不写库。
func ListVoiceAIQuotaUsersForAdmin(ctx context.Context, page, pageSize int, deviceNo string) (contracts.VoiceAIQuotaUserPageResult, error) {
	wxPage, err := DeviceAdmin().ListWxPage(ctx, page, pageSize, strings.TrimSpace(deviceNo))
	if err != nil {
		return contracts.VoiceAIQuotaUserPageResult{}, err
	}
	empty := contracts.VoiceAIQuotaUserPageResult{
		List:     []contracts.VoiceAIQuotaUserListItem{},
		Total:    wxPage.Total,
		Page:     wxPage.Page,
		PageSize: wxPage.PageSize,
	}
	if len(wxPage.List) == 0 {
		return empty, nil
	}
	def, err := loadVoiceAIQuotaDefault(ctx)
	if err != nil {
		return contracts.VoiceAIQuotaUserPageResult{}, err
	}
	wxIDs := make([]int64, 0, len(wxPage.List))
	for _, it := range wxPage.List {
		if it.Id > 0 {
			wxIDs = append(wxIDs, it.Id)
		}
	}
	ovByWx := make(map[int64]voiceQuotaOverrideRow, len(wxIDs))
	if len(wxIDs) > 0 {
		var ovRows []voiceQuotaOverrideRow
		_ = g.DB().Model("ai_quota_user_override").Ctx(ctx).WhereIn("wx_id", wxIDs).Scan(&ovRows)
		for _, ov := range ovRows {
			ovByWx[ov.WxId] = ov
		}
	}
	// 批量读取当月 usage：每行 voice_ai + clinic_ai 两个键。
	keys := make([]string, 0, len(wxPage.List)*2)
	for _, it := range wxPage.List {
		if it.Id <= 0 {
			continue
		}
		keys = append(keys,
			aiQuotaUsageRedisKey(contracts.AIQuotaVoiceAI, it.Id),
			aiQuotaUsageRedisKey(contracts.AIQuotaClinicAI, it.Id),
		)
	}
	usedByKey := map[string]string{}
	if len(keys) > 0 {
		usedByKey, err = voiceQuotaCache.MGet(ctx, keys)
		if err != nil {
			return contracts.VoiceAIQuotaUserPageResult{}, err
		}
	}
	parseUsed := func(key string) int {
		v := strings.TrimSpace(usedByKey[key])
		if v == "" {
			return 0
		}
		n, _ := strconv.Atoi(v)
		if n < 0 {
			return 0
		}
		return n
	}
	list := make([]contracts.VoiceAIQuotaUserListItem, 0, len(wxPage.List))
	for _, it := range wxPage.List {
		voiceLimit := def.VoiceAiMonthlyLimit
		clinicLimit := def.ClinicAiMonthlyLimit
		if ov, ok := ovByWx[it.Id]; ok {
			if ov.VoiceAiMonthlyLimit != nil && *ov.VoiceAiMonthlyLimit > 0 {
				voiceLimit = *ov.VoiceAiMonthlyLimit
			}
			if ov.ClinicAiMonthlyLimit != nil && *ov.ClinicAiMonthlyLimit > 0 {
				clinicLimit = *ov.ClinicAiMonthlyLimit
			}
		}
		voiceUsed := 0
		clinicUsed := 0
		if it.Id > 0 {
			voiceUsed = parseUsed(aiQuotaUsageRedisKey(contracts.AIQuotaVoiceAI, it.Id))
			clinicUsed = parseUsed(aiQuotaUsageRedisKey(contracts.AIQuotaClinicAI, it.Id))
		}
		list = append(list, contracts.VoiceAIQuotaUserListItem{
			DeviceNo: it.DeviceNo,
			WxId:     it.Id,
			Account:  it.Account,
			BabyName: it.BabyName,
			VoiceAi:  contracts.AIQuotaSnapshot{Used: voiceUsed, Limit: voiceLimit, Allowed: voiceUsed < voiceLimit},
			ClinicAi: contracts.AIQuotaSnapshot{Used: clinicUsed, Limit: clinicLimit, Allowed: clinicUsed < clinicLimit},
		})
	}
	return contracts.VoiceAIQuotaUserPageResult{
		List:     list,
		Total:    wxPage.Total,
		Page:     wxPage.Page,
		PageSize: wxPage.PageSize,
	}, nil
}
