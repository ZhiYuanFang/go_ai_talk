package device

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	aiQuotaDefaultSingletonID = 1
	aiQuotaUsageKeyPrefix     = "ai:usage:"
	aiQuotaUsageTTLSeconds    = 90 * 24 * 3600
)

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

func aiQuotaUsageRedisKey(feature AIQuotaFeature, wxID int64) string {
	return fmt.Sprintf("%s%s:%d:%s", aiQuotaUsageKeyPrefix, feature, wxID, aiQuotaMonthBucket())
}

func validateAIQuotaFeature(feature AIQuotaFeature) error {
	switch feature {
	case AIQuotaPolish, AIQuotaVoiceAI:
		return nil
	default:
		return fmt.Errorf("未知 feature: %s", feature)
	}
}

func validateWxIDForAI(wxID int64) error {
	if wxID <= 0 {
		return ErrAINotLoggedIn
	}
	return nil
}

// EnsureAIQuotaDefaultRow 保证 singleton 行存在（yaml 默认 5/5）。
func EnsureAIQuotaDefaultRow(ctx context.Context) error {
	cols := dao.AiQuotaDefault.Columns()
	var row entity.AiQuotaDefault
	err := dao.AiQuotaDefault.Ctx(ctx).Where(cols.Id, aiQuotaDefaultSingletonID).Scan(&row)
	if err == nil && row.Id == aiQuotaDefaultSingletonID {
		return nil
	}
	now := time.Now().Unix()
	_, err = dao.AiQuotaDefault.Ctx(ctx).Data(g.Map{
		cols.Id:                  aiQuotaDefaultSingletonID,
		cols.PolishMonthlyLimit:  5,
		cols.VoiceAiMonthlyLimit: 5,
		cols.UpdatedAt:           now,
	}).Save()
	return err
}

func loadAIQuotaDefault(ctx context.Context) (entity.AiQuotaDefault, error) {
	if err := EnsureAIQuotaDefaultRow(ctx); err != nil {
		return entity.AiQuotaDefault{}, err
	}
	var row entity.AiQuotaDefault
	cols := dao.AiQuotaDefault.Columns()
	if err := dao.AiQuotaDefault.Ctx(ctx).Where(cols.Id, aiQuotaDefaultSingletonID).Scan(&row); err != nil {
		return entity.AiQuotaDefault{}, err
	}
	if row.PolishMonthlyLimit <= 0 {
		row.PolishMonthlyLimit = 5
	}
	if row.VoiceAiMonthlyLimit <= 0 {
		row.VoiceAiMonthlyLimit = 5
	}
	return row, nil
}

func effectiveLimitForFeature(ctx context.Context, wxID int64, feature AIQuotaFeature) (int, error) {
	def, err := loadAIQuotaDefault(ctx)
	if err != nil {
		return 0, err
	}
	limit := def.PolishMonthlyLimit
	if feature == AIQuotaVoiceAI {
		limit = def.VoiceAiMonthlyLimit
	}
	var ov entity.AiQuotaUserOverride
	cols := dao.AiQuotaUserOverride.Columns()
	_ = dao.AiQuotaUserOverride.Ctx(ctx).Where(cols.WxId, wxID).Scan(&ov)
	if ov.WxId == wxID {
		if feature == AIQuotaPolish && ov.PolishMonthlyLimit != nil && *ov.PolishMonthlyLimit > 0 {
			limit = *ov.PolishMonthlyLimit
		}
		if feature == AIQuotaVoiceAI && ov.VoiceAiMonthlyLimit != nil && *ov.VoiceAiMonthlyLimit > 0 {
			limit = *ov.VoiceAiMonthlyLimit
		}
	}
	return limit, nil
}

func readUsageCount(ctx context.Context, feature AIQuotaFeature, wxID int64) (int, error) {
	key := aiQuotaUsageRedisKey(feature, wxID)
	v, err := g.Redis().Do(ctx, "GET", key)
	if err != nil {
		return 0, err
	}
	if v.IsNil() || v.IsEmpty() {
		return 0, nil
	}
	return v.Int(), nil
}

