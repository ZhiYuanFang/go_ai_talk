// 统一 Live Activity Refresh：组装「lead 进窗预测 + 进行中 history」列表态 ContentState（总条数≤3），
// 经 clients/push 扇出 upsert/end。MUST NOT import services/voice。
package history

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	deviceclient "hello/internal/clients/device"
	pushclient "hello/internal/clients/push"
	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/platform/cachekit"
	"hello/internal/platform/predictcfg"

	"github.com/gogf/gf/v2/os/glog"
)

// liveActivityMaxItems 锁屏列表硬顶：进行中优先截断后合计最多 3 条。
const liveActivityMaxItems = 3

// liveActivityRefreshCoalesce 同 deviceNo trailing 合并窗：避免 pending PUT / MQ 到点连打多次扇出。
// 仅为按键短窗 AfterFunc，不是扫表/outbox reconciler ticker。
const liveActivityRefreshCoalesce = 300 * time.Millisecond

const historyLiveActivityEnabledEnv = "PUSH_LIVE_ACTIVITY_ENABLED"

var historyLACache = cachekit.Default()

// laRefresh 合并态：每 deviceNo 至多一个未触发的 timer。
var (
	laRefreshMu     sync.Mutex
	laRefreshTimers = map[string]*time.Timer{}
)

func historyLiveActivityEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(historyLiveActivityEnabledEnv)))
	return v == "1" || v == "true" || v == "on" || v == "yes"
}

// laPendingItem Redis imminent pending JSON 字段（与 voice 侧对齐；禁止 import voice）。
type laPendingItem struct {
	EventId           int64   `json:"eventId"`
	NextAt            int64   `json:"nextAt"`
	Title             string  `json:"title,omitempty"`
	ExpectKind        string  `json:"expectKind,omitempty"`
	ExpectDurationSec int64   `json:"expectDurationSec,omitempty"`
	ExpectAmount      float64 `json:"expectAmount,omitempty"`
	ExpectUnit        string  `json:"expectUnit,omitempty"`
}

// laActiveEtaRow active-eta Redis 行（合并 expectEndAt 到进行中项）。
type laActiveEtaRow struct {
	HistoryId   string `json:"historyId"`
	ExpectEndAt int64  `json:"expectEndAt"`
	Title       string `json:"title,omitempty"`
	StartAt     int64  `json:"startAt,omitempty"`
}

// invalidateImminentPending 删除宝宝 imminent pending 权威键（作废旧日程，服务普通推送对账）。
func invalidateImminentPending(ctx context.Context, deviceNo string) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return
	}
	pendingKey, keyErr := cachekit.PredictImminentPendingKey(deviceNo)
	if keyErr != nil {
		glog.Warningf(ctx, "[history-la] pending_key_err deviceNoLen=%d err=%v", len(deviceNo), keyErr)
		return
	}
	if delErr := historyLACache.Del(ctx, pendingKey); delErr != nil {
		glog.Warningf(ctx, "[history-la] pending_del_err deviceNoLen=%d err=%v", len(deviceNo), delErr)
	}
}

// afterHistoryMutationLiveActivity history CUD 成功后：仅 Del pending（作废旧日程）。
// 业务：C/U/D 后客户端会重传预测；LA 改由 pending PUT / MQ 到点触发，此处 MUST NOT Refresh。
// Side Effects: Del Redis pending。
func afterHistoryMutationLiveActivity(ctx context.Context, deviceNo string) {
	invalidateImminentPending(ctx, deviceNo)
}

// ScheduleRefreshBabyLiveActivity 同 deviceNo trailing 300ms 合并后再投影。
//
// 业务：供预测 pending PUT、MQ 到点（经 internal HTTP）调用；窗内多次 schedule 只执行最后一次到期 Refresh。
// MUST NOT 用于 history CUD。非扫表 ticker。
// Side Effects: 重置/创建进程内 timer；到期调用 RefreshBabyLiveActivity。
func ScheduleRefreshBabyLiveActivity(deviceNo string) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return
	}
	laRefreshMu.Lock()
	defer laRefreshMu.Unlock()
	if t, ok := laRefreshTimers[deviceNo]; ok && t != nil {
		t.Stop()
	}
	dn := deviceNo
	laRefreshTimers[deviceNo] = time.AfterFunc(liveActivityRefreshCoalesce, func() {
		laRefreshMu.Lock()
		delete(laRefreshTimers, dn)
		laRefreshMu.Unlock()
		RefreshBabyLiveActivity(context.Background(), dn)
	})
}

