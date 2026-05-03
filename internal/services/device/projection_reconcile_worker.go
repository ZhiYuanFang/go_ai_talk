package device

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/os/glog"
)

// ReconcileProjectionCachesForWorker 由 worker 经 HTTP 触发：刷新画像缓存及全量事件/动作缓存快照。
func ReconcileProjectionCachesForWorker(ctx context.Context, deviceNos []string) error {
	for _, raw := range deviceNos {
		deviceNo := strings.TrimSpace(raw)
		if deviceNo == "" {
			continue
		}
		if err := RebuildUserProfileCacheByDevice(ctx, deviceNo); err != nil {
			glog.Warningf(ctx, "profile cache rebuild failed: deviceNo=%s err=%v", deviceNo, err)
		}
	}
	if err := RebuildEventCache(ctx); err != nil {
		glog.Warningf(ctx, "event cache rebuild failed: err=%v", err)
	}
	if err := RebuildActionCache(ctx); err != nil {
		glog.Warningf(ctx, "action cache rebuild failed: err=%v", err)
	}
	return nil
}
