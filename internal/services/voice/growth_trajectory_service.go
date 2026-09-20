package voice

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	v1 "hello/api/v1"
	"hello/internal/clients/cash"
	"hello/internal/platform/cachekit"
	"hello/internal/services/contracts"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	growthTrajectoryHorizonDays  = 7
	growthTrajectoryTurnTimeout  = 120 * time.Second
	growthTrajectoryDefaultLimit = 5
)

var growthTrajectoryCache = cachekit.Default()

// 进程内 single-flight：同 wxId 并发 turn 共用/互斥，防连点。
var (
	growthTrajectoryFlightMu sync.Mutex
	growthTrajectoryFlight   = map[int64]chan struct{}{}
)

// GrowthTrajectoryTurnCallback 向 SSE 控制器转发 Python 事件。
type GrowthTrajectoryTurnCallback struct {
	OnThinking func(dataJSON string) error
	OnQuestion func(dataJSON string) error
	OnResult   func(dataJSON string) error
	OnError    func(dataJSON string) error
	OnDone     func(dataJSON string) error
}

type growthTrajectoryLatestRow struct {
	WxId           int64  `orm:"wx_id" json:"wxId"`
	DeviceNo       string `orm:"device_no" json:"deviceNo"`
	ResultMarkdown string `orm:"result_markdown" json:"resultMarkdown"`
	FeedbackJson   string `orm:"feedback_json" json:"feedbackJson"`
	SessionId      string `orm:"session_id" json:"sessionId"`
	UpdatedAt      int64  `orm:"updated_at" json:"updatedAt"`
}

// GrowthTrajectoryLatest 返回该用户该宝宝最新成长轨迹 Markdown（须登录；免开通校验）。
func GrowthTrajectoryLatest(ctx context.Context, deviceNo string, wxID int64) (*v1.DeviceGrowthTrajectoryLatestRes, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	if err := EnsureVoiceAIQuotaSchema(ctx); err != nil {
		return nil, err
	}
	if err := DeviceAdmin().EnsureRegistered(ctx, deviceNo); err != nil {
		return nil, err
	}
	row, ok, err := loadGrowthTrajectoryLatest(ctx, wxID, deviceNo)
	if err != nil {
		return nil, err
	}
	out := &v1.DeviceGrowthTrajectoryLatestRes{SessionId: "", UpdatedAt: 0}
	used, limit := growthTrajectoryUsageSnapshot(ctx, wxID)
	out.UsedToday = used
	out.DailyLimit = limit
	if !ok {
		return out, nil
	}
	out.SessionId = row.SessionId
	out.UpdatedAt = row.UpdatedAt
	md := strings.TrimSpace(row.ResultMarkdown)
	if md != "" {
		out.ResultMarkdown = &md
	}
	return out, nil
}

