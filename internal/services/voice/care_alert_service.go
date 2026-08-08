package voice

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	v1 "hello/api/v1"
	"hello/internal/platform/cachekit"
	"hello/internal/services/aimodel"
	"hello/internal/services/cash"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	careAlertShanghaiLocName = "Asia/Shanghai"
	// careAlertAnalyzeTimeout 首次生成阻塞上限（Flutter 客户端 90s，服务端略放宽）。
	careAlertAnalyzeTimeout = 100 * time.Second
	careAlertLockTTL        = 120 * time.Second
	careAlertPollInterval   = 500 * time.Millisecond
)

var careAlertCache = cachekit.Default()

// careAlertFlight 进程内 single-flight：同 deviceNo+day 并发 GET 共用一次生成。
var (
	careAlertFlightMu sync.Mutex
	careAlertFlight   = map[string]*careAlertFlightWait{}
)

type careAlertFlightWait struct {
	done  chan struct{}
	day   string
	items []v1.CareAlertItemDTO
	err   error
}

// CareAlertDaily 返回宝宝当日护理留意列表：缓存命中直接返回；未命中 single-flight 调 Python 后写入。
// 不扣 clinic_ai 配额；触发者 VIP（wxId）→DeepSeek，非 VIP / 查失败降级→Zhipu。
// wxID 必须 >0（由 controller 校验）；禁止用 deviceNo 反查 wx 作为登录旁路。
func CareAlertDaily(ctx context.Context, deviceNo string, wxID int64) (day string, items []v1.CareAlertItemDTO, err error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return "", nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if wxID <= 0 {
		return "", nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	if err := DeviceAdmin().EnsureRegistered(ctx, deviceNo); err != nil {
		return "", nil, err
	}
	day = shanghaiDayString(time.Now())
	if cached, ok, cErr := loadCareAlertDailyCache(ctx, deviceNo, day); cErr != nil {
		glog.Warningf(ctx, "[CareAlert] 读缓存失败 deviceNoLen=%d day=%s err=%v", len(deviceNo), day, cErr)
	} else if ok {
		return day, cached, nil
	}
	return careAlertDailySingleFlight(ctx, deviceNo, day, wxID)
}

// CareAlertDeleteItem 仅从当日缓存移除 suggestionId；无缓存时返回空列表。
func CareAlertDeleteItem(ctx context.Context, deviceNo, suggestionID string) (day string, items []v1.CareAlertItemDTO, err error) {
	deviceNo = strings.TrimSpace(deviceNo)
	suggestionID = strings.TrimSpace(suggestionID)
	if deviceNo == "" {
		return "", nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if suggestionID == "" {
		return "", nil, gerror.NewCode(gcode.CodeInvalidParameter, "suggestionId 不能为空")
	}
	if err := DeviceAdmin().EnsureRegistered(ctx, deviceNo); err != nil {
		return "", nil, err
	}
	day = shanghaiDayString(time.Now())
	cached, ok, cErr := loadCareAlertDailyCache(ctx, deviceNo, day)
	if cErr != nil {
		return day, nil, cErr
	}
	if !ok {
		return day, []v1.CareAlertItemDTO{}, nil
	}
	out := make([]v1.CareAlertItemDTO, 0, len(cached))
	for _, it := range cached {
		if strings.TrimSpace(it.SuggestionId) == suggestionID {
			continue
		}
		out = append(out, it)
	}
	if err := storeCareAlertDailyCache(ctx, deviceNo, day, out); err != nil {
		return day, nil, err
	}
	return day, out, nil
}

// CareAlertFeedback 固定意图飞轮：本地落日志 + 尽力转发 Python；不扣 clinic 配额、不做 NLP。
func CareAlertFeedback(ctx context.Context, deviceNo, suggestionID, intent string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	suggestionID = strings.TrimSpace(suggestionID)
	intent = strings.TrimSpace(intent)
	if deviceNo == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if suggestionID == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "suggestionId 不能为空")
	}
	if intent != "ignore" && intent != "follow_up" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "intent 必须为 ignore 或 follow_up")
	}
	if err := DeviceAdmin().EnsureRegistered(ctx, deviceNo); err != nil {
		return err
	}
	day := shanghaiDayString(time.Now())
	g.Log().Infof(ctx, "[CareAlert] feedback local deviceNoLen=%d suggestionIdLen=%d intent=%s day=%s",
		len(deviceNo), len(suggestionID), intent, day)
	pythonClient := PythonAIClientFromCfg()
	if pyErr := pythonClient.CareAlertFeedback(ctx, &CareAlertFeedbackRequest{
		DeviceNo:     deviceNo,
		SuggestionID: suggestionID,
		Intent:       intent,
		Day:          day,
	}); pyErr != nil {
		// 飞轮为旁路：Python 未就绪时不阻断客户端忽略/追问主路径。
		glog.Warningf(ctx, "[CareAlert] Python 飞轮失败（已本地确认）intent=%s err=%v", intent, pyErr)
	}
	return nil
}