// RefreshBabyLiveActivity 统一投影：读 pending + 未闭合 history → 进窗过滤/截断 → 各绑定 wx upsert 或 end。
//
// 业务：
//  1. 预测段：仅 lead 进窗（nextAt-leadSeconds≤now，与 VOICE_PREDICT_IMMINENT_LEAD_SECONDS 同语义），nextAt 升序；
//  2. 进行中段：end_time=0，startTime 倒序；总条数≤3，截断进行中优先，展示仍预测在前；
//  3. 空 items → end；非空 → action=upsert（push 侧有 activity token 则 update，否则 start）；
//  4. 开关关则跳过扇出；失败仅 WARN。
//     对外触发应优先 ScheduleRefreshBabyLiveActivity，避免连打。
//
// Side Effects: 经 PUSH_SERVICE_URL 调 live-activity；可读 cachekit pending / active-eta。
func RefreshBabyLiveActivity(ctx context.Context, deviceNo string) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return
	}
	if !historyLiveActivityEnabled() {
		return
	}

	items := buildLiveActivityItems(ctx, deviceNo)
	wxIDs, _, err := deviceclient.RemoteListWxIDsByDeviceNo(ctx, deviceNo)
	if err != nil {
		glog.Warningf(ctx, "[history-la] refresh list_wx_err deviceNoLen=%d err=%v", len(deviceNo), err)
		return
	}
	if len(wxIDs) == 0 {
		return
	}

	cs := map[string]interface{}{"items": items}
	if len(items) == 0 {
		cs["phase"] = "cleared"
	} else {
		cs["phase"] = "active"
	}

	for _, wxID := range wxIDs {
		if wxID <= 0 {
			continue
		}
		if len(items) == 0 {
			if pushErr := pushclient.FanoutLiveActivity(ctx, pushclient.LiveActivityOp{
				WxID:         wxID,
				Action:       "end",
				ContentState: cs,
			}); pushErr != nil {
				glog.Warningf(ctx, "[history-la] refresh_end_failed wxId=%d err=%v", wxID, pushErr)
			}
			continue
		}
		attrs := map[string]interface{}{
			"activityId": fmt.Sprintf("baby:%d", wxID),
			"deviceNo":   deviceNo,
		}
		// AlertTitle/AlertBody：供 push 解析成 event=start 时写入 aps.alert（真机 start 无 alert 不弹出）；
		// update/end 由组装层忽略这些字段，保持静默刷新。
		if pushErr := pushclient.FanoutLiveActivity(ctx, pushclient.LiveActivityOp{
			WxID:         wxID,
			Action:       "upsert",
			Attributes:   attrs,
			ContentState: cs,
			AlertTitle:   "宝宝动态",
			AlertBody:    liveActivityAlertBody(items),
		}); pushErr != nil {
			glog.Warningf(ctx, "[history-la] refresh_upsert_failed wxId=%d err=%v", wxID, pushErr)
		}
	}
}

