// Imminent Live Activity 清屏：挂 history 权威写成功路径（增/删/改），
// 禁止依赖 voice pending PUT；经 clients/push 扇出 end，经 cachekit 作废 Redis pending。
// MUST NOT import services/voice；MUST NOT 借此 end Timing。
package history

import (
	"context"
	"fmt"
	"strings"

	deviceclient "hello/internal/clients/device"
	pushclient "hello/internal/clients/push"
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/os/glog"
)

// imminentActivityID 与 voice 侧约定一致：每 wx 单槽权威 id。
func imminentActivityID(wxID int64) string {
	return fmt.Sprintf("imminent:%d", wxID)
}

// clearImminentOnHistoryMutation history 落库成功后的 Imminent 清屏 + 作废日程。
//
// 业务：
//  1. 一律清：不区分事件类型、瞬时/计时、Update 字段集合；
//  2. 向绑定用户扇出 Imminent end（开关关则跳过 APNs，仍清 Redis pending）；
//  3. Del per-wx Imminent 槽键；Del 该 deviceNo 的 imminent pending 权威键；
//  4. MUST NOT end Timing；失败仅 WARN，不回滚已提交的 history。
//
// Side Effects: push live-activity end；Del Redis slot / pending。
func clearImminentOnHistoryMutation(ctx context.Context, deviceNo string) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return
	}

	// 日程作废与 LA 开关解耦：历史一变，旧 pending 即失效（旧延时 MQ 靠消费对账不匹配）。
	if pendingKey, keyErr := cachekit.PredictImminentPendingKey(deviceNo); keyErr != nil {
		glog.Warningf(ctx, "[history-la] clear pending_key_err deviceNoLen=%d err=%v", len(deviceNo), keyErr)
	} else if delErr := historyLACache.Del(ctx, pendingKey); delErr != nil {
		glog.Warningf(ctx, "[history-la] clear pending_del_err deviceNoLen=%d err=%v", len(deviceNo), delErr)
	}

	if !historyLiveActivityEnabled() {
		return
	}

	wxIDs, _, err := deviceclient.RemoteListWxIDsByDeviceNo(ctx, deviceNo)
	if err != nil {
		glog.Warningf(ctx, "[history-la] clear list_wx_err deviceNoLen=%d err=%v", len(deviceNo), err)
		return
	}
	for _, wxID := range wxIDs {
		endImminentSlotForWx(ctx, wxID)
	}
}

// endImminentSlotForWx 结束该用户 Imminent 槽并清 Redis 槽跟踪。
// ContentState：phase=cleared；清屏无单槽 title/nextAt 缓存，省略字段而非写 0。
// dismissal-date 由 push-service 组装层填过去秒；activity token 行由 push end 路径清理。
func endImminentSlotForWx(ctx context.Context, wxID int64) {
	if wxID <= 0 {
		return
	}
	slotKey, err := cachekit.PredictImminentActivitySlotKey(wxID)
	if err != nil {
		return
	}
	activityID := imminentActivityID(wxID)
	if val, ok, _ := historyLACache.Get(ctx, slotKey); ok && strings.TrimSpace(val) != "" {
		activityID = strings.TrimSpace(val)
	}
	cs := map[string]interface{}{"phase": "cleared"}
	if pushErr := pushclient.FanoutLiveActivity(ctx, pushclient.LiveActivityOp{
		WxID:           wxID,
		Action:         "end",
		AttributesKind: "imminent",
		ActivityID:     activityID,
		ContentState:   cs,
	}); pushErr != nil {
		glog.Warningf(ctx, "[history-la] end_imminent_failed wxId=%d err=%v", wxID, pushErr)
	}
	_ = historyLACache.Del(ctx, slotKey)
}
