package history

import (
	"context"
	"strings"

	"hello/internal/dao"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// ReconcileProjectionCachesForWorker 由 worker 经 HTTP 触发：仅在 history-service 进程内访问本库 history 表并完成读模型/生日/meta 缓存修复。
func ReconcileProjectionCachesForWorker(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := g.DB(dao.History.Group()).Model(dao.History.Table()).
		Fields(dao.History.Columns().DeviceNo).
		Where(dao.History.Columns().DeviceNo + " <> ''").
		OrderDesc(dao.History.Columns().Id).
		Limit(limit).
		All()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(rows))
	deviceNos := make([]string, 0, len(rows))
	for _, row := range rows {
		deviceNo := strings.TrimSpace(row[dao.History.Columns().DeviceNo].String())
		if deviceNo == "" {
			continue
		}
		if _, ok := seen[deviceNo]; ok {
			continue
		}
		seen[deviceNo] = struct{}{}
		deviceNos = append(deviceNos, deviceNo)
	}
	for _, deviceNo := range deviceNos {
		if err := RebuildHistoryCacheByDevice(ctx, deviceNo); err != nil {
			glog.Warningf(ctx, "history cache rebuild failed: deviceNo=%s err=%v", deviceNo, err)
		}
		if err := RebuildBirthdayCacheByDevice(ctx, deviceNo); err != nil {
			glog.Warningf(ctx, "birthday cache rebuild failed: deviceNo=%s err=%v", deviceNo, err)
		}
	}
	_ = RebuildHistoryMetaCache(ctx)
	return deviceNos, nil
}
