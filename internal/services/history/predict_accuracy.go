// 预测准确率（Hub 运维）：对齐 Flutter trend_predict_accuracy / predict-list-accuracy-bars。
// 仅计已存 expect_start_time；命中窗 1800s；归日 Asia/Shanghai；禁止重放伪造 expect。
package history

import (
	"context"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	// predictAccuracyHitWindowSec 与 Flutter kPredictAccuracyHitWindowSec 一致：|实际−预计|≤30 分钟。
	predictAccuracyHitWindowSec = 1800
	// predictAccuracyPastDaysBeforeToday 近窗含今天共 7 日（6+1）。
	predictAccuracyPastDaysBeforeToday = 6
	// predictAccuracyRowLimit 单次拉窗上限，防异常大表拖垮 Admin。
	predictAccuracyRowLimit = 2000
)

// PredictAccuracyDayKey 某上海自然日（00:00）及展示用 yyyy-MM-dd。
type PredictAccuracyDayKey struct {
	Date      string `json:"date"`      // yyyy-MM-dd
	DayStart  int64  `json:"dayStart"`  // 该日 00:00 unix 秒（上海）
}

// PredictAccuracyCard 单事件根（或回退 eventId）近 7 日准确率。
type PredictAccuracyCard struct {
	RootEventId   int64   `json:"rootEventId"`
	Title         string  `json:"title"`
	DailyPercents []*int  `json:"dailyPercents"` // 与 Days 等长；无样本为 null
}

// PredictAccuracyOverview Admin 总览响应体。
type PredictAccuracyOverview struct {
	Timezone string                  `json:"timezone"`
	Days     []PredictAccuracyDayKey `json:"days"`
	Cards    []PredictAccuracyCard   `json:"cards"`
}

// PredictAccuracySample 日明细单条。
type PredictAccuracySample struct {
	HistoryId       int64  `json:"historyId"`
	EventId         int64  `json:"eventId"`
	EventName       string `json:"eventName"`
	ExpectStartTime int64  `json:"expectStartTime"`
	ActualStartTime int64  `json:"actualStartTime"`
	Hit             bool   `json:"hit"`
}

// shanghaiLoc 加载 Asia/Shanghai；失败则固定 +8。
func shanghaiLoc() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}

// predictAccuracyWindowDays 近 7 个上海自然日（旧→新，含今天）。
func predictAccuracyWindowDays(now time.Time) []PredictAccuracyDayKey {
	loc := shanghaiLoc()
	n := now.In(loc)
	today := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, loc)
	out := make([]PredictAccuracyDayKey, 0, predictAccuracyPastDaysBeforeToday+1)
	for i := predictAccuracyPastDaysBeforeToday; i >= 0; i-- {
		d := today.AddDate(0, 0, -i)
		out = append(out, PredictAccuracyDayKey{
			Date:     d.Format("2006-01-02"),
			DayStart: d.Unix(),
		})
	}
	return out
}

// dayIndexInWindow 实际开始落入窗口哪一天；不在窗内返回 -1。
func dayIndexInWindow(actualSec int64, days []PredictAccuracyDayKey) int {
	if actualSec <= 0 || len(days) == 0 {
		return -1
	}
	loc := shanghaiLoc()
	t := time.Unix(actualSec, 0).In(loc)
	key := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc).Unix()
	for i, d := range days {
		if d.DayStart == key {
			return i
		}
	}
	return -1
}

// resolveEventRootID 沿 parentId 上溯到根；目录缺失或环则返回自身。
func resolveEventRootID(eventID int64, byID map[int64]entity.Event) int64 {
	if eventID <= 0 {
		return eventID
	}
	seen := map[int64]struct{}{}
	cur := eventID
	for {
		if _, ok := seen[cur]; ok {
			return eventID
		}
		seen[cur] = struct{}{}
		ev, ok := byID[cur]
		if !ok {
			return cur
		}
		if ev.ParentId <= 0 {
			return cur
		}
		cur = ev.ParentId
	}
}

// loadEventCatalogMap 经既有 ListEventOptions（device HTTP 契约）拉字典；失败返回空 map（按 eventId 回退）。
func loadEventCatalogMap(ctx context.Context) map[int64]entity.Event {
	list, err := ListEventOptions(ctx)
	if err != nil {
		glog.Warningf(ctx, "[predict-accuracy] event_catalog_err err=%v", err)
		return map[int64]entity.Event{}
	}
	out := make(map[int64]entity.Event, len(list))
	for _, ev := range list {
		out[ev.Id] = ev
	}
	return out
}

// loadPredictAccuracyWindowRows 一次拉近窗内带正 expect_start_time 的 history 行。
func loadPredictAccuracyWindowRows(ctx context.Context, deviceNo string, startUnix, endUnix int64) ([]entity.History, error) {
	c := dao.History.Columns()
	rows, err := dao.History.Ctx(ctx).
		Fields(historyListSelectFields()...).
		Where(c.DeviceNo, deviceNo).
		WhereGTE(c.StartTime, startUnix).
		WhereLTE(c.StartTime, endUnix).
		WhereGT(c.ExpectStartTime, 0).
		OrderAsc(c.StartTime).
		Limit(predictAccuracyRowLimit).
		All()
	if err != nil {
		return nil, err
	}
	items := make([]entity.History, 0, len(rows))
	for _, row := range rows {
		items = append(items, historyRowToEntity(row))
	}
	return items, nil
}

