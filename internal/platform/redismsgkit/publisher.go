package redismsgkit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type Publisher interface {
	Publish(ctx context.Context, channel, payload string) error
}

type redisPublisher struct{}

func NewPublisher() Publisher {
	return &redisPublisher{}
}

// 发布一个消息到指定的频道(频道名:channel,消息内容:payload)
// eg. 发布一个消息到指定的频道(频道名:user:123,消息内容:张三)
// redisPublisher.Publish(ctx, "user:123", "张三")
// 返回: nil
func (p *redisPublisher) Publish(ctx context.Context, channel, payload string) error {
	channel = strings.TrimSpace(channel)
	if channel == "" {
		return ErrInvalidChannel
	}
	if strings.TrimSpace(payload) == "" {
		return ErrEmptyPayload
	}
	if _, err := g.Redis().Do(ctx, "PUBLISH", channel, payload); err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return nil
}

type observedPublisher struct {
	base     Publisher
	observer PublishObserver
}

func WithPublisherObserver(base Publisher, observer PublishObserver) Publisher {
	if observer == nil {
		observer = NoopObserver{}
	}
	return &observedPublisher{base: base, observer: observer}
}

func (p *observedPublisher) Publish(ctx context.Context, channel, payload string) error {
	begin := time.Now()
	err := p.base.Publish(ctx, channel, payload)               // 发布消息
	p.observer.OnPublish(ctx, channel, time.Since(begin), err) // 统一打日志
	return err
}

// DefaultPublisher 带 LoggingObserver 的全局 Publisher。
func DefaultPublisher() Publisher {
	return WithPublisherObserver(NewPublisher(), LoggingObserver{})
}