func careAlertDailySingleFlight(ctx context.Context, deviceNo, day string, wxID int64) (string, []v1.CareAlertItemDTO, error) {
	flightKey := deviceNo + "|" + day
	careAlertFlightMu.Lock()
	if w, ok := careAlertFlight[flightKey]; ok {
		careAlertFlightMu.Unlock()
		select {
		case <-ctx.Done():
			return day, nil, ctx.Err()
		case <-w.done:
			return w.day, w.items, w.err
		}
	}
	w := &careAlertFlightWait{done: make(chan struct{})}
	careAlertFlight[flightKey] = w
	careAlertFlightMu.Unlock()

	defer func() {
		careAlertFlightMu.Lock()
		delete(careAlertFlight, flightKey)
		close(w.done)
		careAlertFlightMu.Unlock()
	}()

	// 触发者权益：本 flight 首位请求者的 wxId 决定选模。
	items, err := careAlertDailyGenerate(ctx, deviceNo, day, wxID)
	w.day = day
	w.items = items
	w.err = err
	return day, items, err
}

func careAlertDailyGenerate(ctx context.Context, deviceNo, day string, wxID int64) ([]v1.CareAlertItemDTO, error) {
	// 双重检查：等待锁期间可能已有他实例写好缓存。
	if cached, ok, err := loadCareAlertDailyCache(ctx, deviceNo, day); err == nil && ok {
		return cached, nil
	}

	lockKey, err := cachekit.CareAlertDailyLockKey(deviceNo, day)
	if err != nil {
		return nil, err
	}
	got, lockErr := careAlertCache.SetNXEX(ctx, lockKey, "1", careAlertLockTTL)
	if lockErr != nil {
		glog.Warningf(ctx, "[CareAlert] 加锁失败，退化为本进程生成 day=%s err=%v", day, lockErr)
		got = true
	}
	if !got {
		return waitCareAlertDailyCache(ctx, deviceNo, day)
	}
	defer func() { _ = careAlertCache.Del(context.Background(), lockKey) }()

	if cached, ok, err := loadCareAlertDailyCache(ctx, deviceNo, day); err == nil && ok {
		return cached, nil
	}

	genCtx, cancel := context.WithTimeout(ctx, careAlertAnalyzeTimeout)
	defer cancel()

	profile, vip := resolveCareAlertModelProfile(genCtx, wxID)
	ageMonths := careAlertAgeMonths(genCtx, deviceNo)

	pythonClient := PythonAIClientFromCfg()
	pyRes, pyErr := pythonClient.CareAlertAnalyze(genCtx, &CareAlertAnalyzeRequest{
		DeviceNo:       deviceNo,
		Day:            day,
		Model:          string(profile.Provider), // deepseek|zhipu（契约简写）
		ModelCfg:       PythonModelCfg{Provider: string(profile.Provider), Name: profile.Model, MaxInFlight: profile.MaxInFlight},
		AgeMonths:      ageMonths,
		HistorySummary: map[string]interface{}{},
		KgContext:      map[string]interface{}{},
	})
	if pyErr != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, pyErr, "护理留意分析失败")
	}

	items := normalizeCareAlertItems(pyRes.Items)
	if err := storeCareAlertDailyCache(ctx, deviceNo, day, items); err != nil {
		glog.Warningf(ctx, "[CareAlert] 写缓存失败 day=%s err=%v", day, err)
		// 仍返回本次结果，避免客户端因缓存故障完全不可用。
	}
	glog.Infof(ctx, "[CareAlert] generated day=%s count=%d provider=%s model=%s wxId=%d vip=%v",
		day, len(items), profile.Provider, profile.Model, wxID, vip)
	return items, nil
}

