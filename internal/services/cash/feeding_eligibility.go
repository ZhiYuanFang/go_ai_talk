package cash

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// 喂养资格场景键（与客户端约定，Admin 不可新建）。
const (
	SceneKeyUCGEntry       = "ucg_entry"
	SceneKeyCareAlertEntry = "care_alert_entry"
)

// FeedingEligibilityScene 场景阈值行。
type FeedingEligibilityScene struct {
	SceneKey         string `json:"sceneKey"`
	RequiredDays     int    `json:"requiredDays"`
	MinRecordsPerDay int    `json:"minRecordsPerDay"`
	UpdatedAt        int64  `json:"updatedAt"`
}

// FeedingEligibilityResult 场景资格结果（UCG / 值得留意同构）。
type FeedingEligibilityResult struct {
	Qualified     bool   `json:"qualified"`
	RequiredDays  int    `json:"requiredDays"`
	EffectiveDays int    `json:"effectiveDays"`
	RemainingDays int    `json:"remainingDays"`
	Message       string `json:"message,omitempty"`
}

// UCGEligibilityResult 兼容旧名。
type UCGEligibilityResult = FeedingEligibilityResult

// GetFeedingEligibilityScene 读场景配置；缺失时回落安全默认。
func GetFeedingEligibilityScene(ctx context.Context, sceneKey string) (FeedingEligibilityScene, error) {
	sceneKey = strings.TrimSpace(sceneKey)
	def := defaultScene(sceneKey)
	if sceneKey == "" {
		return def, gerror.NewCode(gcode.CodeInvalidParameter, "sceneKey 不能为空")
	}
	c := cachekit.Default()
	cfgKey := cachekit.CashFeedingEligibilitySceneKey(sceneKey)
	if raw, ok, err := c.Get(ctx, cfgKey); err == nil && ok && raw != "" {
		var sc FeedingEligibilityScene
		if json.Unmarshal([]byte(raw), &sc) == nil && sc.SceneKey != "" {
			return normalizeScene(sc), nil
		}
	}
	var row struct {
		SceneKey         string `json:"scene_key"`
		RequiredDays     int    `json:"required_days"`
		MinRecordsPerDay int    `json:"min_records_per_day"`
		UpdatedAt        int64  `json:"updated_at"`
	}
	err := g.DB().Model("feeding_eligibility_scene").Ctx(ctx).Where("scene_key", sceneKey).Scan(&row)
	if err != nil {
		return def, err
	}
	if row.SceneKey == "" {
		return def, nil
	}
	sc := FeedingEligibilityScene{
		SceneKey: row.SceneKey, RequiredDays: row.RequiredDays,
		MinRecordsPerDay: row.MinRecordsPerDay, UpdatedAt: row.UpdatedAt,
	}
	sc = normalizeScene(sc)
	if b, mErr := json.Marshal(sc); mErr == nil {
		_ = c.SetEX(ctx, cfgKey, string(b), 10*time.Minute)
	}
	return sc, nil
}

func defaultScene(sceneKey string) FeedingEligibilityScene {
	switch sceneKey {
	case SceneKeyCareAlertEntry:
		return FeedingEligibilityScene{SceneKey: sceneKey, RequiredDays: 2, MinRecordsPerDay: 10}
	default:
		return FeedingEligibilityScene{SceneKey: sceneKey, RequiredDays: 7, MinRecordsPerDay: 10}
	}
}

func normalizeScene(sc FeedingEligibilityScene) FeedingEligibilityScene {
	if sc.RequiredDays <= 0 {
		sc.RequiredDays = 1
	}
	if sc.MinRecordsPerDay <= 0 {
		sc.MinRecordsPerDay = 1
	}
	return sc
}

// CountConsecutiveEffectiveDays 从昨天起向前连续有效日数（days[0]=昨天；今日不在序列中）。
//
// 业务：独立于场景；门槛由调用方传入；遇无效日中断。
func CountConsecutiveEffectiveDays(days []historyFeedingDayCount, minRecordsPerDay int) int {
	if minRecordsPerDay <= 0 {
		minRecordsPerDay = 1
	}
	n := 0
	for _, d := range days {
		if d.Count >= minRecordsPerDay {
			n++
			continue
		}
		break
	}
	return n
}

// SynthesizeFeedingEligibility 按场景阈值合成资格（不拉数、不缓存）。
//
// days 契约：由 history feeding-day-stats 提供，days[0]=上海昨天，向过去，不含今日。
// message 为激励文案，非客户端进度数字权威（进度由 effectiveDays 等字段拼）。
func SynthesizeFeedingEligibility(scene FeedingEligibilityScene, days []historyFeedingDayCount, qualifiedMsg, pendingMsg string) *FeedingEligibilityResult {
	scene = normalizeScene(scene)
	effective := CountConsecutiveEffectiveDays(days, scene.MinRecordsPerDay)
	remaining := scene.RequiredDays - effective
	if remaining < 0 {
		remaining = 0
	}
	out := &FeedingEligibilityResult{
		Qualified:     effective >= scene.RequiredDays,
		RequiredDays:  scene.RequiredDays,
		EffectiveDays: effective,
		RemainingDays: remaining,
	}
	if out.Qualified {
		out.Message = qualifiedMsg
	} else {
		out.Message = pendingMsg
	}
	return out
}