type accuracyAgg struct {
	title  string
	hits   []int
	counts []int
}

// GetPredictAccuracyOverview 近 7 日按根聚合准确率总览。
func GetPredictAccuracyOverview(ctx context.Context, deviceNo string) (*PredictAccuracyOverview, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	days := predictAccuracyWindowDays(time.Now())
	if len(days) == 0 {
		return &PredictAccuracyOverview{Timezone: "Asia/Shanghai", Days: days, Cards: []PredictAccuracyCard{}}, nil
	}
	windowStart := days[0].DayStart
	// 右端用「今天结束」：次日 00:00 - 1s，覆盖今日全日实际开始。
	loc := shanghaiLoc()
	todayEnd := time.Unix(days[len(days)-1].DayStart, 0).In(loc).AddDate(0, 0, 1).Unix() - 1

	rows, err := loadPredictAccuracyWindowRows(ctx, deviceNo, windowStart, todayEnd)
	if err != nil {
		return nil, err
	}
	catalog := loadEventCatalogMap(ctx)
	aggs := map[int64]*accuracyAgg{}

	for _, h := range rows {
		if h.ExpectStartTime == nil || *h.ExpectStartTime <= 0 || h.StartTime <= 0 {
			continue
		}
		expectSec := *h.ExpectStartTime
		actualSec := h.StartTime
		di := dayIndexInWindow(actualSec, days)
		if di < 0 {
			continue
		}
		rootID := resolveEventRootID(h.EventId, catalog)
		agg, ok := aggs[rootID]
		if !ok {
			title := strings.TrimSpace(h.EventName)
			if ev, ok2 := catalog[rootID]; ok2 && strings.TrimSpace(ev.Name) != "" {
				title = strings.TrimSpace(ev.Name)
			}
			if title == "" {
				title = "事件#" + strconv.FormatInt(rootID, 10)
			}
			agg = &accuracyAgg{
				title:  title,
				hits:   make([]int, len(days)),
				counts: make([]int, len(days)),
			}
			aggs[rootID] = agg
		}
		agg.counts[di]++
		diff := actualSec - expectSec
		if diff < 0 {
			diff = -diff
		}
		if diff <= predictAccuracyHitWindowSec {
			agg.hits[di]++
		}
		// 标题：窗内较新行可刷新（rows 已按 start 升序）
		if name := strings.TrimSpace(h.EventName); name != "" && rootID == h.EventId {
			agg.title = name
		}
	}

	rootIDs := make([]int64, 0, len(aggs))
	for id := range aggs {
		rootIDs = append(rootIDs, id)
	}
	sort.Slice(rootIDs, func(i, j int) bool { return rootIDs[i] < rootIDs[j] })

	cards := make([]PredictAccuracyCard, 0, len(rootIDs))
	for _, id := range rootIDs {
		agg := aggs[id]
		percents := make([]*int, len(days))
		for i := range days {
			n := agg.counts[i]
			if n < 1 {
				continue
			}
			p := int(math.Round(100.0 * float64(agg.hits[i]) / float64(n)))
			percents[i] = &p
		}
		cards = append(cards, PredictAccuracyCard{
			RootEventId:   id,
			Title:         agg.title,
			DailyPercents: percents,
		})
	}
	return &PredictAccuracyOverview{
		Timezone: "Asia/Shanghai",
		Days:     days,
		Cards:    cards,
	}, nil
}

// GetPredictAccuracyDaySamples 某根某上海自然日的合格样本明细。
func GetPredictAccuracyDaySamples(ctx context.Context, deviceNo string, rootEventID int64, dayYYYYMMDD string) ([]PredictAccuracySample, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	dayYYYYMMDD = strings.TrimSpace(dayYYYYMMDD)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	if rootEventID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "rootEventId 无效")
	}
	loc := shanghaiLoc()
	dayStart, err := time.ParseInLocation("2006-01-02", dayYYYYMMDD, loc)
	if err != nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "day 须为 yyyy-MM-dd")
	}
	dayEnd := dayStart.AddDate(0, 0, 1).Unix() - 1
	startUnix := dayStart.Unix()

	rows, err := loadPredictAccuracyWindowRows(ctx, deviceNo, startUnix, dayEnd)
	if err != nil {
		return nil, err
	}
	catalog := loadEventCatalogMap(ctx)
	out := make([]PredictAccuracySample, 0)
	for _, h := range rows {
		if h.ExpectStartTime == nil || *h.ExpectStartTime <= 0 || h.StartTime <= 0 {
			continue
		}
		if resolveEventRootID(h.EventId, catalog) != rootEventID {
			continue
		}
		expectSec := *h.ExpectStartTime
		actualSec := h.StartTime
		diff := actualSec - expectSec
		if diff < 0 {
			diff = -diff
		}
		out = append(out, PredictAccuracySample{
			HistoryId:       h.Id,
			EventId:         h.EventId,
			EventName:       strings.TrimSpace(h.EventName),
			ExpectStartTime: expectSec,
			ActualStartTime: actualSec,
			Hit:             diff <= predictAccuracyHitWindowSec,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ActualStartTime < out[j].ActualStartTime
	})
	return out, nil
}
