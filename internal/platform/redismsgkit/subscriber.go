package redismsgkit

import (
	"context"
	"strings"
	"time"

	"hello/internal/platform/rediscfg"

	"github.com/gogf/gf/v2/os/glog"
	"github.com/redis/go-redis/v9"
)

// MessageHandler 处理订阅到的单条消息 payload。
type MessageHandler func(ctx context.Context, payload string) error

type observedSubscriber struct {
	observer SubscribeObserver
}

// RunSubscriber 在调用 goroutine 内阻塞订阅 channel；cluster 地址逗号分隔时走 go-redis Cluster。
func RunSubscriber(ctx context.Context, channel string, handler MessageHandler, observer SubscribeObserver) {
	channel = strings.TrimSpace(channel)
	if channel == "" || handler == nil {
		return
	}
	if observer == nil {
		observer = NoopObserver{}
	}
	obs := &observedSubscriber{observer: observer}

	// 与 rediscfg.ApplyDefaultFromEnv / g.Redis() 共用 GF_REDIS_DEFAULT_ADDRESS（yaml 已无 redis 段）。
	addrStr := rediscfg.DefaultAddressFromEnv()
	if addrStr == "" {
		glog.Errorf(ctx, "[redismsgkit] GF_REDIS_DEFAULT_ADDRESS 未配置，跳过订阅 channel=%s", channel)
		return
	}
	addrs := strings.Split(addrStr, ",")
	for i := range addrs {
		addrs[i] = strings.TrimSpace(addrs[i])
	}
	//  Redis 的 订阅对象 连接
	var subConn *redis.PubSub
	if len(addrs) > 1 {
		rdb := redis.NewClusterClient(&redis.ClusterOptions{Addrs: addrs})
		subConn = rdb.Subscribe(ctx, channel)
	} else {
		rdb := redis.NewClient(&redis.Options{Addr: addrs[0]})
		subConn = rdb.Subscribe(ctx, channel)
	}
	defer func() { _ = subConn.Close() }()
	// 开始接收消息(channel:频道名,msg:消息内容)
	ch := subConn.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				time.Sleep(time.Second)
				continue
			}
			if msg == nil {
				continue
			}
			if err := handler(ctx, msg.Payload); err != nil {
				obs.observer.OnSubscribeMessage(ctx, channel, err)
			}
		}
	}
}

// StartSubscriber 后台 goroutine 订阅。
func StartSubscriber(ctx context.Context, channel string, handler MessageHandler) {
	go RunSubscriber(ctx, channel, handler, LoggingObserver{})
}
