package cachekit

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/os/glog"
)

type CacheObserver interface {
	OnOperation(ctx context.Context, operation, key string, duration time.Duration, err error)
}

type NoopObserver struct{}

func (NoopObserver) OnOperation(_ context.Context, _ string, _ string, _ time.Duration, _ error) {}

type LoggingObserver struct{}

func (LoggingObserver) OnOperation(ctx context.Context, operation, key string, duration time.Duration, err error) {
	// 失败走 warning，成功走 debug，避免高频缓存日志污染默认日志级别。
	if err != nil {
		glog.Warningf(ctx, "cachekit %s failed key=%s duration=%s err=%v", operation, key, duration, err)
		return
	}
	glog.Debugf(ctx, "cachekit %s success key=%s duration=%s", operation, key, duration)
}

