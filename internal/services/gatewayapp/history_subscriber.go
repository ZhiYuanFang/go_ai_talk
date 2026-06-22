package gatewayapp

import (
	"context"
	"strings"

	"hello/internal/platform/redismsgkit"

	"github.com/gogf/gf/v2/encoding/gjson"
)

// StartHistoryNotifySubscriber 在后台订阅 Redis 频道并向 WS Hub 广播（进程级单例 goroutine）。
func StartHistoryNotifySubscriber(ctx context.Context) {
	redismsgkit.StartSubscriber(ctx, redismsgkit.ChannelAppHistoryNotify, func(ctx context.Context, payload string) error {
		deviceNo := strings.TrimSpace(gjson.New(payload).Get("device_no").String())
		if deviceNo == "" {
			return nil
		}
		HistoryHub().BroadcastText(ctx, deviceNo, payload)
		return nil
	})
}
