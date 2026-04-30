package eventkit

import (
	"fmt"
	"strings"
)

const (
	DefaultExchange = "voice.events"
)

// RoutingKey 统一路由键格式：domain.bounded-context.event。
func RoutingKey(domain, boundedContext, event string) (string, error) {
	domain = normalize(domain)
	boundedContext = normalize(boundedContext)
	event = normalize(event)
	if domain == "" || boundedContext == "" || event == "" {
		return "", ErrEmptyRoutingKey
	}
	return fmt.Sprintf("%s.%s.%s", domain, boundedContext, event), nil
}

func normalize(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	// 使用短横线替代空格，避免路由键包含不稳定空白。
	v = strings.ReplaceAll(v, " ", "-")
	return v
}

