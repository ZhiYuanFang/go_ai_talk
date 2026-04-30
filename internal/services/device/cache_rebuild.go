package device

import (
	"context"
	"strings"

	"hello/internal/dao"
	"hello/internal/model/entity"
)

// RebuildEventCache 重建事件缓存快照。
func RebuildEventCache(ctx context.Context) error {
	rows := make([]entity.Event, 0)
	if err := dao.Event.Ctx(ctx).Fields(
		dao.Event.Columns().Id,
		dao.Event.Columns().Name,
		dao.Event.Columns().NeedQuantity,
		dao.Event.Columns().ExtraNames,
	).OrderAsc(dao.Event.Columns().Id).Scan(&rows); err != nil {
		return err
	}
	return deviceCache.setEventOptions(ctx, rows)
}

// RebuildActionCache 重建动作缓存快照。
func RebuildActionCache(ctx context.Context) error {
	rows := make([]entity.Action, 0)
	if err := dao.Action.Ctx(ctx).Fields(
		dao.Action.Columns().Id,
		dao.Action.Columns().Name,
		dao.Action.Columns().TargetType,
	).OrderAsc(dao.Action.Columns().Id).Scan(&rows); err != nil {
		return err
	}
	return deviceCache.setActionOptions(ctx, rows)
}

// RebuildUserProfileCacheByDevice 按设备重建设备画像缓存。
func RebuildUserProfileCacheByDevice(ctx context.Context, deviceNo string) error {
	profile, err := DeviceProfile().GetProfile(ctx, strings.TrimSpace(deviceNo))
	if err != nil {
		return err
	}
	return deviceCache.setUserProfile(ctx, cachedUserProfile{
		DeviceNo: profile.DeviceNo,
		Birthday: profile.Birthday,
		Sex:      profile.Sex,
	})
}