// GrowthTrajectoryTurn 开流前校验开通与日限，再 single-flight 调 Python 透传 SSE；result 落库后 INCR + claim。
func GrowthTrajectoryTurn(ctx context.Context, deviceNo, action, sessionID string, answer *v1.DeviceGrowthTrajectoryAnswerDTO, wxID int64, cb *GrowthTrajectoryTurnCallback) error {
	deviceNo = strings.TrimSpace(deviceNo)
	action = strings.TrimSpace(strings.ToLower(action))
	sessionID = strings.TrimSpace(sessionID)
	if deviceNo == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	if action != "start" && action != "answer" && action != "restart" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "action 必须为 start|answer|restart")
	}
	if action == "answer" {
		if sessionID == "" {
			return gerror.NewCode(gcode.CodeInvalidParameter, "answer 须带 sessionId")
		}
		if answer == nil || strings.TrimSpace(answer.QuestionId) == "" {
			return gerror.NewCode(gcode.CodeInvalidParameter, "answer.questionId 不能为空")
		}
	}
	if err := requireGrowthTrajectoryAccess(ctx, deviceNo, wxID); err != nil {
		return err
	}
	if err := EnsureVoiceAIQuotaSchema(ctx); err != nil {
		return err
	}
	if err := DeviceAdmin().EnsureRegistered(ctx, deviceNo); err != nil {
		return err
	}
	// 开流前日限预检（问答不计次；已达上限则拒绝任何新流）；日限跟人。
	if err := checkGrowthTrajectoryDailyLimit(ctx, wxID); err != nil {
		return err
	}

	growthTrajectoryFlightMu.Lock()
	if _, busy := growthTrajectoryFlight[wxID]; busy {
		growthTrajectoryFlightMu.Unlock()
		return gerror.NewCode(gcode.CodeInvalidOperation, "成长轨迹预测进行中，请稍候")
	}
	done := make(chan struct{})
	growthTrajectoryFlight[wxID] = done
	growthTrajectoryFlightMu.Unlock()
	defer func() {
		growthTrajectoryFlightMu.Lock()
		delete(growthTrajectoryFlight, wxID)
		close(done)
		growthTrajectoryFlightMu.Unlock()
	}()

	priorFeedback, err := loadGrowthTrajectoryPriorFeedback(ctx, wxID, deviceNo)
	if err != nil {
		glog.Warningf(ctx, "[GrowthTrajectory] 读 prior_feedback 失败 deviceNoLen=%d err=%v", len(deviceNo), err)
		priorFeedback = []interface{}{}
	}

	var pyAnswer *GrowthTrajectoryAnswer
	if answer != nil {
		pyAnswer = &GrowthTrajectoryAnswer{
			QuestionID: strings.TrimSpace(answer.QuestionId),
			Value:      answer.Value,
		}
	}

	turnCtx, cancel := context.WithTimeout(ctx, growthTrajectoryTurnTimeout)
	defer cancel()

	pythonClient := PythonAIClientFromCfg()
	streamErr := pythonClient.GrowthTrajectoryTurnStream(turnCtx, &GrowthTrajectoryTurnRequest{
		DeviceNo:      deviceNo,
		SessionID:     sessionID,
		Action:        action,
		Answer:        pyAnswer,
		PriorFeedback: priorFeedback,
		HorizonDays:   growthTrajectoryHorizonDays,
		Model:         map[string]interface{}{},
	}, &GrowthTrajectoryTurnStreamCallback{
		OnThinking: func(dataJSON string) error {
			if cb != nil && cb.OnThinking != nil {
				return cb.OnThinking(dataJSON)
			}
			return nil
		},
		OnQuestion: func(dataJSON string) error {
			if cb != nil && cb.OnQuestion != nil {
				return cb.OnQuestion(dataJSON)
			}
			return nil
		},
		OnResult: func(dataJSON string) error {
			// 先落库再 INCR / claim，再转发 SSE。
			if uErr := upsertGrowthTrajectoryResultFromEvent(ctx, wxID, deviceNo, dataJSON); uErr != nil {
				glog.Warningf(ctx, "[GrowthTrajectory] result 落库失败 deviceNoLen=%d err=%v", len(deviceNo), uErr)
				return gerror.WrapCode(gcode.CodeInternalError, uErr, "成长轨迹结果保存失败")
			}
			if iErr := incrGrowthTrajectoryDailyUsage(ctx, wxID); iErr != nil {
				glog.Warningf(ctx, "[GrowthTrajectory] 日限 INCR 失败 wxId=%d err=%v", wxID, iErr)
			}
			if cErr := cash.RemoteClaimFeatureTrial(ctx, wxID, "growth_trajectory_predict"); cErr != nil {
				glog.Warningf(ctx, "[GrowthTrajectory] claim trial 失败 wxId=%d err=%v", wxID, cErr)
			}
			enriched := enrichGrowthTrajectoryResultJSON(ctx, wxID, dataJSON)
			if cb != nil && cb.OnResult != nil {
				return cb.OnResult(enriched)
			}
			return nil
		},
		OnError: func(dataJSON string) error {
			if cb != nil && cb.OnError != nil {
				return cb.OnError(dataJSON)
			}
			return nil
		},
		OnDone: func(dataJSON string) error {
			if cb != nil && cb.OnDone != nil {
				return cb.OnDone(dataJSON)
			}
			return nil
		},
	})
	if streamErr != nil {
		return gerror.WrapCode(gcode.CodeInternalError, streamErr, "成长轨迹预测暂时不可用，请稍后再试")
	}
	return nil
}

