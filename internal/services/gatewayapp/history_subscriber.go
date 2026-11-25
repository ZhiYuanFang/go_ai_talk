package gatewayapp

import (
	"context"
	"strings"

	"hello/internal/platform/redismsgkit"

	"github.com/gogf/gf/v2/encoding/gjson"
)

// StartHistoryNotifySubscriber 在后台订阅 Redis 频道并向 WS Hub 与 SSE Hub 广播（进程级单例 goroutine）。
// WS 仍 BroadcastText 原文；SSE 另写 event/data，互不改写对方载荷。
func StartHistoryNotifySubscriber(ctx context.Context) {
	redismsgkit.StartSubscriber(ctx, redismsgkit.ChannelAppHistoryNotify, func(ctx context.Context, payload string) error {
		deviceNo := strings.TrimSpace(gjson.New(payload).Get("device_no").String())
		if deviceNo == "" {
			return nil
		}
		HistoryHub().BroadcastText(ctx, deviceNo, payload)
		HistorySSE().Broadcast(ctx, deviceNo, payload)
		MerchantSSE().BroadcastDevice(ctx, deviceNo, payload)
		return nil
	})
}
