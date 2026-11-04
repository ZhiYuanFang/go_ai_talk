// Timing Live Activity：挂 history 权威持久化成功路径一次扇出（经 clients/push），禁止与 voice adapter 双钩。
package history

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	deviceclient "hello/internal/clients/device"
	pushclient "hello/internal/clients/push"
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/os/glog"
)

const historyLiveActivityEnabledEnv = "PUSH_LIVE_ACTIVITY_ENABLED"

var historyLACache = cachekit.Default()

func historyLiveActivityEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(historyLiveActivityEnabledEnv)))
	return v == "1" || v == "true" || v == "on" || v == "yes"
}

func timingActivityID(deviceNo string, historyID int64) string {
	return fmt.Sprintf("timing:%s:%d", strings.TrimSpace(deviceNo), historyID)
}

// maybeStartTimingLiveActivity 未闭合计时落库成功后扇出 Timing start。
// Args: eventID 为 history.EventId（与 Flutter 事件目录 id 对齐；<=0 时客户端走 BrandLogo 兜底）；
//
//	startTime 为 unix 秒（与库内 history.start_time BIGINT 语义一致，禁止毫秒）。
//
// Side Effects: 经 PUSH_SERVICE_URL 调 live-activity；写 Redis 槽跟踪。
// 契约：固定 Attributes MUST 含 eventId、eventName、已知 startTime；ContentState 同步；
// 载荷 MUST NOT 含 logo/color/logoFile/远程图 URL（视觉由客户端按 eventId 本地解析）。
func maybeStartTimingLiveActivity(ctx context.Context, deviceNo string, historyID, eventID int64, eventName string, startTime, endTime int64) {
	if !historyLiveActivityEnabled() || historyID <= 0 || endTime != 0 {
		return
	}
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return
	}
	// eventId 缺省不伪造正 id；仍可 start 计时文案，锁屏视觉走客户端兜底。
	if eventID <= 0 {
		glog.Warningf(ctx, "[history-la] timing_start eventId_missing historyId=%d deviceNoLen=%d", historyID, len(deviceNo))
	}
	// Push-to-Start 常不套用 ContentState：展示名写入固定 Attributes；行内空名则主档回填。
	displayName := notifyEventDisplayName(ctx, eventID, eventName)
	wxIDs, _, err := deviceclient.RemoteListWxIDsByDeviceNo(ctx, deviceNo)
	if err != nil {
		glog.Warningf(ctx, "[history-la] timing_start list_wx_err deviceNoLen=%d err=%v", len(deviceNo), err)
		return
	}
	activityID := timingActivityID(deviceNo, historyID)
	// 固定 Attributes：eventId 供本地色/图；eventName/startTime 供锁屏标题与计时回退；禁止 logo/color。
	attrs := map[string]interface{}{
		"activityId": activityID,
		"deviceNo":   deviceNo,
		"historyId":  historyID,
		"eventId":    eventID,
		"eventName":  displayName,
	}
	if startTime > 0 {
		attrs["startTime"] = startTime
	}
	// ContentState 仍带同名字段；startTime 必须为 unix 秒。
	cs := map[string]interface{}{
		"eventName": displayName,
		"startTime": startTime,
		"phase":     "timing",
	}
	for _, wxID := range wxIDs {
		if pushErr := pushclient.FanoutLiveActivity(ctx, pushclient.LiveActivityOp{
			WxID:           wxID,
			Action:         "start",
			AttributesKind: "timing",
			ActivityID:     activityID,
			Attributes:     attrs,
			ContentState:   cs,
			AlertTitle:     "计时中",
			AlertBody:      displayName,
		}); pushErr != nil {
			glog.Warningf(ctx, "[history-la] timing_start_failed wxId=%d historyId=%d err=%v", wxID, historyID, pushErr)
			continue
		}
		if slotKey, kErr := cachekit.PredictTimingActivitySlotKey(deviceNo, historyID); kErr == nil {
			_ = historyLACache.Set(ctx, slotKey, activityID)
		}
	}
}

// activeEtaExpectRow 只读解析 active-eta Redis JSON 所需字段（与 voice PredictActiveEtaItem 对齐；禁止 import voice）。
type activeEtaExpectRow struct {
	HistoryId   string `json:"historyId"`
	ExpectEndAt int64  `json:"expectEndAt"`
}

// lookupActiveEtaExpectEndAt 从 PredictActiveEtaKey 按 historyId 取正 expectEndAt；无匹配/失败返回 0。
func lookupActiveEtaExpectEndAt(ctx context.Context, deviceNo string, historyID int64) int64 {
	if historyID <= 0 {
		return 0
	}
	key, err := cachekit.PredictActiveEtaKey(deviceNo)
	if err != nil {
		return 0
	}
	val, ok, getErr := historyLACache.Get(ctx, key)
	if getErr != nil || !ok || strings.TrimSpace(val) == "" {
		return 0
	}
	var rows []activeEtaExpectRow
	if json.Unmarshal([]byte(val), &rows) != nil {
		glog.Warningf(ctx, "[history-la] active_eta corrupt deviceNoLen=%d historyId=%d", len(deviceNo), historyID)
		return 0
	}
	want := strconv.FormatInt(historyID, 10)
	for _, row := range rows {
		if strings.TrimSpace(row.HistoryId) != want {
			continue
		}
		if row.ExpectEndAt > 0 {
			return row.ExpectEndAt
		}
		return 0
	}
	return 0
}

