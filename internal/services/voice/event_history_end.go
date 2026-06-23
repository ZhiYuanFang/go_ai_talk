package voice

import (
	"context"
	"fmt"

	"hello/internal/model/entity"
)

// applyVoiceEventEndHistory 语音结束动作落库。
// 以 history-service EndLatestHistoryIfMatch.updated 为权威，避免 voice 侧 GetLatestHistory
// 缓存预判与 DB 不一致时 updated=false 仍播报成功且无 app:history:notify WS 推送。
// EndLatest 未匹配时降级瞬时 AddHistory，并在需要时自动闭合上一条未结束计时。
func applyVoiceEventEndHistory(ctx context.Context, deviceNo string, event entity.Event, displayTargetName, remark string, nowTime int64) (reply string, err error) {
	rowName := historyRowEventName(event, displayTargetName)

	updated, err := DeviceHistory().EndLatestHistoryIfMatch(ctx, deviceNo, event.Id, nowTime, remark)
	if err != nil {
		return "更新结束时间失败,请重试", err
	}
	if updated {
		return fmt.Sprintf("好的，已记录%s结束", displayTargetName), nil
	}

	// EndLatest 未匹配：降级为瞬时结束（history-service AddHistory → publish create）
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