func requireGrowthTrajectoryAccess(ctx context.Context, deviceNo string, wxID int64) error {
	acc, err := cash.RemoteGrowthTrajectoryAccess(ctx, deviceNo, wxID)
	if err != nil {
		glog.Warningf(ctx, "[GrowthTrajectory] access 调用失败 deviceNoLen=%d wxId=%d err=%v", len(deviceNo), wxID, err)
		return gerror.WrapCode(gcode.CodeOperationFailed, err, "暂时无法校验成长轨迹开通状态")
	}
	if acc == nil || !acc.Allowed {
		return gerror.NewCode(gcode.CodeInvalidOperation, "未开通成长轨迹预测")
	}
	return nil
}

func checkGrowthTrajectoryDailyLimit(ctx context.Context, wxID int64) error {
	limit, err := GetGrowthTrajectoryDailyLimit(ctx)
	if err != nil {
		glog.Warningf(ctx, "[GrowthTrajectory] 读日限失败，使用默认 %d err=%v", growthTrajectoryDefaultLimit, err)
		limit = growthTrajectoryDefaultLimit
	}
	used, err := readGrowthTrajectoryDailyUsage(ctx, wxID)
	if err != nil {
		return gerror.WrapCode(gcode.CodeOperationFailed, err, "暂时无法校验成长轨迹日限")
	}
	if used >= limit {
		return gerror.NewCode(contracts.GCodeGrowthTrajectoryDailyLimit(), contracts.ErrGrowthTrajectoryDailyLimit.Error())
	}
	return nil
}

func shanghaiDayCompact(t time.Time) string {
	return t.In(shanghaiLocation()).Format("20060102")
}

func growthTrajectoryDailyUsageTTL(now time.Time) time.Duration {
	loc := shanghaiLocation()
	now = now.In(loc)
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 2, 0, 0, 0, loc)
	d := next.Sub(now)
	if d < time.Hour {
		return time.Hour
	}
	return d
}

func readGrowthTrajectoryDailyUsage(ctx context.Context, wxID int64) (int, error) {
	day := shanghaiDayCompact(time.Now())
	key, err := cachekit.GrowthTrajectoryDailyUsageKey(wxID, day)
	if err != nil {
		return 0, err
	}
	raw, ok, err := growthTrajectoryCache.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, nil
	}
	n, _ := strconv.Atoi(strings.TrimSpace(raw))
	return n, nil
}

func growthTrajectoryUsageSnapshot(ctx context.Context, wxID int64) (used, limit int) {
	limit, err := GetGrowthTrajectoryDailyLimit(ctx)
	if err != nil || limit <= 0 {
		limit = growthTrajectoryDefaultLimit
	}
	used, uErr := readGrowthTrajectoryDailyUsage(ctx, wxID)
	if uErr != nil {
		glog.Warningf(ctx, "[GrowthTrajectory] 读用量失败 wxId=%d err=%v", wxID, uErr)
		used = 0
	}
	return used, limit
}

func enrichGrowthTrajectoryResultJSON(ctx context.Context, wxID int64, dataJSON string) string {
	used, limit := growthTrajectoryUsageSnapshot(ctx, wxID)
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(dataJSON), &m); err != nil || m == nil {
		m = map[string]interface{}{}
		var typed struct {
			Markdown  string `json:"markdown"`
			SessionID string `json:"sessionId"`
		}
		if json.Unmarshal([]byte(dataJSON), &typed) == nil {
			if typed.Markdown != "" {
				m["markdown"] = typed.Markdown
			}
			if typed.SessionID != "" {
				m["sessionId"] = typed.SessionID
			}
		}
	}
	m["usedToday"] = used
	m["dailyLimit"] = limit
	b, err := json.Marshal(m)
	if err != nil {
		return dataJSON
	}
	return string(b)
}

