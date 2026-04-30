package eventkit

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/os/glog"
)

type PublishObserver interface {
	OnPublish(ctx context.Context, routingKey string, duration time.Duration, err error)
}

type NoopObserver struct{}

func (NoopObserver) OnPublish(_ context.Context, _ string, _ time.Duration, _ error) {}

type LoggingObserver struct{}

func (LoggingObserver) OnPublish(ctx context.Context, routingKey string, duration time.Duration, err error) {
	// 失败与成功分级输出，便于在低日志级别下聚焦异常发布链路。
	if err != nil {
		glog.Warningf(ctx, "eventkit publish failed routingKey=%s duration=%s err=%v", routingKey, duration, err)
		return
	}
	glog.Debugf(ctx, "eventkit publish success routingKey=%s duration=%s", routingKey, duration)
}

type observedPublisher struct {
	base     Publisher
	observer PublishObserver
}

func WithObserver(base Publisher, observer PublishObserver) Publisher {
	if observer == nil {
		observer = NoopObserver{}
	}
	return &observedPublisher{
		base:     base,
		observer: observer,
	}
}

func (p *observedPublisher) CheckDependency(ctx context.Context) error {
	return p.base.CheckDependency(ctx)
}

func (p *observedPublisher) Publish(ctx context.Context, routingKey string, payload any) error {
	begin := time.Now()
	err := p.base.Publish(ctx, routingKey, payload)
	// 统一打点发布耗时，便于观察下游 MQ 管理 API 抖动。
	p.observer.OnPublish(ctx, routingKey, time.Since(begin), err)
	return err
}