// updateTimingLiveActivity 未闭合 history 更新后扇出 Timing ContentState update（不重建活动，图标可沿用 start Attributes）。
// Args: eventID 用于主档回填展示名；startTime 为 unix 秒，>0 才写入，禁止显式 0。
// Side Effects: 经 PUSH_SERVICE_URL 调 live-activity update；无 activity token 时 push-service 跳过。
// 因 APNs content-state 整包替换：若 active-eta Redis 仍有该 historyId 的正 expectEndAt，MUST 一并写回，避免冲掉副行。
func updateTimingLiveActivity(ctx context.Context, deviceNo string, historyID, eventID int64, eventName string, startTime int64) {
	if !historyLiveActivityEnabled() || historyID <= 0 {
		return
	}
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return
	}
	displayName := notifyEventDisplayName(ctx, eventID, eventName)
	wxIDs, _, err := deviceclient.RemoteListWxIDsByDeviceNo(ctx, deviceNo)
	if err != nil {
		glog.Warningf(ctx, "[history-la] timing_update list_wx_err deviceNoLen=%d err=%v", len(deviceNo), err)
		return
	}
	activityID := timingActivityID(deviceNo, historyID)
	cs := map[string]interface{}{"phase": "timing"}
	if displayName != "" {
		cs["eventName"] = displayName
	}
	if startTime > 0 {
		cs["startTime"] = startTime
	}
	// 只读 platform 键合并 ETA；无匹配则 omit（不阻断改名/改 startTime）。
	if expectEndAt := lookupActiveEtaExpectEndAt(ctx, deviceNo, historyID); expectEndAt > 0 {
		cs["expectEndAt"] = expectEndAt
	}
	for _, wxID := range wxIDs {
		if pushErr := pushclient.FanoutLiveActivity(ctx, pushclient.LiveActivityOp{
			WxID:           wxID,
			Action:         "update",
			AttributesKind: "timing",
			ActivityID:     activityID,
			ContentState:   cs,
		}); pushErr != nil {
			glog.Warningf(ctx, "[history-la] timing_update_failed wxId=%d historyId=%d err=%v", wxID, historyID, pushErr)
		}
	}
}

// maybeEndTimingLiveActivity 闭合（endTime 写入）后扇出 Timing end。
// Args: startTime/eventName 来自该 history 行，写入 end ContentState，避免客户端把缺省 startTime 解成 0。
// endTime==0 视为未闭合，不发 end（删除路径请用 endTimingLiveActivity）。
func maybeEndTimingLiveActivity(ctx context.Context, deviceNo string, historyID int64, eventName string, startTime, endTime int64) {
	if endTime == 0 {
		return
	}
	endTimingLiveActivity(ctx, deviceNo, historyID, eventName, startTime)
}

// endTimingLiveActivity 按 activity id 扇出 Timing end（闭合与删除共用）。
// 不检查 endTime：删除未闭合行时仍须摘锁屏。无 LA token 时 push-service 跳过。
// ContentState：phase=ended；已知 startTime>0 / eventName 非空则带上，MUST NOT 把已知 startTime 写成 0。
// dismissal-date 由 push-service 组装层填过去秒（立即摘条）。
func endTimingLiveActivity(ctx context.Context, deviceNo string, historyID int64, eventName string, startTime int64) {
	if !historyLiveActivityEnabled() || historyID <= 0 {
		return
	}
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return
	}
	wxIDs, _, err := deviceclient.RemoteListWxIDsByDeviceNo(ctx, deviceNo)
	if err != nil {
		glog.Warningf(ctx, "[history-la] timing_end list_wx_err deviceNoLen=%d err=%v", len(deviceNo), err)
		return
	}
	activityID := timingActivityID(deviceNo, historyID)
	cs := map[string]interface{}{"phase": "ended"}
	eventName = strings.TrimSpace(eventName)
	if eventName != "" {
		cs["eventName"] = eventName
	}
	// 已知真实开始时间才写入，禁止显式 0（Flutter 会从 epoch 错乱计时）。
	if startTime > 0 {
		cs["startTime"] = startTime
	}
	for _, wxID := range wxIDs {
		if pushErr := pushclient.FanoutLiveActivity(ctx, pushclient.LiveActivityOp{
			WxID:           wxID,
			Action:         "end",
			AttributesKind: "timing",
			ActivityID:     activityID,
			ContentState:   cs,
		}); pushErr != nil {
			glog.Warningf(ctx, "[history-la] timing_end_failed wxId=%d historyId=%d err=%v", wxID, historyID, pushErr)
		}
	}
	if slotKey, kErr := cachekit.PredictTimingActivitySlotKey(deviceNo, historyID); kErr == nil {
		_ = historyLACache.Del(ctx, slotKey)
	}
}
