package history

import (
	"context"
	"strings"

	"hello/internal/model/entity"
)

// RebuildHistoryCacheByDevice 按设备重建历史读模型缓存。
func RebuildHistoryCacheByDevice(ctx context.Context, deviceNo string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil
	}
	items, err := ListDeviceHistory(ctx, deviceNo)
	if err != nil {
		return err
	}
	if err = historyCache.setHistoryList(ctx, deviceNo, items); err != nil {
		return err
	}
	if len(items) > 0 {
		return historyCache.setLatestHistory(ctx, deviceNo, items[0])
	}
	return nil
}

// RebuildHistoryMetaCache 重建历史相关元数据缓存（事件选项与生日）。
func RebuildHistoryMetaCache(ctx context.Context) error {
	items, err := ListEventOptions(ctx)
	if err != nil {
		return err
	}
	return historyCache.setEventOptions(ctx, items)
}

// RebuildBirthdayCacheByDevice 按设备重建生日画像缓存。
func RebuildBirthdayCacheByDevice(ctx context.Context, deviceNo string) error {
	birthday, sex, err := GetDeviceBirthday(ctx, deviceNo)
	if err != nil {
		return err
	}
	return historyCache.setBirthday(ctx, strings.TrimSpace(deviceNo), birthday, sex)
}

// BuildHistoryEntityFromProjection 从投影事件构建历史实体。
func BuildHistoryEntityFromProjection(evt historyProjectionEvent) entity.History {
	return entity.History{
		Id:          evt.HistoryID,
		DeviceNo:    strings.TrimSpace(evt.DeviceNo),
		EventId:     evt.EventIDRef,
		EventName:   strings.TrimSpace(evt.EventName),
		EventNumber: evt.EventNum,
		StartTime:   strings.TrimSpace(evt.StartTime),
		EndTime:     strings.TrimSpace(evt.EndTime),
		Remark:      strings.TrimSpace(evt.Remark),
	}
}