// GetFeedingEligibilityByScene 读配置 → 拉 history（窗口=requiredDays 个已闭合日）→ 合成 → 按请求日缓存。
//
// 宿主：仅 cash-service。history 返回 days[0]=昨天；今日不计入 streak。
// 缓存键含上海「请求日」：跨日 0 点后新键 miss 重算，已闭合满额即可立即合格（无 ticker）。
func GetFeedingEligibilityByScene(ctx context.Context, deviceNo, sceneKey string) (*FeedingEligibilityResult, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	sceneKey = strings.TrimSpace(sceneKey)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少设备号")
	}
	scene, err := GetFeedingEligibilityScene(ctx, sceneKey)
	if err != nil {
		return nil, err
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	dayKey := time.Now().In(loc).Format("20060102")
	ver := scene.UpdatedAt
	cacheKey := cachekit.CashFeedingEligibilityKey(sceneKey, deviceNo, dayKey, ver)
	c := cachekit.Default()
	if raw, ok, gErr := c.Get(ctx, cacheKey); gErr == nil && ok && raw != "" {
		var cached FeedingEligibilityResult
		if json.Unmarshal([]byte(raw), &cached) == nil {
			return &cached, nil
		}
	}

	stats, err := FetchFeedingDayStats(ctx, deviceNo, scene.RequiredDays)
	if err != nil {
		return nil, err
	}
	qualifiedMsg := "已满足连续有效喂养日"
	pendingMsg := "继续保持每日有效喂养记录"
	switch sceneKey {
	case SceneKeyUCGEntry:
		// 还需要连续喂养x天才能解锁广场,拥有个人邀请码
		pendingMsg = "继续保持每日有效喂养记录以解锁广场,拥有个人邀请码"
	case SceneKeyCareAlertEntry:
		qualifiedMsg = "已满足值得留意喂养资格"
		pendingMsg = "继续累计有效喂养日以激活值得留意"
	}
	out := SynthesizeFeedingEligibility(scene, stats.Days, qualifiedMsg, pendingMsg)
	if b, mErr := json.Marshal(out); mErr == nil {
		_ = c.SetEX(ctx, cacheKey, string(b), 36*time.Hour)
	}
	return out, nil
}

// GetUCGEligibility UCG 入场资格（场景 ucg_entry）。
func GetUCGEligibility(ctx context.Context, deviceNo string) (*FeedingEligibilityResult, error) {
	return GetFeedingEligibilityByScene(ctx, deviceNo, SceneKeyUCGEntry)
}

// GetCareAlertFeedingEligibility 值得留意喂养资格（场景 care_alert_entry）。
func GetCareAlertFeedingEligibility(ctx context.Context, deviceNo string) (*FeedingEligibilityResult, error) {
	return GetFeedingEligibilityByScene(ctx, deviceNo, SceneKeyCareAlertEntry)
}

// AdminListFeedingEligibilityScenes 管理端场景列表。
func AdminListFeedingEligibilityScenes(ctx context.Context) ([]FeedingEligibilityScene, error) {
	var rows []struct {
		SceneKey         string `json:"scene_key"`
		RequiredDays     int    `json:"required_days"`
		MinRecordsPerDay int    `json:"min_records_per_day"`
		UpdatedAt        int64  `json:"updated_at"`
	}
	err := g.DB().Model("feeding_eligibility_scene").Ctx(ctx).OrderAsc("scene_key").Scan(&rows)
	if err != nil {
		return nil, err
	}
	out := make([]FeedingEligibilityScene, 0, len(rows))
	for _, r := range rows {
		out = append(out, normalizeScene(FeedingEligibilityScene{
			SceneKey: r.SceneKey, RequiredDays: r.RequiredDays,
			MinRecordsPerDay: r.MinRecordsPerDay, UpdatedAt: r.UpdatedAt,
		}))
	}
	return out, nil
}

// AdminUpdateFeedingEligibilityScene 更新已有场景阈值；禁止新建未知 scene_key。
func AdminUpdateFeedingEligibilityScene(ctx context.Context, sceneKey string, requiredDays, minRecordsPerDay int) error {
	sceneKey = strings.TrimSpace(sceneKey)
	if sceneKey != SceneKeyUCGEntry && sceneKey != SceneKeyCareAlertEntry {
		return gerror.NewCode(gcode.CodeInvalidParameter, "未知场景（须与客户端约定）")
	}
	if requiredDays <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "requiredDays 须为正整数")
	}
	if minRecordsPerDay <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "minRecordsPerDay 须为正整数")
	}
	exist, err := g.DB().Model("feeding_eligibility_scene").Ctx(ctx).Where("scene_key", sceneKey).One()
	if err != nil {
		return err
	}
	if exist.IsEmpty() {
		return gerror.NewCode(gcode.CodeInvalidParameter, "场景不存在")
	}
	now := time.Now().Unix()
	_, err = g.DB().Model("feeding_eligibility_scene").Ctx(ctx).Where("scene_key", sceneKey).Data(g.Map{
		"required_days": requiredDays, "min_records_per_day": minRecordsPerDay, "updated_at": now,
	}).Update()
	if err != nil {
		return err
	}
	invalidateFeedingEligibilityCaches(ctx, sceneKey)
	return nil
}

func invalidateFeedingEligibilityCaches(ctx context.Context, sceneKey string) {
	c := cachekit.Default()
	_ = c.Del(ctx, cachekit.CashFeedingEligibilitySceneKey(sceneKey))
	// 资格结果键含 updated_at，改配置后旧键自然 miss；无需扫设备键。
}
