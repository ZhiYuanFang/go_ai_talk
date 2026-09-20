package voice

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	v1 "hello/api/v1"
	"hello/internal/clients/cash"
	"hello/internal/platform/cachekit"
	"hello/internal/services/aimodel"
	"hello/internal/services/contracts"

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
	careAlertDefaultLimit   = 5
)

var careAlertCache = cachekit.Default()

// careAlertFlight 进程内 single-flight：同 wxId+deviceNo 并发生成共用一次。
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

type careAlertLatestRow struct {
	WxId      int64  `orm:"wx_id" json:"wxId"`
	DeviceNo  string `orm:"device_no" json:"deviceNo"`
	Day       string `orm:"day" json:"day"`
	ItemsJson string `orm:"items_json" json:"itemsJson"`
	UpdatedAt int64  `orm:"updated_at" json:"updatedAt"`
}

// requireCareAlertAccess 经 cash internal 校验喂养资格 ∧（开通∨VIP∨试用未用）；失败 fail-closed。
func requireCareAlertAccess(ctx context.Context, deviceNo string, wxID int64) error {
	acc, err := cash.RemoteCareAlertAccess(ctx, deviceNo, wxID)
	if err != nil {
		glog.Warningf(ctx, "[CareAlert] access 调用失败 deviceNoLen=%d wxId=%d err=%v", len(deviceNo), wxID, err)
		return gerror.WrapCode(gcode.CodeOperationFailed, err, "暂时无法校验值得留意开通状态")
	}
	if acc == nil || !acc.Allowed {
		reason := "未满足值得留意查看条件"
		if acc != nil && !acc.FeedingQualified {
			reason = "未满足值得留意喂养资格"
		} else if acc != nil && !acc.FeatureActive && !acc.TrialAvailable {
			reason = "未开通值得留意智能提醒"
		}
		return gerror.NewCode(gcode.CodeInvalidOperation, reason)
	}
	return nil
}