func waitCareAlertDailyCache(ctx context.Context, deviceNo, day string) ([]v1.CareAlertItemDTO, error) {
	deadline := time.Now().Add(careAlertAnalyzeTimeout)
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if cached, ok, err := loadCareAlertDailyCache(ctx, deviceNo, day); err == nil && ok {
			return cached, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(careAlertPollInterval):
		}
	}
	return nil, gerror.NewCode(gcode.CodeInternalError, "护理留意生成超时，请稍后重试")
}

func loadCareAlertDailyCache(ctx context.Context, deviceNo, day string) ([]v1.CareAlertItemDTO, bool, error) {
	key, err := cachekit.CareAlertDailyKey(deviceNo, day)
	if err != nil {
		return nil, false, err
	}
	raw, ok, err := careAlertCache.Get(ctx, key)
	if err != nil || !ok {
		return nil, false, err
	}
	var payload careAlertCachePayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, false, err
	}
	if payload.Items == nil {
		payload.Items = []v1.CareAlertItemDTO{}
	}
	return payload.Items, true, nil
}

func storeCareAlertDailyCache(ctx context.Context, deviceNo, day string, items []v1.CareAlertItemDTO) error {
	key, err := cachekit.CareAlertDailyKey(deviceNo, day)
	if err != nil {
		return err
	}
	if items == nil {
		items = []v1.CareAlertItemDTO{}
	}
	raw, err := json.Marshal(careAlertCachePayload{Day: day, Items: items})
	if err != nil {
		return err
	}
	return careAlertCache.SetEX(ctx, key, string(raw), careAlertCacheTTL(time.Now()))
}

type careAlertCachePayload struct {
	Day   string               `json:"day"`
	Items []v1.CareAlertItemDTO `json:"items"`
}

func careAlertCacheTTL(now time.Time) time.Duration {
	loc := shanghaiLocation()
	now = now.In(loc)
	// 次日 01:00 过期，覆盖跨午夜多看护读；最短 1h 避免时钟边界瞬间过期。
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 1, 0, 0, 0, loc)
	d := next.Sub(now)
	if d < time.Hour {
		return time.Hour
	}
	return d
}

func shanghaiLocation() *time.Location {
	loc, err := time.LoadLocation(careAlertShanghaiLocName)
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}

func shanghaiDayString(t time.Time) string {
	return t.In(shanghaiLocation()).Format("2006-01-02")
}

// isAccountVIP 经 cash-service 内部契约按 wxId 读账号 VIP；失败/超时降级为 false（Zhipu）并打 Warning。
func isAccountVIP(ctx context.Context, wxID int64) bool {
	if wxID <= 0 {
		return false
	}
	vip, err := cash.RemoteIsVipByWxID(ctx, wxID)
	if err != nil {
		glog.Warningf(ctx, "[CareAlert] VIP 查询降级为非 VIP wxId=%d err=%v", wxID, err)
		return false
	}
	return vip
}