func touchUsageKeyTTL(ctx context.Context, key string) {
	_, _ = g.Redis().Do(ctx, "EXPIRE", key, aiQuotaUsageTTLSeconds)
}

// CheckAIQuota 只读预检，不修改用量。
func CheckAIQuota(ctx context.Context, wxID int64, feature AIQuotaFeature) (AIQuotaSnapshot, error) {
	if err := validateWxIDForAI(wxID); err != nil {
		return AIQuotaSnapshot{}, err
	}
	if err := validateAIQuotaFeature(feature); err != nil {
		return AIQuotaSnapshot{}, err
	}
	limit, err := effectiveLimitForFeature(ctx, wxID, feature)
	if err != nil {
		return AIQuotaSnapshot{}, err
	}
	used, err := readUsageCount(ctx, feature, wxID)
	if err != nil {
		return AIQuotaSnapshot{}, err
	}
	return AIQuotaSnapshot{
		Used:    used,
		Limit:   limit,
		Allowed: used < limit,
	}, nil
}

// ConsumeAIQuota AI 成功返回后扣减；超额时回滚 INCR 并返回 ErrAIQuotaExhausted。
func ConsumeAIQuota(ctx context.Context, wxID int64, feature AIQuotaFeature) (AIQuotaSnapshot, error) {
	if err := validateWxIDForAI(wxID); err != nil {
		return AIQuotaSnapshot{}, err
	}
	if err := validateAIQuotaFeature(feature); err != nil {
		return AIQuotaSnapshot{}, err
	}
	limit, err := effectiveLimitForFeature(ctx, wxID, feature)
	if err != nil {
		return AIQuotaSnapshot{}, err
	}
	key := aiQuotaUsageRedisKey(feature, wxID)
	n, err := g.Redis().Do(ctx, "INCR", key)
	if err != nil {
		return AIQuotaSnapshot{}, err
	}
	touchUsageKeyTTL(ctx, key)
	used := n.Int()
	if used > limit {
		_, _ = g.Redis().Do(ctx, "DECR", key)
		return AIQuotaSnapshot{Used: limit, Limit: limit, Allowed: false}, ErrAIQuotaExhausted
	}
	return AIQuotaSnapshot{Used: used, Limit: limit, Allowed: true}, nil
}

// GetAIQuotaAppStatus 返回当前用户两 feature 快照。
func GetAIQuotaAppStatus(ctx context.Context, wxID int64) (AIQuotaAppStatus, error) {
	if err := validateWxIDForAI(wxID); err != nil {
		return AIQuotaAppStatus{}, err
	}
	polish, err := CheckAIQuota(ctx, wxID, AIQuotaPolish)
	if err != nil {
		return AIQuotaAppStatus{}, err
	}
	voice, err := CheckAIQuota(ctx, wxID, AIQuotaVoiceAI)
	if err != nil {
		return AIQuotaAppStatus{}, err
	}
	return AIQuotaAppStatus{Polish: polish, VoiceAi: voice}, nil
}

// GetAIQuotaDefaultForAdmin 读取全局默认。
func GetAIQuotaDefaultForAdmin(ctx context.Context) (AIQuotaDefaultDTO, error) {
	row, err := loadAIQuotaDefault(ctx)
	if err != nil {
		return AIQuotaDefaultDTO{}, err
	}
	return AIQuotaDefaultDTO{
		PolishMonthlyLimit:  row.PolishMonthlyLimit,
		VoiceAiMonthlyLimit: row.VoiceAiMonthlyLimit,
		UpdatedAt:           row.UpdatedAt,
	}, nil
}