// CareAlertDaily 返回用户对该宝宝的最新护理留意。
// force=false：仅读 latest，不生成、不消耗日额度。
// force=true：校验日限后生成，成功落库 latest + INCR + 可能 claim 试用。
func CareAlertDaily(ctx context.Context, deviceNo string, wxID int64, force bool) (day string, items []v1.CareAlertItemDTO, usedToday, dailyLimit int, err error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return "", nil, 0, 0, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if wxID <= 0 {
		return "", nil, 0, 0, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	if err := requireCareAlertAccess(ctx, deviceNo, wxID); err != nil {
		return "", nil, 0, 0, err
	}
	if err := EnsureVoiceAIQuotaSchema(ctx); err != nil {
		return "", nil, 0, 0, err
	}
	if err := DeviceAdmin().EnsureRegistered(ctx, deviceNo); err != nil {
		return "", nil, 0, 0, err
	}
	usedToday, dailyLimit = careAlertUsageSnapshot(ctx, wxID)
	day = shanghaiDayString(time.Now())

	if !force {
		row, ok, lErr := loadCareAlertLatest(ctx, wxID, deviceNo)
		if lErr != nil {
			return day, nil, usedToday, dailyLimit, lErr
		}
		if !ok {
			return day, []v1.CareAlertItemDTO{}, usedToday, dailyLimit, nil
		}
		if strings.TrimSpace(row.Day) != "" {
			day = row.Day
		}
		parsed, pErr := parseCareAlertItemsJSON(row.ItemsJson)
		if pErr != nil {
			glog.Warningf(ctx, "[CareAlert] latest items 解析失败 wxId=%d err=%v", wxID, pErr)
			return day, []v1.CareAlertItemDTO{}, usedToday, dailyLimit, nil
		}
		return day, parsed, usedToday, dailyLimit, nil
	}

	if err := checkCareAlertDailyLimit(ctx, wxID); err != nil {
		return day, nil, usedToday, dailyLimit, err
	}
	day, items, err = careAlertGenerateSingleFlight(ctx, deviceNo, day, wxID)
	usedToday, dailyLimit = careAlertUsageSnapshot(ctx, wxID)
	return day, items, usedToday, dailyLimit, err
}

// CareAlertDailyStreamCallback 向 SSE 控制器转发 thinking / 终态 result / error。
type CareAlertDailyStreamCallback struct {
	OnThinking func(dataJSON string) error
	OnResult   func(dataJSON string) error // final client-facing result with items+usage
	OnError    func(dataJSON string) error
}

// CareAlertDailyStream 护理留意强制生成 SSE：预检通过后调 Python 流式分析；result 落库后回写用量。
// 预检失败（开通/日限等）返回普通 error，供控制器在开 SSE 头前走 JSON envelope。
func CareAlertDailyStream(ctx context.Context, deviceNo string, wxID int64, cb *CareAlertDailyStreamCallback) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	// 与 CareAlertDaily force 路径相同的门禁。
	if err := requireCareAlertAccess(ctx, deviceNo, wxID); err != nil {
		return err
	}
	if err := EnsureVoiceAIQuotaSchema(ctx); err != nil {
		return err
	}
	if err := DeviceAdmin().EnsureRegistered(ctx, deviceNo); err != nil {
		return err
	}
	if err := checkCareAlertDailyLimit(ctx, wxID); err != nil {
		return err
	}

	day := shanghaiDayString(time.Now())
	lockKey, err := cachekit.CareAlertDailyLockKey(wxID, deviceNo)
	if err != nil {
		return err
	}
	got, lockErr := careAlertCache.SetNXEX(ctx, lockKey, "1", careAlertLockTTL)
	if lockErr != nil {
		glog.Warningf(ctx, "[CareAlert] SSE 加锁失败，退化为本进程生成 day=%s err=%v", day, lockErr)
		got = true
	}
	if !got {
		return gerror.NewCode(gcode.CodeInvalidOperation, "护理留意生成进行中，请稍候")
	}
	defer func() { _ = careAlertCache.Del(context.Background(), lockKey) }()

	genCtx, cancel := context.WithTimeout(ctx, careAlertAnalyzeTimeout)
	defer cancel()

	ent, runtime, modelCfg, _ := ResolveLaneModel(genCtx, wxID, aimodel.LaneCareAlert, contracts.AIQuotaCareAlert, PrivilegeAccount)
	if modelCfg != nil {
		rel, acqErr := aimodel.Acquire(genCtx, runtime)
		if acqErr != nil {
			return gerror.WrapCode(gcode.CodeInternalError, acqErr, "护理留意队列繁忙")
		}
		defer rel()
	}
	ageMonths := careAlertAgeMonths(genCtx, deviceNo)

	pythonClient := PythonAIClientFromCfg()
	streamErr := pythonClient.CareAlertAnalyzeStream(genCtx, &CareAlertAnalyzeRequest{
		DeviceNo:       deviceNo,
		Day:            day,
		Model:          modelCfg,
		AgeMonths:      ageMonths,
		HistorySummary: map[string]interface{}{},
		KgContext:      map[string]interface{}{},
	}, &CareAlertAnalyzeStreamCallback{
		OnThinking: func(dataJSON string) error {
			if cb != nil && cb.OnThinking != nil {
				return cb.OnThinking(dataJSON)
			}
			return nil
		},
		OnResult: func(dataJSON string) error {
			// 解析 Python result（type=result + items），落库并计次后回写客户端终态。
			var payload struct {
				Type  string                 `json:"type"`
				Day   string                 `json:"day"`
				Items []CareAlertAnalyzeItem `json:"items"`
			}
			if err := json.Unmarshal([]byte(dataJSON), &payload); err != nil {
				return gerror.WrapCode(gcode.CodeInternalError, err, "护理留意流式结果解析失败")
			}
			resultDay := strings.TrimSpace(payload.Day)
			if resultDay == "" {
				resultDay = day
			}
			items := normalizeCareAlertItems(payload.Items)
			if err := upsertCareAlertLatest(ctx, wxID, deviceNo, resultDay, items); err != nil {
				return gerror.WrapCode(gcode.CodeInternalError, err, "护理留意结果保存失败")
			}
			if iErr := incrCareAlertDailyUsage(ctx, wxID); iErr != nil {
				glog.Warningf(ctx, "[CareAlert] SSE 日限 INCR 失败 wxId=%d err=%v", wxID, iErr)
			}
			if cErr := cash.RemoteClaimFeatureTrial(ctx, wxID, "care_alert_smart_remind"); cErr != nil {
				glog.Warningf(ctx, "[CareAlert] SSE claim trial 失败 wxId=%d err=%v", wxID, cErr)
			}
			ConsumeVoiceFeatureIfNeeded(ctx, wxID, contracts.AIQuotaCareAlert, ent)
			usedToday, dailyLimit := careAlertUsageSnapshot(ctx, wxID)
			if items == nil {
				items = []v1.CareAlertItemDTO{}
			}
			out, mErr := json.Marshal(map[string]interface{}{
				"type":       "result",
				"day":        resultDay,
				"items":      items,
				"usedToday":  usedToday,
				"dailyLimit": dailyLimit,
			})
			if mErr != nil {
				return gerror.WrapCode(gcode.CodeInternalError, mErr, "护理留意流式结果序列化失败")
			}
			modelName := ""
			modelProvider := ""
			if modelCfg != nil {
				modelName = modelCfg.Name
				modelProvider = modelCfg.Provider
			}
			glog.Infof(ctx, "[CareAlert] SSE generated day=%s count=%d provider=%s model=%s wxId=%d premium=%v vip=%v",
				resultDay, len(items), modelProvider, modelName, wxID, ent.Premium, ent.VIP)
			if cb != nil && cb.OnResult != nil {
				return cb.OnResult(string(out))
			}
			return nil
		},
		OnError: func(dataJSON string) error {
			if cb != nil && cb.OnError != nil {
				return cb.OnError(dataJSON)
			}
			return nil
		},
	})
	if streamErr != nil {
		return gerror.WrapCode(gcode.CodeInternalError, streamErr, "护理留意分析暂时不可用，请稍后再试")
	}
	return nil
}

