package ucg

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/os/glog"
)

// StartAuditWorker 启动 Green 异步审核轮询（帖子 pending_audit + 资料 Redis 队列）。
func StartAuditWorker(ctx context.Context) {
	cfg := LoadGreenConfig(ctx)
	interval := time.Duration(cfg.AuditIntervalSeconds) * time.Second
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runAuditTick(ctx)
			}
		}
	}()
	glog.Infof(ctx, "ucg green audit worker started, interval=%s, greenEnabled=%v", interval, cfg.Enabled)
}

func runAuditTick(ctx context.Context) {
	posts, err := listPendingAuditPosts(ctx, 20)
	if err != nil {
		glog.Warningf(ctx, "ucg audit: list pending posts failed: %v", err)
	} else {
		for _, post := range posts {
			if err = auditPost(ctx, post); err != nil {
				glog.Warningf(ctx, "ucg audit: post %d failed: %v", post.Id, err)
			}
		}
	}
	wxIDs, err := listPendingProfileWxIDs(ctx, 20)
	if err != nil {
		glog.Warningf(ctx, "ucg audit: list pending profiles failed: %v", err)
		return
	}
	for _, wxID := range wxIDs {
		if err = auditProfilePatch(ctx, wxID); err != nil {
			glog.Warningf(ctx, "ucg audit: profile wxId=%d failed: %v", wxID, err)
		}
	}
}
