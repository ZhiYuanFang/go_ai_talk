// Timing Live Activity：挂 history 权威持久化成功路径一次扇出（经 clients/push），禁止与 voice adapter 双钩。
package history

import (
	"context"
	"fmt"
	"os"
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
// 契约：固定 Attributes MUST 含 eventId；ContentState.startTime MUST 为 unix 秒；
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
	wxIDs, _, err := deviceclient.RemoteListWxIDsByDeviceNo(ctx, deviceNo)
	if err != nil {
		glog.Warningf(ctx, "[history-la] timing_start list_wx_err deviceNoLen=%d err=%v", len(deviceNo), err)
		return
	}
	activityID := timingActivityID(deviceNo, historyID)
	// 固定 Attributes：含 eventId 供 Extension 读 App Group 色/图；禁止写入 logo/color。
	attrs := map[string]interface{}{
		"activityId": activityID,
		"deviceNo":   deviceNo,
		"historyId":  historyID,
		"eventId":    eventID,
	}
	// ContentState.startTime 必须为 unix 秒，供锁屏系统计时展示「已计时」。
	cs := map[string]interface{}{
		"eventName": strings.TrimSpace(eventName),
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
			AlertBody:      strings.TrimSpace(eventName),
		}); pushErr != nil {
			glog.Warningf(ctx, "[history-la] timing_start_failed wxId=%d historyId=%d err=%v", wxID, historyID, pushErr)
			continue
		}
		if slotKey, kErr := cachekit.PredictTimingActivitySlotKey(deviceNo, historyID); kErr == nil {
			_ = historyLACache.Set(ctx, slotKey, activityID)
		}
	}
}

// maybeEndTimingLiveActivity 闭合（endTime 写入）后扇出 Timing end。
func maybeEndTimingLiveActivity(ctx context.Context, deviceNo string, historyID int64, endTime int64) {
	if !historyLiveActivityEnabled() || historyID <= 0 || endTime == 0 {
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
	for _, wxID := range wxIDs {
		if pushErr := pushclient.FanoutLiveActivity(ctx, pushclient.LiveActivityOp{
			WxID:           wxID,
			Action:         "end",
			AttributesKind: "timing",
			ActivityID:     activityID,
			ContentState:   map[string]interface{}{"phase": "ended"},
		}); pushErr != nil {
			glog.Warningf(ctx, "[history-la] timing_end_failed wxId=%d historyId=%d err=%v", wxID, historyID, pushErr)
		}
	}
	if slotKey, kErr := cachekit.PredictTimingActivitySlotKey(deviceNo, historyID); kErr == nil {
		_ = historyLACache.Del(ctx, slotKey)
	}
}