// buildLiveActivityItems 组装列表：lead 进窗预测 + 进行中，合计≤3（进行中优先截断）；禁止 logo/color/远程图 URL。
//
// 业务：
//  1. 预测候选：pending 中 nextAt-leadSeconds≤now（lead=predictcfg.LeadSeconds，对齐普通推送 fireAt）；
//  2. 进行中候选：end_time=0，按 startTime 倒序；
//  3. 截断：先取至多 liveActivityMaxItems 条 timing，剩余名额填预测；发射顺序仍「预测段+进行中段」；
//  4. imminent title：投影时拼 `{nick}要{事件}了`（独立于 voice 普通推送；expect 仍走独立字段）。
func buildLiveActivityItems(ctx context.Context, deviceNo string) []map[string]interface{} {
	now := time.Now().Unix()
	leadSec := predictcfg.LeadSeconds()

	pending := loadLAPending(ctx, deviceNo)
	// 仅保留进窗预测；远点仍可留在 Redis 供 MQ，不进锁屏。
	inWindow := make([]laPendingItem, 0, len(pending))
	for _, p := range pending {
		if p.EventId <= 0 || p.NextAt <= 0 {
			continue
		}
		if p.NextAt-leadSec > now {
			continue
		}
		inWindow = append(inWindow, p)
	}
	sort.SliceStable(inWindow, func(i, j int) bool {
		return inWindow[i].NextAt < inWindow[j].NextAt
	})

	openRows := listOpenHistories(ctx, deviceNo)
	sort.SliceStable(openRows, func(i, j int) bool {
		return openRows[i].StartTime > openRows[j].StartTime
	})
	etaByHistory := loadActiveEtaExpectMap(ctx, deviceNo)

	// 进行中优先占位，再把剩余槽位给最近预测。
	takeT := len(openRows)
	if takeT > liveActivityMaxItems {
		takeT = liveActivityMaxItems
	}
	takeP := liveActivityMaxItems - takeT
	if takeP > len(inWindow) {
		takeP = len(inWindow)
	}
	selectedPredict := inWindow[:takeP]
	selectedTiming := openRows[:takeT]

	// 仅当有预测行时才取昵称，避免纯 timing 扇出多打画像。
	nick := ""
	if takeP > 0 {
		nick = liveActivityBabyNick(ctx, deviceNo)
	}

	out := make([]map[string]interface{}, 0, takeP+takeT)
	for _, p := range selectedPredict {
		item := map[string]interface{}{
			"kind":    "imminent",
			"eventId": p.EventId,
			"nextAt":  p.NextAt,
		}
		// Redis 存裸事件名；投影成品句供锁屏左侧；量/时不进 title。
		if composed := composeImminentLiveActivityTitle(nick, p.Title); composed != "" {
			item["title"] = composed
		}
		applyPendingExpect(item, p)
		out = append(out, item)
	}
	for _, h := range selectedTiming {
		item := map[string]interface{}{
			"kind":      "timing",
			"historyId": h.Id,
			"eventId":   h.EventId,
		}
		name := notifyEventDisplayName(ctx, h.EventId, h.EventName)
		if name != "" {
			item["eventName"] = name
		}
		if h.StartTime > 0 {
			item["startTime"] = h.StartTime
		}
		if expectEndAt, ok := etaByHistory[h.Id]; ok && expectEndAt > 0 {
			item["expectEndAt"] = expectEndAt
		} else if h.ExpectEndTime != nil && *h.ExpectEndTime > 0 {
			item["expectEndAt"] = *h.ExpectEndTime
		}
		out = append(out, item)
	}
	return out
}

// applyPendingExpect 将可用 expect 结构化字段写入 item（供 Flutter 右侧短槽）；不拼进 title。
func applyPendingExpect(item map[string]interface{}, p laPendingItem) {
	kind := strings.TrimSpace(strings.ToLower(p.ExpectKind))
	switch kind {
	case "duration":
		if p.ExpectDurationSec <= 0 {
			return
		}
		item["expectKind"] = "duration"
		item["expectDurationSec"] = p.ExpectDurationSec
	case "amount":
		if p.ExpectAmount <= 0 {
			return
		}
		item["expectKind"] = "amount"
		item["expectAmount"] = p.ExpectAmount
		if u := strings.TrimSpace(p.ExpectUnit); u != "" {
			item["expectUnit"] = u
		}
	}
}