// CareAlertDeleteItem 从该用户该宝宝 latest 移除 suggestionId；无记录时返回空列表。
func CareAlertDeleteItem(ctx context.Context, deviceNo, suggestionID string, wxID int64) (day string, items []v1.CareAlertItemDTO, err error) {
	deviceNo = strings.TrimSpace(deviceNo)
	suggestionID = strings.TrimSpace(suggestionID)
	if deviceNo == "" {
		return "", nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if suggestionID == "" {
		return "", nil, gerror.NewCode(gcode.CodeInvalidParameter, "suggestionId 不能为空")
	}
	if wxID <= 0 {
		return "", nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	if err := requireCareAlertAccess(ctx, deviceNo, wxID); err != nil {
		return "", nil, err
	}
	if err := EnsureVoiceAIQuotaSchema(ctx); err != nil {
		return "", nil, err
	}
	if err := DeviceAdmin().EnsureRegistered(ctx, deviceNo); err != nil {
		return "", nil, err
	}
	day = shanghaiDayString(time.Now())
	row, ok, cErr := loadCareAlertLatest(ctx, wxID, deviceNo)
	if cErr != nil {
		return day, nil, cErr
	}
	if !ok {
		return day, []v1.CareAlertItemDTO{}, nil
	}
	if strings.TrimSpace(row.Day) != "" {
		day = row.Day
	}
	cached, pErr := parseCareAlertItemsJSON(row.ItemsJson)
	if pErr != nil {
		return day, []v1.CareAlertItemDTO{}, nil
	}
	out := make([]v1.CareAlertItemDTO, 0, len(cached))
	for _, it := range cached {
		if strings.TrimSpace(it.SuggestionId) == suggestionID {
			continue
		}
		out = append(out, it)
	}
	if err := upsertCareAlertLatest(ctx, wxID, deviceNo, day, out); err != nil {
		return day, nil, err
	}
	return day, out, nil
}

// CareAlertFeedback 固定意图飞轮：本地落日志 + 尽力转发 Python；不扣日额度。
func CareAlertFeedback(ctx context.Context, deviceNo, suggestionID, intent string, wxID int64) error {
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
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	if err := requireCareAlertAccess(ctx, deviceNo, wxID); err != nil {
		return err
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
		glog.Warningf(ctx, "[CareAlert] Python 飞轮失败（已本地确认）intent=%s err=%v", intent, pyErr)
	}
	return nil
}

func careAlertGenerateSingleFlight(ctx context.Context, deviceNo, day string, wxID int64) (string, []v1.CareAlertItemDTO, error) {
	flightKey := fmt.Sprintf("%d|%s", wxID, deviceNo)
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

	items, err := careAlertGenerate(ctx, deviceNo, day, wxID)
	w.day = day
	w.items = items
	w.err = err
	return day, items, err
}

func careAlertGenerate(ctx context.Context, deviceNo, day string, wxID int64) ([]v1.CareAlertItemDTO, error) {
	lockKey, err := cachekit.CareAlertDailyLockKey(wxID, deviceNo)
	if err != nil {
		return nil, err
	}
	got, lockErr := careAlertCache.SetNXEX(ctx, lockKey, "1", careAlertLockTTL)
	if lockErr != nil {
		glog.Warningf(ctx, "[CareAlert] 加锁失败，退化为本进程生成 day=%s err=%v", day, lockErr)
		got = true
	}
	if !got {
		return nil, gerror.NewCode(gcode.CodeInvalidOperation, "护理留意生成进行中，请稍候")
	}
	defer func() { _ = careAlertCache.Del(context.Background(), lockKey) }()

	genCtx, cancel := context.WithTimeout(ctx, careAlertAnalyzeTimeout)
	defer cancel()

	ent, runtime, modelCfg, _ := ResolveLaneModel(genCtx, wxID, aimodel.LaneCareAlert, contracts.AIQuotaCareAlert, PrivilegeAccount)
	if modelCfg != nil {
		rel, acqErr := aimodel.Acquire(genCtx, runtime)
		if acqErr != nil {
			return nil, gerror.WrapCode(gcode.CodeInternalError, acqErr, "护理留意队列繁忙")
		}
		defer rel()
	}
	ageMonths := careAlertAgeMonths(genCtx, deviceNo)

	pythonClient := PythonAIClientFromCfg()
	pyRes, pyErr := pythonClient.CareAlertAnalyze(genCtx, &CareAlertAnalyzeRequest{
		DeviceNo:       deviceNo,
		Day:            day,
		Model:          modelCfg,
		AgeMonths:      ageMonths,
		HistorySummary: map[string]interface{}{},
		KgContext:      map[string]interface{}{},
	})
	if pyErr != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, pyErr, "护理留意分析失败")
	}

	items := normalizeCareAlertItems(pyRes.Items)
	if err := upsertCareAlertLatest(ctx, wxID, deviceNo, day, items); err != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "护理留意结果保存失败")
	}
	if iErr := incrCareAlertDailyUsage(ctx, wxID); iErr != nil {
		glog.Warningf(ctx, "[CareAlert] 日限 INCR 失败 wxId=%d err=%v", wxID, iErr)
	}
	// 首次成功落库后 claim 试用（幂等；失败仅告警，结果已落库）。
	if cErr := cash.RemoteClaimFeatureTrial(ctx, wxID, "care_alert_smart_remind"); cErr != nil {
		glog.Warningf(ctx, "[CareAlert] claim trial 失败 wxId=%d err=%v", wxID, cErr)
	}
	ConsumeVoiceFeatureIfNeeded(ctx, wxID, contracts.AIQuotaCareAlert, ent)
	modelName := ""
	modelProvider := ""
	if modelCfg != nil {
		modelName = modelCfg.Name
		modelProvider = modelCfg.Provider
	}
	glog.Infof(ctx, "[CareAlert] generated day=%s count=%d provider=%s model=%s wxId=%d premium=%v vip=%v",
		day, len(items), modelProvider, modelName, wxID, ent.Premium, ent.VIP)
	return items, nil
}

