package runtimecheck

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"hello/internal/platform/cachekit"
	"hello/internal/platform/eventkit"
)

const (
	mqAPIEnv  = "MQ_HTTP_API_BASE"
	mqUserEnv = "MQ_USER"
	mqPassEnv = "MQ_PASSWORD"
)

// CheckDependencies 在服务启动前执行关键依赖探测。
// 任一依赖失败都返回错误，调用方应直接终止启动（fail-fast）。
func CheckDependencies(ctx context.Context) error {
	cache := cachekit.WithObserver(cachekit.NewRedisCache(), cachekit.LoggingObserver{})
	if err := pingRedisSafe(ctx, cache); err != nil {
		return fmt.Errorf("redis dependency check failed: %w", err)
	}

	// RabbitMQ 通过管理 API 做连通性探测，确保后续发布路径可用。
	publisher, err := eventkit.NewHTTPPublisher(eventkit.HTTPPublisherConfig{
		APIBase:  strings.TrimSpace(os.Getenv(mqAPIEnv)),
		User:     strings.TrimSpace(os.Getenv(mqUserEnv)),
		Password: strings.TrimSpace(os.Getenv(mqPassEnv)),
		Exchange: eventkit.DefaultExchange,
	})
	if err != nil {
		return fmt.Errorf("rabbitmq dependency check failed: %w", err)
	}
	if err = eventkit.WithObserver(publisher, eventkit.LoggingObserver{}).CheckDependency(ctx); err != nil {
		return fmt.Errorf("rabbitmq dependency check failed: %w", err)
	}
	return nil
}

func pingRedisSafe(ctx context.Context, cache cachekit.Cache) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("redis config missing or invalid: %v", r)
			_ = debug.Stack()
		}
	}()
	return cache.Ping(ctx)
}

