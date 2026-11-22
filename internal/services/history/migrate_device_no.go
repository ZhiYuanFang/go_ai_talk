package history

import (
	"context"
	"fmt"
	"strings"

	"hello/internal/dao"
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// MigrateDeviceNo 将 history / suggest 表中 device_no 从 old 更新为 new（宝宝号原地改名）。
// 业务逻辑：一次 UPDATE；best-effort 删除旧号 Redis 列表/latest/version；不跟迁审计快照以外的表。
// Args: oldDeviceNo / newDeviceNo 均须非空且不相等。
// Returns: 错误；成功时旧号下业务行应已迁至新号。
func MigrateDeviceNo(ctx context.Context, oldDeviceNo, newDeviceNo string) error {
	old := strings.TrimSpace(oldDeviceNo)
	neu := strings.TrimSpace(newDeviceNo)
	if old == "" || neu == "" {
		return fmt.Errorf("deviceNo 不能为空")
	}
	if old == neu {
		return fmt.Errorf("新旧 deviceNo 不能相同")
	}
	db := g.DB(dao.History.Group())
	if _, err := db.Model(dao.History.Table()).Ctx(ctx).
		Data(dao.History.Columns().DeviceNo, neu).
		Where(dao.History.Columns().DeviceNo, old).
		Update(); err != nil {
		return fmt.Errorf("migrate history: %w", err)
	}
	if _, err := db.Model(dao.Suggest.Table()).Ctx(ctx).
		Data(dao.Suggest.Columns().DeviceNo, neu).
		Where(dao.Suggest.Columns().DeviceNo, old).
		Update(); err != nil {
		return fmt.Errorf("migrate suggest: %w", err)
	}
	// best-effort：删旧号读缓存，避免脏列表；失败仅打日志。
	delHistoryRedisBestEffort(ctx, old)
	bumpPieceCacheEpoch(ctx, neu)
	glog.Infof(ctx, "[history] MigrateDeviceNo ok oldLen=%d newLen=%d", len(old), len(neu))
	return nil
}

func delHistoryRedisBestEffort(ctx context.Context, deviceNo string) {
	c := cachekit.Default()
	for _, keyFn := range []func(string) (string, error){
		cachekit.HistoryListKey,
		cachekit.HistoryLatestKey,
		cachekit.HistoryVersionKey,
	} {
		key, err := keyFn(deviceNo)
		if err != nil || key == "" {
			continue
		}
		if err := c.Del(ctx, key); err != nil {
			glog.Warningf(ctx, "[history] del redis after migrate failed key=%s err=%v", key, err)
		}
	}
}
