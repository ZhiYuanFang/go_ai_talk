// Package ucg AMQP 连接与 handler 桥接：eventkit 收到 delivery → ucgAuditAMQPHandler → dispatchUcgAuditPayload。
package ucg

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	rabbitAMQPURLEnv    = "RABBITMQ_AMQP_URL"   // 完整 amqp:// URL，优先于分项 env
	rabbitHostEnv       = "RABBITMQ_HOST"       // docker-compose 默认 rabbitmq
	rabbitAMQPPortEnv   = "RABBITMQ_AMQP_PORT"  // 默认 5672
	mqUserEnv           = "MQ_USER"
	mqPasswordEnv       = "MQ_PASSWORD"
	ucgAuditPrefetchEnv = "UCG_AUDIT_MQ_PREFETCH" // 单 consumer 未 ack 上限；prefetch=5 时最多 5 条并发处理
)

// ucgAuditAMQPHandler eventkit AMQPMessageHandler 实现；返回值决定 Ack/Nack。
// nil → Ack 删消息；非 nil → eventkit handleDelivery Nack(requeue=true) → 消息回队 → 可能再调 Green。
func ucgAuditAMQPHandler(ctx context.Context, queueName, routingKey string, body []byte) error {
	_ = routingKey // UCG 审核按 queueName 路由，不用 routingKey
	return dispatchUcgAuditPayload(ctx, queueName, string(body))
}

// ucgAuditAMQPURL 解析 RabbitMQ 连接串；失败时 StartUcgMQConsumers 不启动 consumer。
func ucgAuditAMQPURL() (string, error) {
	if u := strings.TrimSpace(os.Getenv(rabbitAMQPURLEnv)); u != "" {
		return u, nil
	}
	host := strings.TrimSpace(os.Getenv(rabbitHostEnv))
	if host == "" {
		host = "rabbitmq"
	}
	port := 5672
	if v := strings.TrimSpace(os.Getenv(rabbitAMQPPortEnv)); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			port = n
		}
	}
	user := strings.TrimSpace(os.Getenv(mqUserEnv))
	if user == "" {
		user = "guest"
	}
	pass := os.Getenv(mqPasswordEnv)
	if strings.TrimSpace(pass) == "" {
		pass = "guest"
	}
	u := &url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(user, pass),
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   "/",
	}
	return u.String(), nil
}

// ucgAuditPrefetch 审核队列 prefetch；风暴时多条 unacked 会并行重复调 Green。
func ucgAuditPrefetch() int {
	n := 5
	if v := strings.TrimSpace(os.Getenv(ucgAuditPrefetchEnv)); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			n = p
		}
	}
	return n
}

func redactAMQPURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "amqp://***"
	}
	if u.User != nil {
		u.User = url.UserPassword(u.User.Username(), "***")
	}
	return u.String()
}

const ucgRecommendConsumerEnabledEnv = "UCG_RECOMMEND_MQ_CONSUMER_ENABLED"

const ucgRecommendPrefetchEnv = "UCG_RECOMMEND_MQ_PREFETCH"

func ucgRecommendConsumerEnabled() bool {
	v := strings.TrimSpace(os.Getenv(ucgRecommendConsumerEnabledEnv))
	if v == "" {
		return true
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return true
	}
	return b
}

func ucgRecommendPrefetch() int {
	n := 5
	if v := strings.TrimSpace(os.Getenv(ucgRecommendPrefetchEnv)); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			n = p
		}
	}
	return n
}