func checkCareAlertDailyLimit(ctx context.Context, wxID int64) error {
	limit, err := GetCareAlertDailyLimit(ctx)
	if err != nil {
		glog.Warningf(ctx, "[CareAlert] 读日限失败，使用默认 %d err=%v", careAlertDefaultLimit, err)
		limit = careAlertDefaultLimit
	}
	used, err := readCareAlertDailyUsage(ctx, wxID)
	if err != nil {
		return gerror.WrapCode(gcode.CodeOperationFailed, err, "暂时无法校验值得留意日限")
	}
	if used >= limit {
		return gerror.NewCode(contracts.GCodeCareAlertDailyLimit(), contracts.ErrCareAlertDailyLimit.Error())
	}
	return nil
}

func careAlertUsageSnapshot(ctx context.Context, wxID int64) (used, limit int) {
	limit, err := GetCareAlertDailyLimit(ctx)
	if err != nil || limit <= 0 {
		limit = careAlertDefaultLimit
	}
	used, uErr := readCareAlertDailyUsage(ctx, wxID)
	if uErr != nil {
		glog.Warningf(ctx, "[CareAlert] 读用量失败 wxId=%d err=%v", wxID, uErr)
		used = 0
	}
	return used, limit
}

func readCareAlertDailyUsage(ctx context.Context, wxID int64) (int, error) {
	day := shanghaiDayCompact(time.Now())
	key, err := cachekit.CareAlertDailyUsageKey(wxID, day)
	if err != nil {
		return 0, err
	}
	raw, ok, err := careAlertCache.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, nil
	}
	n, _ := strconv.Atoi(strings.TrimSpace(raw))
	return n, nil
}

