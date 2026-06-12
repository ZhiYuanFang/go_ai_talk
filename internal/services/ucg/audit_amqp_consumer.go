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
	rabbitAMQPURLEnv    = "RABBITMQ_AMQP_URL"
	rabbitHostEnv       = "RABBITMQ_HOST"
	rabbitAMQPPortEnv   = "RABBITMQ_AMQP_PORT"
	mqUserEnv           = "MQ_USER"
	mqPasswordEnv       = "MQ_PASSWORD"
	ucgAuditPrefetchEnv = "UCG_AUDIT_MQ_PREFETCH"
)

func ucgAuditAMQPHandler(ctx context.Context, queueName, routingKey string, body []byte) error {
	_ = routingKey
	return dispatchUcgAuditPayload(ctx, queueName, string(body))
}

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

// ucgRecommendConsumerEnabledEnv 推荐分 MQ consumer 开关，默认开启。
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