// UpdateAIQuotaDefaultForAdmin 更新全局默认。
func UpdateAIQuotaDefaultForAdmin(ctx context.Context, polishLimit, voiceLimit int) (AIQuotaDefaultDTO, error) {
	if polishLimit <= 0 || voiceLimit <= 0 {
		return AIQuotaDefaultDTO{}, errors.New("额度须为正整数")
	}
	if err := EnsureAIQuotaDefaultRow(ctx); err != nil {
		return AIQuotaDefaultDTO{}, err
	}
	now := time.Now().Unix()
	cols := dao.AiQuotaDefault.Columns()
	_, err := dao.AiQuotaDefault.Ctx(ctx).Where(cols.Id, aiQuotaDefaultSingletonID).Data(g.Map{
		cols.PolishMonthlyLimit:  polishLimit,
		cols.VoiceAiMonthlyLimit: voiceLimit,
		cols.UpdatedAt:           now,
	}).Update()
	if err != nil {
		return AIQuotaDefaultDTO{}, err
	}
	return AIQuotaDefaultDTO{
		PolishMonthlyLimit:  polishLimit,
		VoiceAiMonthlyLimit: voiceLimit,
		UpdatedAt:           now,
	}, nil
}

// GetAIQuotaUserOverrideForAdmin 读取单人 override。
func GetAIQuotaUserOverrideForAdmin(ctx context.Context, wxID int64) (AIQuotaUserOverrideDTO, error) {
	if wxID <= 0 {
		return AIQuotaUserOverrideDTO{}, errors.New("wxId 无效")
	}
	var row entity.AiQuotaUserOverride
	cols := dao.AiQuotaUserOverride.Columns()
	err := dao.AiQuotaUserOverride.Ctx(ctx).Where(cols.WxId, wxID).Scan(&row)
	if err != nil {
		return AIQuotaUserOverrideDTO{}, err
	}
	if row.WxId != wxID {
		return AIQuotaUserOverrideDTO{WxId: wxID}, nil
	}
	return AIQuotaUserOverrideDTO{
		WxId:                row.WxId,
		PolishMonthlyLimit:  row.PolishMonthlyLimit,
		VoiceAiMonthlyLimit: row.VoiceAiMonthlyLimit,
		UpdatedAt:           row.UpdatedAt,
	}, nil
}

// UpdateAIQuotaUserOverrideForAdmin 写入或清除单人 override（nil 表示清除该 feature override）。
func UpdateAIQuotaUserOverrideForAdmin(ctx context.Context, wxID int64, polishLimit, voiceLimit *int) (AIQuotaUserOverrideDTO, error) {
	if wxID <= 0 {
		return AIQuotaUserOverrideDTO{}, errors.New("wxId 无效")
	}
	if polishLimit != nil && *polishLimit <= 0 {
		return AIQuotaUserOverrideDTO{}, errors.New("polishMonthlyLimit 须为正整数")
	}
	if voiceLimit != nil && *voiceLimit <= 0 {
		return AIQuotaUserOverrideDTO{}, errors.New("voiceAiMonthlyLimit 须为正整数")
	}
	now := time.Now().Unix()
	cols := dao.AiQuotaUserOverride.Columns()
	data := g.Map{
		cols.WxId:      wxID,
		cols.UpdatedAt: now,
	}
	if polishLimit == nil {
		data[cols.PolishMonthlyLimit] = nil
	} else {
		data[cols.PolishMonthlyLimit] = *polishLimit
	}
	if voiceLimit == nil {
		data[cols.VoiceAiMonthlyLimit] = nil
	} else {
		data[cols.VoiceAiMonthlyLimit] = *voiceLimit
	}
	_, err := dao.AiQuotaUserOverride.Ctx(ctx).Data(data).Save()
	if err != nil {
		return AIQuotaUserOverrideDTO{}, err
	}
	return GetAIQuotaUserOverrideForAdmin(ctx, wxID)
}

// WxIDByDeviceNo 由 device_no 反查 wx 主键；未绑定返回 0。
func WxIDByDeviceNo(ctx context.Context, deviceNo string) (int64, error) {
	deviceNo = trimDeviceNo(deviceNo)
	if deviceNo == "" {
		return 0, nil
	}
	one, err := dao.Wx.Ctx(ctx).Where(dao.Wx.Columns().DeviceNo, deviceNo).Limit(1).One()
	if err != nil {
		return 0, err
	}
	if one.IsEmpty() {
		return 0, nil
	}
	return one["id"].Int64(), nil
}

func trimDeviceNo(deviceNo string) string {
	return strings.TrimSpace(deviceNo)
}