func incrCareAlertDailyUsage(ctx context.Context, wxID int64) error {
	day := shanghaiDayCompact(time.Now())
	key, err := cachekit.CareAlertDailyUsageKey(wxID, day)
	if err != nil {
		return err
	}
	n, err := careAlertCache.Incr(ctx, key)
	if err != nil {
		return err
	}
	if n == 1 {
		_ = careAlertCache.Expire(ctx, key, growthTrajectoryDailyUsageTTL(time.Now()))
	}
	return nil
}

func loadCareAlertLatest(ctx context.Context, wxID int64, deviceNo string) (careAlertLatestRow, bool, error) {
	// 尚无最新记录时空集为正常路径，须 One+IsEmpty，禁止 Scan 将 ErrNoRows 当系统失败。
	one, err := g.DB().Model("care_alert_latest").Ctx(ctx).
		Where("wx_id", wxID).Where("device_no", deviceNo).One()
	if err != nil {
		return careAlertLatestRow{}, false, err
	}
	if one.IsEmpty() {
		return careAlertLatestRow{}, false, nil
	}
	var row careAlertLatestRow
	if err = one.Struct(&row); err != nil {
		return careAlertLatestRow{}, false, err
	}
	return row, true, nil
}

func upsertCareAlertLatest(ctx context.Context, wxID int64, deviceNo, day string, items []v1.CareAlertItemDTO) error {
	if items == nil {
		items = []v1.CareAlertItemDTO{}
	}
	raw, err := json.Marshal(items)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	data := g.Map{
		"wx_id": wxID, "device_no": deviceNo, "day": day,
		"items_json": string(raw), "updated_at": now,
	}
	n, err := g.DB().Model("care_alert_latest").Ctx(ctx).
		Where("wx_id", wxID).Where("device_no", deviceNo).Count()
	if err != nil {
		return err
	}
	if n == 0 {
		_, err = g.DB().Model("care_alert_latest").Ctx(ctx).Data(data).Insert()
		return err
	}
	_, err = g.DB().Model("care_alert_latest").Ctx(ctx).
		Where("wx_id", wxID).Where("device_no", deviceNo).Data(data).Update()
	return err
}

func parseCareAlertItemsJSON(raw string) ([]v1.CareAlertItemDTO, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []v1.CareAlertItemDTO{}, nil
	}
	var items []v1.CareAlertItemDTO
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, err
	}
	if items == nil {
		items = []v1.CareAlertItemDTO{}
	}
	return items, nil
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