func incrGrowthTrajectoryDailyUsage(ctx context.Context, wxID int64) error {
	day := shanghaiDayCompact(time.Now())
	key, err := cachekit.GrowthTrajectoryDailyUsageKey(wxID, day)
	if err != nil {
		return err
	}
	n, err := growthTrajectoryCache.Incr(ctx, key)
	if err != nil {
		return err
	}
	if n == 1 {
		_ = growthTrajectoryCache.Expire(ctx, key, growthTrajectoryDailyUsageTTL(time.Now()))
	}
	return nil
}

func loadGrowthTrajectoryLatest(ctx context.Context, wxID int64, deviceNo string) (growthTrajectoryLatestRow, bool, error) {
	// 尚无最新记录时空集为正常路径，须 One+IsEmpty，禁止 Scan 将 ErrNoRows 当系统失败。
	one, err := g.DB().Model("growth_trajectory_latest").Ctx(ctx).
		Where("wx_id", wxID).Where("device_no", deviceNo).One()
	if err != nil {
		return growthTrajectoryLatestRow{}, false, err
	}
	if one.IsEmpty() {
		return growthTrajectoryLatestRow{}, false, nil
	}
	var row growthTrajectoryLatestRow
	if err = one.Struct(&row); err != nil {
		return growthTrajectoryLatestRow{}, false, err
	}
	return row, true, nil
}

func loadGrowthTrajectoryPriorFeedback(ctx context.Context, wxID int64, deviceNo string) ([]interface{}, error) {
	row, ok, err := loadGrowthTrajectoryLatest(ctx, wxID, deviceNo)
	if err != nil {
		return nil, err
	}
	if !ok || strings.TrimSpace(row.FeedbackJson) == "" {
		return []interface{}{}, nil
	}
	var arr []interface{}
	if err := json.Unmarshal([]byte(row.FeedbackJson), &arr); err != nil {
		glog.Warningf(ctx, "[GrowthTrajectory] feedback_json 非数组，降级空 prior deviceNoLen=%d", len(deviceNo))
		return []interface{}{}, nil
	}
	if arr == nil {
		arr = []interface{}{}
	}
	return arr, nil
}

func upsertGrowthTrajectoryResultFromEvent(ctx context.Context, wxID int64, deviceNo, dataJSON string) error {
	var payload struct {
		Markdown  string          `json:"markdown"`
		SessionID string          `json:"sessionId"`
		Feedback  json.RawMessage `json:"feedback"`
	}
	if err := json.Unmarshal([]byte(dataJSON), &payload); err != nil {
		return err
	}
	md := strings.TrimSpace(payload.Markdown)
	sid := strings.TrimSpace(payload.SessionID)
	now := time.Now().Unix()

	feedbackJSON := ""
	if len(payload.Feedback) > 0 && string(payload.Feedback) != "null" {
		feedbackJSON = string(payload.Feedback)
	} else if row, ok, _ := loadGrowthTrajectoryLatest(ctx, wxID, deviceNo); ok {
		feedbackJSON = row.FeedbackJson
	}
	if feedbackJSON == "" {
		feedbackJSON = "[]"
	}

	data := g.Map{
		"wx_id":           wxID,
		"device_no":       deviceNo,
		"result_markdown": md,
		"feedback_json":   feedbackJSON,
		"session_id":      sid,
		"updated_at":      now,
	}
	n, err := g.DB().Model("growth_trajectory_latest").Ctx(ctx).
		Where("wx_id", wxID).Where("device_no", deviceNo).Count()
	if err != nil {
		return err
	}
	if n == 0 {
		_, err = g.DB().Model("growth_trajectory_latest").Ctx(ctx).Data(data).Insert()
		return err
	}
	_, err = g.DB().Model("growth_trajectory_latest").Ctx(ctx).
		Where("wx_id", wxID).Where("device_no", deviceNo).Data(data).Update()
	return err
}
