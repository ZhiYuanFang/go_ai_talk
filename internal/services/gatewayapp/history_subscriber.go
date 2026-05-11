package gatewayapp

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/redis/go-redis/v9"
)

const historyNotifyChannel = "app:history:notify"

// StartHistoryNotifySubscriber 在后台订阅 Redis 频道并向 WS Hub 广播（进程级单例 goroutine）。
func StartHistoryNotifySubscriber(ctx context.Context) {
	go runHistorySubscriber(ctx)
}

func runHistorySubscriber(ctx context.Context) {
	addrStr := strings.TrimSpace(g.Cfg().MustGet(ctx, "redis.default.address").String())
	if addrStr == "" {
		glog.Errorf(ctx, "[gateway-app-sub] redis.default.address 未配置，跳过订阅")
		return
	}
	addrs := strings.Split(addrStr, ",")
	for i := range addrs {
		addrs[i] = strings.TrimSpace(addrs[i])
	}
	var subConn *redis.PubSub
	if len(addrs) > 1 {
		rdb := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: addrs,
		})
		subConn = rdb.Subscribe(ctx, historyNotifyChannel)
	} else {
		rdb := redis.NewClient(&redis.Options{Addr: addrs[0]})
		subConn = rdb.Subscribe(ctx, historyNotifyChannel)
	}
	defer func() { _ = subConn.Close() }()
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
			// 直接透传 JSON 字符串给前端。
			deviceNo := strings.TrimSpace(gjson.New(msg.Payload).Get("device_no").String())
			if deviceNo == "" {
				continue
			}
			HistoryHub().BroadcastText(ctx, deviceNo, msg.Payload)
		}
	}
}
