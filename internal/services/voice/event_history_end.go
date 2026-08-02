package voice

import (
	"context"
	"fmt"

	"hello/internal/model/entity"
)

// applyVoiceEventEndHistory 语音结束动作落库。
// 以 history-service EndLatestHistoryIfMatch.updated 为权威：按同 eventId 闭合最近未结束行
//（允许中间夹杂其它事件），避免 GetLatest 预判与 DB 不一致时误播报且无 WS 推送。
// 无未闭合同 event 时降级瞬时 AddHistory；若降级前全局最新为其它未结束事件，可顺带尝试闭合。
func applyVoiceEventEndHistory(ctx context.Context, deviceNo string, event entity.Event, displayTargetName, remark string, nowTime int64) (reply string, err error) {
	rowName := historyRowEventName(event, displayTargetName)

	// 优先闭合未结束的同 event 记录；命中则不再新建，解决「睡眠未结束却新建瞬时睡眠」。
	updated, err := DeviceHistory().EndLatestHistoryIfMatch(ctx, deviceNo, event.Id, nowTime, remark)
	if err != nil {
		return "更新结束时间失败,请重试", err
	}
	if updated {
		return fmt.Sprintf("好的，已记录%s结束", displayTargetName), nil
	}

	// 无未闭合同 event：降级为瞬时结束（history-service AddHistory → publish create）
	lastEvent, _ := DeviceHistory().GetLatestHistory(ctx, deviceNo)
	_, err = DeviceHistory().AddHistory(ctx, entity.History{
		DeviceNo:  deviceNo,
		EventId:   event.Id,
		EventName: rowName,
		EventUnit: historyRowEventUnit(event),
		StartTime: nowTime,
		EndTime:   nowTime,
		Remark:    remark,
	})
	if err != nil {
		return "记录事件失败,请重试", err
	}

	if lastEvent.EndTime == 0 && lastEvent.EventId > 0 && lastEvent.EventId != event.Id {
		prevClosed, closeErr := DeviceHistory().EndLatestHistoryIfMatch(ctx, deviceNo, lastEvent.EventId, nowTime, "")
		if closeErr != nil {
			return fmt.Sprintf("好的，已记录%s结束，%s结束失败,请手动结束", displayTargetName, lastEvent.EventName), closeErr
		}
		if !prevClosed && lastEvent.Id > 0 {
			lastEvent.EndTime = nowTime
			lastEvent.DeviceNo = deviceNo
			if uErr := DeviceHistory().UpdateHistory(ctx, lastEvent); uErr != nil {
				return fmt.Sprintf("好的，已记录%s结束，%s结束失败,请手动结束", displayTargetName, lastEvent.EventName), uErr
			}
		}
		return fmt.Sprintf("好的，已记录%s结束，%s自动结束", displayTargetName, lastEvent.EventName), nil
	}

	return fmt.Sprintf("好的，已记录%s结束", displayTargetName), nil
}
