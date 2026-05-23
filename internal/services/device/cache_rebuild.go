package device

import (
	"context"
	"strings"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/os/glog"
)

// refreshEventOptionsCacheAfterMutate 事件表变更后从 DB 重建 Redis 事件字典；勿用 ListEvents，避免读旧缓存。
func refreshEventOptionsCacheAfterMutate(ctx context.Context) {
	if err := RebuildEventCache(ctx); err != nil {
		glog.Warningf(ctx, "[device-admin] 重建事件 Redis 缓存失败 err=%v", err)
	}
}

// RebuildEventCache 重建事件缓存快照。
func RebuildEventCache(ctx context.Context) error {
	rows := make([]entity.Event, 0)
	if err := dao.Event.Ctx(ctx).Fields(eventListFields()...).OrderAsc(dao.Event.Columns().Id).Scan(&rows); err != nil {
		return err
	}
	normalizeEventRows(rows)
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
