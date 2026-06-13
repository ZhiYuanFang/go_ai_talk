package eventkit

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// AMQPMessageHandler 处理单条 delivery；返回 nil 表示 Ack，非 nil 表示 Nack(requeue)。
// routingKey 来自 delivery.RoutingKey，queue 订阅场景下可用于区分事件类型。
type AMQPMessageHandler func(ctx context.Context, queueName, routingKey string, body []byte) error

// AMQPQueueSubscription 单队列 consumer 注册项（共用 connection 时各队列独立 channel）。
type AMQPQueueSubscription struct {
	QueueName         string
	Prefetch          int
	ConsumerTagPrefix string
	Handler           AMQPMessageHandler
}

// SharedAMQPConfig 单 connection 多 channel push consumer。
type SharedAMQPConfig struct {
	URL          string
	Subscriptions []AMQPQueueSubscription
	ReconnectMin time.Duration
	ReconnectMax time.Duration
}

// AMQPConsumerConfig 兼容旧用法：多队列共用同一 handler。
type AMQPConsumerConfig struct {
	URL               string
	Queues            []string
	Prefetch          int
	ConsumerTagPrefix string
	ReconnectMin      time.Duration
	ReconnectMax      time.Duration
}

// RunSharedAMQPConsumers 在 ctx 存活期间维持单 AMQP 连接，为每订阅项启动独立 channel + Consume goroutine；
// 连接或任一 channel 异常关闭后 exponential backoff 重连。
func RunSharedAMQPConsumers(ctx context.Context, cfg SharedAMQPConfig) {
	if len(cfg.Subscriptions) == 0 {
		return
	}
	minWait := cfg.ReconnectMin
	if minWait <= 0 {
		minWait = time.Second
	}
	maxWait := cfg.ReconnectMax
	if maxWait <= 0 {
		maxWait = 30 * time.Second
	}
	tagPrefix := "eventkit"

	go func() {
		backoff := minWait
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			sessionErr := runSharedAMQPSession(ctx, cfg.URL, tagPrefix, cfg.Subscriptions)
			select {
			case <-ctx.Done():
				return
			default:
			}
			if sessionErr == nil {
				backoff = minWait
				continue
			}
			time.Sleep(backoff)
			if backoff < maxWait {
				backoff *= 2
				if backoff > maxWait {
					backoff = maxWait
				}
			}
		}
	}()
}

// RunAMQPConsumers 兼容包装：所有队列共用同一 handler。
func RunAMQPConsumers(ctx context.Context, cfg AMQPConsumerConfig, handler AMQPMessageHandler) {
	if handler == nil || len(cfg.Queues) == 0 {
		return
	}
	prefetch := cfg.Prefetch
	if prefetch <= 0 {
		prefetch = 1
	}
	tagPrefix := cfg.ConsumerTagPrefix
	if tagPrefix == "" {
		tagPrefix = "eventkit"
	}
	subs := make([]AMQPQueueSubscription, len(cfg.Queues))
	for i, q := range cfg.Queues {
		subs[i] = AMQPQueueSubscription{
			QueueName:         q,
			Prefetch:          prefetch,
			ConsumerTagPrefix: tagPrefix,
			Handler:           handler,
		}
	}
	RunSharedAMQPConsumers(ctx, SharedAMQPConfig{
		URL:           cfg.URL,
		Subscriptions: subs,
		ReconnectMin:  cfg.ReconnectMin,
		ReconnectMax:  cfg.ReconnectMax,
	})
}

func runSharedAMQPSession(ctx context.Context, url, defaultTagPrefix string, subs []AMQPQueueSubscription) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("amqp dial: %w", err)
	}
	defer conn.Close()

	connClosed := conn.NotifyClose(make(chan *amqp.Error, 1))

	type queueWorker struct {
		ch     *amqp.Channel
		closed chan *amqp.Error
	}
	workers := make([]queueWorker, 0, len(subs))
	var wg sync.WaitGroup

	for i, sub := range subs {
		if sub.Handler == nil || sub.QueueName == "" {
			continue
		}
		prefetch := sub.Prefetch
		if prefetch <= 0 {
			prefetch = 1
		}
		tagPrefix := sub.ConsumerTagPrefix
		if tagPrefix == "" {
			tagPrefix = defaultTagPrefix
		}
		ch, chErr := conn.Channel()
		if chErr != nil {
			return fmt.Errorf("amqp channel queue=%s: %w", sub.QueueName, chErr)
		}
		if err = ch.Qos(prefetch, 0, false); err != nil {
			_ = ch.Close()
			return fmt.Errorf("amqp qos queue=%s: %w", sub.QueueName, err)
		}
		tag := fmt.Sprintf("%s-%s-%d", tagPrefix, sub.QueueName, i)
		deliveries, err := ch.Consume(sub.QueueName, tag, false, false, false, false, nil)
		if err != nil {
			_ = ch.Close()
			return fmt.Errorf("amqp consume queue=%s: %w", sub.QueueName, err)
		}
		closed := ch.NotifyClose(make(chan *amqp.Error, 1))
		workers = append(workers, queueWorker{ch: ch, closed: closed})

		handler := sub.Handler
		queueName := sub.QueueName
		wg.Add(1)
		go func(q string, msgs <-chan amqp.Delivery, h AMQPMessageHandler) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case d, ok := <-msgs:
					if !ok {
						return
					}
					handleDelivery(ctx, q, d, h)
				}
			}
		}(queueName, deliveries, handler)
	}

	errCh := make(chan error, len(workers)+1)
	go func() {
		if err := <-connClosed; err != nil {
			errCh <- fmt.Errorf("amqp connection closed: %w", err)
			return
		}
		errCh <- fmt.Errorf("amqp connection closed")
	}()
	for _, w := range workers {
		w := w
		go func() {
			if err := <-w.closed; err != nil {
				errCh <- fmt.Errorf("amqp channel closed: %w", err)
				return
			}
			errCh <- fmt.Errorf("amqp channel closed")
		}()
	}

	select {
	case <-ctx.Done():
		for _, w := range workers {
			_ = w.ch.Close()
		}
		wg.Wait()
		return ctx.Err()
	case err := <-errCh:
		_ = conn.Close()
		for _, w := range workers {
			_ = w.ch.Close()
		}
		wg.Wait()
		return err
	}
}

// handleDelivery 统一 Ack/Nack 语义：handler 返回 err 则 Nack(requeue=true) 消息回队。
// UCG 审核 Green 风暴：Phase1 Green 失败或 persist verdict 失败时 handler 返回 err → 无限 requeue → 重复调 Green。
func handleDelivery(ctx context.Context, queueName string, d amqp.Delivery, handler AMQPMessageHandler) {
	if err := handler(ctx, queueName, d.RoutingKey, d.Body); err != nil {
		_ = d.Nack(false, true) // requeue=true：消息立即回到队列 ready，consumer 空闲时会再次投递
		return
	}
	_ = d.Ack(false) // 成功处理，从队列删除
}
