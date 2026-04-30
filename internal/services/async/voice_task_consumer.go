package async

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"hello/internal/platform/cachekit"
	"hello/internal/platform/eventkit"
	"hello/internal/shared/mq"

	"github.com/gogf/gf/v2/os/glog"
)

const (
	mqConsumerMaxRetryEnv   = "MQ_CONSUMER_MAX_RETRIES"
	mqConsumerWorkersEnv    = "MQ_CONSUMER_WORKERS"
	mqConsumerPollMsEnv     = "MQ_CONSUMER_POLL_INTERVAL_MS"
	mqConsumerQueueName     = "voice.task.requested.q"
	mqConsumerDoneKeyPrefix = "voice:mq:done:"
)

type queueConsumer interface {
	PullOne(ctx context.Context, queueName string) (string, bool, error)
}

type VoiceTaskConsumer struct {
	queueConsumerFactory func() (queueConsumer, error)
	publisherFactory     func() (Publisher, error)
	cache                cachekit.Cache
}

func NewVoiceTaskConsumer() *VoiceTaskConsumer {
	return &VoiceTaskConsumer{
		queueConsumerFactory: func() (queueConsumer, error) { return newHTTPQueueConsumer() },
		publisherFactory:     func() (Publisher, error) { return newObservedEventPublisher() },
		cache:                cachekit.WithObserver(cachekit.NewRedisCache(), cachekit.LoggingObserver{}),
	}
}

func (c *VoiceTaskConsumer) consumeOnce(ctx context.Context) error {
	evt, ok, err := c.pullOne(ctx)
	if err != nil || !ok {
		return err
	}
	return c.handleWithRetry(ctx, evt)
}

func (c *VoiceTaskConsumer) handleWithRetry(ctx context.Context, evt VoiceTaskRequestedEvent) error {
	done, err := c.isAlreadyDone(ctx, evt.EventID)
	if err == nil && done {
		return nil
	}
	maxRetry := envIntOrDefault(mqConsumerMaxRetryEnv, 3)
	if maxRetry <= 0 {
		maxRetry = 1
	}
	var lastErr error
	for i := 1; i <= maxRetry; i++ {
		lastErr = c.defaultProcess(ctx, evt)
		if lastErr == nil {
			_ = c.markDone(ctx, evt.EventID)
			_ = c.publishEvent(ctx, "voice.task.completed", map[string]any{
				"event_id":    evt.EventID,
				"device_no":   evt.Payload.DeviceNo,
				"occurred_at": time.Now().Format(time.RFC3339Nano),
			})
			return nil
		}
		time.Sleep(time.Duration(i) * 100 * time.Millisecond)
	}
	_ = c.publishEvent(ctx, "voice.task.failed", map[string]any{
		"event_id":    evt.EventID,
		"device_no":   evt.Payload.DeviceNo,
		"reason":      lastErr.Error(),
		"occurred_at": time.Now().Format(time.RFC3339Nano),
	})
	return lastErr
}

func (c *VoiceTaskConsumer) defaultProcess(ctx context.Context, evt VoiceTaskRequestedEvent) error {
	if strings.TrimSpace(evt.Payload.DeviceNo) == "" || strings.TrimSpace(evt.Payload.Transcript) == "" {
		return fmt.Errorf("invalid payload")
	}
	return nil
}

func (c *VoiceTaskConsumer) isAlreadyDone(ctx context.Context, eventID string) (bool, error) {
	if strings.TrimSpace(eventID) == "" {
		return false, nil
	}
	ret, err := c.cache.Exists(ctx, mqConsumerDoneKeyPrefix+eventID)
	if err != nil {
		return false, err
	}
	return ret, nil
}

func (c *VoiceTaskConsumer) markDone(ctx context.Context, eventID string) error {
	if strings.TrimSpace(eventID) == "" {
		return nil
	}
	return c.cache.SetEX(ctx, mqConsumerDoneKeyPrefix+eventID, "1", 24*time.Hour)
}

func (c *VoiceTaskConsumer) pullOne(ctx context.Context) (VoiceTaskRequestedEvent, bool, error) {
	consumer, err := c.queueConsumerFactory()
	if err != nil {
		return VoiceTaskRequestedEvent{}, false, err
	}
	payload, ok, err := consumer.PullOne(ctx, mqConsumerQueueName)
	if err != nil || !ok {
		return VoiceTaskRequestedEvent{}, ok, err
	}
	var evt VoiceTaskRequestedEvent
	if err = json.Unmarshal([]byte(payload), &evt); err != nil {
		return VoiceTaskRequestedEvent{}, false, err
	}
	return evt, true, nil
}

func (c *VoiceTaskConsumer) publishEvent(ctx context.Context, routingKey string, payload map[string]any) error {
	pub, err := c.publisherFactory()
	if err != nil {
		return err
	}
	return pub.Publish(ctx, routingKey, payload)
}

func StartVoiceTaskConsumer(ctx context.Context) {
	workers := envIntOrDefault(mqConsumerWorkersEnv, 1)
	if workers <= 0 {
		workers = 1
	}
	pollIntervalMs := envIntOrDefault(mqConsumerPollMsEnv, 2000)
	if pollIntervalMs < 100 {
		pollIntervalMs = 100
	}
	for i := 0; i < workers; i++ {
		consumer := NewVoiceTaskConsumer()
		workerID := i + 1
		go runVoiceTaskWorker(ctx, workerID, consumer, time.Duration(pollIntervalMs)*time.Millisecond)
	}
}

func runVoiceTaskWorker(ctx context.Context, workerID int, consumer *VoiceTaskConsumer, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := consumer.consumeOnce(ctx); err != nil {
				glog.Warningf(ctx, "voice task worker %d consume failed: %v", workerID, err)
			}
		}
	}
}

func newHTTPQueueConsumer() (*eventkit.HTTPQueueConsumer, error) {
	return mq.NewHTTPQueueConsumer()
}

func envIntOrDefault(key string, d int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if n, err := strconv.Atoi(v); err == nil {
		return n
	}
	return d
}
