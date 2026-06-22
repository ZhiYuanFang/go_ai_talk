package runtimecheck

import (
	"context"
	"fmt"
	"hello/internal/platform/cachekit"
	"hello/internal/platform/eventkit"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gogf/gf/v2/os/glog"
)

const (
	mqAPIEnv  = "MQ_HTTP_API_BASE"
	mqUserEnv = "MQ_USER"
	mqPassEnv = "MQ_PASSWORD"
)

// DependencyOptions 启动探活选项。生产容灾下 API 进程可将 RequireRabbitMQ 设为 false，
// 使 RabbitMQ 短暂不可达时仍可启动；worker 等消费者进程应保持 true。
type DependencyOptions struct {
	RequireRabbitMQ bool
}

// CheckDependencies 在服务启动前执行关键依赖探测。
// Redis 始终 fail-fast；RabbitMQ 是否阻断启动由 opts.RequireRabbitMQ 决定。
func CheckDependencies(ctx context.Context, opts DependencyOptions) error {
	cache := cachekit.Default()
	if err := pingRedisSafe(ctx, cache); err != nil {
		return fmt.Errorf("redis dependency check failed: %w", err)
	}
	return checkRabbitMQ(ctx, opts.RequireRabbitMQ)
}

func checkRabbitMQ(ctx context.Context, required bool) error {
	apiBase := strings.TrimSpace(os.Getenv(mqAPIEnv))
	if apiBase == "" {
		if required {
			return fmt.Errorf("rabbitmq dependency check failed: %w", eventkit.ErrEmptyAPIBase)
		}
		glog.Warning(ctx, "metric=rabbitmq_startup_degraded reason=MQ_HTTP_API_BASE_empty startup_check=skipped")
		return nil
	}
	publisher, err := eventkit.NewHTTPPublisher(eventkit.HTTPPublisherConfig{
		APIBase:  apiBase,
		User:     strings.TrimSpace(os.Getenv(mqUserEnv)),
		Password: strings.TrimSpace(os.Getenv(mqPassEnv)),
		Exchange: eventkit.DefaultExchange,
	})
	if err != nil {
		if required {
			return fmt.Errorf("rabbitmq dependency check failed: %w", err)
		}
		glog.Warningf(ctx, "metric=rabbitmq_startup_degraded reason=publisher_init err=%v", err)
		return nil
	}
	if err = eventkit.WithObserver(publisher, eventkit.LoggingObserver{}).CheckDependency(ctx); err != nil {
		if required {
			return fmt.Errorf("rabbitmq dependency check failed: %w", err)
		}
		glog.Warningf(ctx, "metric=rabbitmq_startup_degraded reason=connectivity_check err=%v", err)
		return nil
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