// liveActivityBabyNick 经 history 画像契约取宝宝名；空或失败回落「宝宝」（与普通推送 nick 语义对齐）。
//
// Args: ctx、deviceNo 设备号。
// Returns: 非空展示名。
// Side Effects: 可能 miss 回源 device profile；失败仅 WARN。
func liveActivityBabyNick(ctx context.Context, deviceNo string) string {
	const fallback = "宝宝"
	babyName, _, _, err := GetDeviceBirthday(ctx, deviceNo)
	if err != nil {
		glog.Warningf(ctx, "[history-la] profile_err deviceNoLen=%d err=%v (fallback nick=%s)",
			len(deviceNo), err, fallback)
		return fallback
	}
	if n := strings.TrimSpace(babyName); n != "" {
		return n
	}
	return fallback
}

// composeImminentLiveActivityTitle 拼锁屏预测标题 `{nick}要{事件}了`；事件名为空返回 ""（省略 title 键）。
//
// Args: nick 已解析昵称（调用方保证非空或回落值）；eventName Redis pending 裸事件名。
// Returns: 成品句或空串。
func composeImminentLiveActivityTitle(nick, eventName string) string {
	eventName = strings.TrimSpace(eventName)
	if eventName == "" {
		return ""
	}
	nick = strings.TrimSpace(nick)
	if nick == "" {
		nick = "宝宝"
	}
	return nick + "要" + eventName + "了"
}

func loadLAPending(ctx context.Context, deviceNo string) []laPendingItem {
	key, err := cachekit.PredictImminentPendingKey(deviceNo)
	if err != nil {
		return nil
	}
	val, ok, getErr := historyLACache.Get(ctx, key)
	if getErr != nil || !ok || strings.TrimSpace(val) == "" {
		return nil
	}
	var list []laPendingItem
	if json.Unmarshal([]byte(val), &list) != nil {
		glog.Warningf(ctx, "[history-la] pending corrupt deviceNoLen=%d", len(deviceNo))
		return nil
	}
	return list
}

// listOpenHistories 列出设备全部未闭合（end_time=0）history 行。
func listOpenHistories(ctx context.Context, deviceNo string) []entity.History {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil
	}
	rows, err := dao.History.Ctx(ctx).
		Fields(historyListSelectFields()...).
		Where(dao.History.Columns().DeviceNo, deviceNo).
		Where(dao.History.Columns().EndTime, 0).
		OrderDesc(dao.History.Columns().StartTime).
		All()
	if err != nil {
		glog.Warningf(ctx, "[history-la] list_open_err deviceNoLen=%d err=%v", len(deviceNo), err)
		return nil
	}
	out := make([]entity.History, 0, len(rows))
	for _, row := range rows {
		out = append(out, historyRowToEntity(row))
	}
	return out
}

func loadActiveEtaExpectMap(ctx context.Context, deviceNo string) map[int64]int64 {
	key, err := cachekit.PredictActiveEtaKey(deviceNo)
	if err != nil {
		return nil
	}
	val, ok, getErr := historyLACache.Get(ctx, key)
	if getErr != nil || !ok || strings.TrimSpace(val) == "" {
		return nil
	}
	var rows []laActiveEtaRow
	if json.Unmarshal([]byte(val), &rows) != nil {
		glog.Warningf(ctx, "[history-la] active_eta corrupt deviceNoLen=%d", len(deviceNo))
		return nil
	}
	m := make(map[int64]int64, len(rows))
	for _, row := range rows {
		hid, perr := strconv.ParseInt(strings.TrimSpace(row.HistoryId), 10, 64)
		if perr != nil || hid <= 0 || row.ExpectEndAt <= 0 {
			continue
		}
		m[hid] = row.ExpectEndAt
	}
	return m
}

// liveActivityAlertBody APNs 伴随文案（供 start 的 aps.alert）：首条 imminent 用投影后的成品 title（已含 nick）；否则 eventName / 兜底。
func liveActivityAlertBody(items []map[string]interface{}) string {
	if len(items) == 0 {
		return ""
	}
	first := items[0]
	if kind, _ := first["kind"].(string); kind == "imminent" {
		if t, ok := first["title"].(string); ok && strings.TrimSpace(t) != "" {
			return strings.TrimSpace(t)
		}
	}
	if n, ok := first["eventName"].(string); ok && strings.TrimSpace(n) != "" {
		return strings.TrimSpace(n)
	}
	return "有新的动态"
}
