package mq

import (
	"os"
	"strings"

	"hello/internal/platform/eventkit"
)

const (
	ProducerAPIEnv   = "MQ_HTTP_API_BASE"
	ProducerUserEnv  = "MQ_USER"
	ProducerPassEnv  = "MQ_PASSWORD"
	ProducerExchange = "voice.events"
)

// NewObservedEventPublisher 统一构造带观测能力的事件发布器。
func NewObservedEventPublisher() (eventkit.Publisher, error) {
	base, err := eventkit.NewHTTPPublisher(eventkit.HTTPPublisherConfig{
		APIBase:  strings.TrimSpace(os.Getenv(ProducerAPIEnv)),
		User:     strings.TrimSpace(os.Getenv(ProducerUserEnv)),
		Password: strings.TrimSpace(os.Getenv(ProducerPassEnv)),
		Exchange: ProducerExchange,
	})
	if err != nil {
		return nil, err
	}
	return eventkit.WithObserver(base, eventkit.LoggingObserver{}), nil
}

// NewHTTPPublisherForDelayed 构造可 PublishDelayed 的 HTTP 发布器（目标交换机仍由 PublishDelayed 固定为 voice.delayed）。
func NewHTTPPublisherForDelayed() (*eventkit.HTTPPublisher, error) {
	return eventkit.NewHTTPPublisher(eventkit.HTTPPublisherConfig{
		APIBase:  strings.TrimSpace(os.Getenv(ProducerAPIEnv)),
		User:     strings.TrimSpace(os.Getenv(ProducerUserEnv)),
		Password: strings.TrimSpace(os.Getenv(ProducerPassEnv)),
		Exchange: ProducerExchange,
	})
}

// NewHTTPQueueConsumer 统一构造队列拉取消费器。
func NewHTTPQueueConsumer() (*eventkit.HTTPQueueConsumer, error) {
	return eventkit.NewHTTPQueueConsumer(eventkit.HTTPQueueConsumerConfig{
		APIBase:  strings.TrimSpace(os.Getenv(ProducerAPIEnv)),
		User:     strings.TrimSpace(os.Getenv(ProducerUserEnv)),
		Password: strings.TrimSpace(os.Getenv(ProducerPassEnv)),
	})
}
