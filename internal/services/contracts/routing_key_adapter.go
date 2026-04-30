package contracts

import (
	"context"
	"fmt"
	"strings"

	"hello/internal/platform/eventkit"

	"github.com/gogf/gf/v2/os/glog"
)

// ParseRoutingKey 在 contracts/service 边界统一路由键解析。
func ParseRoutingKey(ctx context.Context, raw string, source string) (eventkit.RouteKey, bool) {
	key, ok := eventkit.ParseRoutingKey(raw)
	if !ok {
		glog.Warningf(ctx, "metric=invalid_routing_key_reject source=%s routingKey=%s", strings.TrimSpace(source), strings.TrimSpace(raw))
	}
	return key, ok
}

// ParseRoutingKeyCompat 兼容旧字符串入口，迁移期输出弃用告警并走统一解析。
func ParseRoutingKeyCompat(ctx context.Context, raw string, source string) (eventkit.RouteKey, error) {
	glog.Warningf(ctx, "deprecated routing key entry source=%s routingKey=%s", strings.TrimSpace(source), strings.TrimSpace(raw))
	key, ok := ParseRoutingKey(ctx, raw, source)
	if !ok {
		return "", fmt.Errorf("invalid routing key: %s", strings.TrimSpace(raw))
	}
	return key, nil
}