// resolveCareAlertModelProfile 触发者 VIP→DeepSeek，否则 Zhipu；复用 aimodel 闸门字段，不耦合 clinic 配额。
// 返回值 vip 为实际用于选模的布尔（含降级后的 false）。
func resolveCareAlertModelProfile(ctx context.Context, wxID int64) (aimodel.Profile, bool) {
	base, err := aimodel.LoadProfile(ctx, aimodel.LaneClinic)
	if err != nil {
		base = aimodel.DefaultSeedProfile(aimodel.LaneClinic)
	}
	vip := isAccountVIP(ctx, wxID)
	if vip {
		base.Provider = aimodel.ProviderDeepSeek
		if !aimodel.IsAllowedModel(aimodel.ProviderDeepSeek, base.Model) {
			base.Model = "deepseek-v4-flash"
		}
	} else {
		base.Provider = aimodel.ProviderZhipu
		if !aimodel.IsAllowedModel(aimodel.ProviderZhipu, base.Model) {
			base.Model = "glm-4.7-flash"
		}
	}
	if base.TimeoutSec < 90 {
		base.TimeoutSec = 90
	}
	return base, vip
}

func careAlertAgeMonths(ctx context.Context, deviceNo string) int {
	profile, err := DeviceProfile().GetProfile(ctx, deviceNo)
	if err != nil || profile.Birthday <= 0 {
		return 0
	}
	return ageMonthsFromBirthdayUnix(profile.Birthday, time.Now())
}

func ageMonthsFromBirthdayUnix(birthdayUnix int64, now time.Time) int {
	b := time.Unix(birthdayUnix, 0).In(shanghaiLocation())
	now = now.In(shanghaiLocation())
	months := (now.Year()-b.Year())*12 + int(now.Month()) - int(b.Month())
	if now.Day() < b.Day() {
		months--
	}
	if months < 0 {
		return 0
	}
	return months
}

func normalizeCareAlertItems(raw []CareAlertAnalyzeItem) []v1.CareAlertItemDTO {
	out := make([]v1.CareAlertItemDTO, 0, len(raw))
	for _, it := range raw {
		eventID := strings.TrimSpace(it.EventID)
		if eventID == "" {
			continue
		}
		sid := strings.TrimSpace(it.SuggestionID)
		if sid == "" {
			sid = newSuggestionUUID()
		}
		name := strings.TrimSpace(it.EventName)
		summary := strings.TrimSpace(it.SummaryLine)
		prompt := strings.TrimSpace(it.FollowUpPrompt)
		if prompt == "" {
			if summary != "" {
				prompt = summary
			} else if name != "" {
				prompt = fmt.Sprintf("关于%s，我想了解一下最近是否需要留意什么？", name)
			} else {
				prompt = "关于宝宝最近的护理情况，我想了解一下是否需要留意什么？"
			}
		}
		reasons := make([]v1.CareAlertReasonDTO, 0, len(it.Reasons))
		for _, r := range it.Reasons {
			reasons = append(reasons, v1.CareAlertReasonDTO{
				Type:            strings.TrimSpace(r.Type),
				Score:           r.Score,
				ExpectationUsed: r.ExpectationUsed,
				AgeMonths:       r.AgeMonths,
				MedianGapMs:     r.MedianGapMs,
				LastGapMs:       r.LastGapMs,
				ExpectGapMaxMs:  r.ExpectGapMaxMs,
				P75DurMs:        r.P75DurMs,
				ElapsedMs:       r.ElapsedMs,
				ExpectDurMaxMs:  r.ExpectDurMaxMs,
				DailyAvg:        r.DailyAvg,
				Recent48hCount:  r.Recent48hCount,
				StillExpected:   r.StillExpected,
				DetailLines:     r.DetailLines,
			})
		}
		if reasons == nil {
			reasons = []v1.CareAlertReasonDTO{}
		}
		out = append(out, v1.CareAlertItemDTO{
			SuggestionId:   sid,
			EventId:        eventID,
			EventName:      name,
			SummaryLine:    summary,
			FollowUpPrompt: prompt,
			Reasons:        reasons,
		})
	}
	return out
}

func newSuggestionUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("ca-%d", time.Now().UnixNano())
	}
	// RFC 4122 version 4
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16]),
	)
}
