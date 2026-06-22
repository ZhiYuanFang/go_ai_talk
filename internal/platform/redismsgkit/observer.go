package redismsgkit

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/os/glog"
)

type PublishObserver interface {
	OnPublish(ctx context.Context, channel string, duration time.Duration, err error)
}

type SubscribeObserver interface {
	OnSubscribeMessage(ctx context.Context, channel string, err error)
}

type NoopObserver struct{}

func (NoopObserver) OnPublish(_ context.Context, _ string, _ time.Duration, _ error) {}
func (NoopObserver) OnSubscribeMessage(_ context.Context, _ string, _ error)         {}

type LoggingObserver struct{}

func (LoggingObserver) OnPublish(ctx context.Context, channel string, duration time.Duration, err error) {
	if err != nil {
		glog.Warningf(ctx, "redismsgkit publish failed channel=%s duration=%s err=%v", channel, duration, err)
		return
	}
	glog.Debugf(ctx, "redismsgkit publish success channel=%s duration=%s", channel, duration)
}

func (LoggingObserver) OnSubscribeMessage(ctx context.Context, channel string, err error) {
	if err != nil {
		glog.Warningf(ctx, "redismsgkit subscribe handler failed channel=%s err=%v", channel, err)
	}
}
