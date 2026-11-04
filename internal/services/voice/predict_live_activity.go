// 预测 Imminent Live Activity 编排（voice 侧）：限流、清场、单槽跟踪；经 clients/push 扇出。
// Timing 起停 MUST 挂 history 落库路径，本文件不触发 Timing。
package voice

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	deviceclient "hello/internal/clients/device"
	pushclient "hello/internal/clients/push"
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/os/glog"
)

const (
	predictLiveActivityEnabledEnv = "PUSH_LIVE_ACTIVITY_ENABLED"
	predictImminentRateLimitTTL   = 60 * time.Second
)

// predictLiveActivityEnabled 与 push-service 共用 env；默认关。
func predictLiveActivityEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(predictLiveActivityEnabledEnv)))
	return v == "1" || v == "true" || v == "on" || v == "yes"
}

// imminentActivityID 每 wx 单槽权威 id（多监护人各一条）。
func imminentActivityID(wxID int64) string {
	return fmt.Sprintf("imminent:%d", wxID)
}

// tryOccupyImminentRateLimit 用户级 1/min 先到占坑；未过闸返回 false。
func tryOccupyImminentRateLimit(ctx context.Context, wxID int64) bool {
	key, err := cachekit.PredictImminentRateLimitKey(wxID)
	if err != nil {
		glog.Warningf(ctx, "[predict-la] rate_key_err wxId=%d err=%v", wxID, err)
		return false
	}
	ok, nxErr := predictImminentCache.SetNXEX(ctx, key, "1", predictImminentRateLimitTTL)
	if nxErr != nil {
		glog.Warningf(ctx, "[predict-la] rate_setnx_err wxId=%d err=%v", wxID, nxErr)
		return false
	}
	return ok
}

// clearImminentLiveActivitiesForDevice pending 同步清场：向绑定用户 end 全部 Imminent，不得 end Timing。
func clearImminentLiveActivitiesForDevice(ctx context.Context, deviceNo string) {
	if !predictLiveActivityEnabled() {
		return
	}
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return
	}
	wxIDs, _, err := deviceclient.RemoteListWxIDsByDeviceNo(ctx, deviceNo)
	if err != nil {
		glog.Warningf(ctx, "[predict-la] clear list_wx_err deviceNoLen=%d err=%v", len(deviceNo), err)
		return
	}
	for _, wxID := range wxIDs {
		// pending 清场无单槽 title/nextAt 缓存；省略字段而非写 0，dismissal-date 组装层兜底。
		endImminentSlotForWx(ctx, wxID, "", 0)
	}
}

// endImminentSlotForWx 结束该用户 Imminent 槽（有跟踪 id 则 end；清 Redis 槽）。
// ContentState：phase=cleared；已知 title/nextAt 才写入，MUST NOT 把 nextAt 写成 0（Flutter 会解成「已过时」花屏）。
// dismissal-date 由 push-service 组装层填过去秒（立即摘条）。
func endImminentSlotForWx(ctx context.Context, wxID int64, title string, nextAt int64) {
	if wxID <= 0 {
		return
	}
	slotKey, err := cachekit.PredictImminentActivitySlotKey(wxID)
	if err != nil {
		return
	}
	activityID := imminentActivityID(wxID)
	if val, ok, _ := predictImminentCache.Get(ctx, slotKey); ok && strings.TrimSpace(val) != "" {
		activityID = strings.TrimSpace(val)
	}
	cs := map[string]interface{}{"phase": "cleared"}
	title = strings.TrimSpace(title)
	if title != "" {
		cs["title"] = title
	}
	if nextAt > 0 {
		cs["nextAt"] = nextAt
	}
	if pushErr := pushclient.FanoutLiveActivity(ctx, pushclient.LiveActivityOp{
		WxID:           wxID,
		Action:         "end",
		AttributesKind: "imminent",
		ActivityID:     activityID,
		ContentState:   cs,
	}); pushErr != nil {
		glog.Warningf(ctx, "[predict-la] end_imminent_failed wxId=%d err=%v", wxID, pushErr)
	}
	_ = predictImminentCache.Del(ctx, slotKey)
}

// startImminentLiveActivityForWx 过闸后：先 end 旧槽再 start；写 Redis 槽。
// Args: eventID 与 Flutter 事件目录 id 对齐；nextAt 为 pending/MQ 载荷中的 unix 秒（非毫秒）。
// 契约：固定 Attributes MUST 含 eventId、nextAt、title；ContentState MUST 含 title 与 nextAt（unix 秒）；
// 载荷 MUST NOT 含 logo/color/logoFile/远程图 URL（视觉由客户端按 eventId 本地解析）。
func startImminentLiveActivityForWx(ctx context.Context, wxID int64, deviceNo, alert string, eventID, nextAt int64) {
	if wxID <= 0 || !predictLiveActivityEnabled() {
		return
	}
	activityID := imminentActivityID(wxID)
	// 单槽换新：先 end 旧再 start。旧槽 title/nextAt 未缓存则省略，禁止写 0。
	endImminentSlotForWx(ctx, wxID, "", 0)

	// 固定 Attributes：含 eventId 取色/图、title 供 Push-to-Start 后锁屏标题回退；nextAt 为 unix 秒；禁止 logo/color。
	attrs := map[string]interface{}{
		"activityId": activityID,
		"deviceNo":   deviceNo,
		"eventId":    eventID,
		"nextAt":     nextAt,
		"title":      strings.TrimSpace(alert),
	}
	// ContentState.nextAt 与 Attributes 同为 unix 秒，供锁屏倒计时。
	cs := map[string]interface{}{
		"title":  alert,
		"phase":  "imminent",
		"nextAt": nextAt,
	}
	if pushErr := pushclient.FanoutLiveActivity(ctx, pushclient.LiveActivityOp{
		WxID:           wxID,
		Action:         "start",
		AttributesKind: "imminent",
		ActivityID:     activityID,
		Attributes:     attrs,
		ContentState:   cs,
		AlertTitle:     "预测临近",
		AlertBody:      alert,
	}); pushErr != nil {
		glog.Warningf(ctx, "[predict-la] start_imminent_failed wxId=%d err=%v", wxID, pushErr)
		return
	}
	slotKey, err := cachekit.PredictImminentActivitySlotKey(wxID)
	if err != nil {
		return
	}
	_ = predictImminentCache.Set(ctx, slotKey, activityID)
}
