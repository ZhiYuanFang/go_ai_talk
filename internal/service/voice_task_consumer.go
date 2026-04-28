package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	mqConsumerEnabledEnv    = "MQ_CONSUMER_ENABLED"
	mqConsumerMaxRetryEnv   = "MQ_CONSUMER_MAX_RETRIES"
	mqConsumerWorkersEnv    = "MQ_CONSUMER_WORKERS"
	mqConsumerPollMsEnv     = "MQ_CONSUMER_POLL_INTERVAL_MS"
	mqConsumerQueueName     = "voice.task.requested.q"
	mqConsumerDoneKeyPrefix = "voice:mq:done:"
)

type voiceTaskConsumer struct {
	httpClient      *http.Client
	processFn       func(ctx context.Context, evt voiceTaskRequestedEvent) error
	isAlreadyDoneFn func(ctx context.Context, eventID string) (bool, error)
	markDoneFn      func(ctx context.Context, eventID string) error
	publishFn       func(ctx context.Context, routingKey string, payload map[string]interface{}) error
}

func newVoiceTaskConsumer() *voiceTaskConsumer {
	c := &voiceTaskConsumer{
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
	c.processFn = c.defaultProcess
	c.isAlreadyDoneFn = c.isAlreadyDone
	c.markDoneFn = c.markDone
	c.publishFn = c.publishSimple
	return c
}

func (c *voiceTaskConsumer) enabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(mqConsumerEnabledEnv)))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func (c *voiceTaskConsumer) consumeOnce(ctx context.Context) error {
	if !c.enabled() {
		return nil
	}
	evt, ok, err := c.pullOne(ctx)
	if err != nil || !ok {
		return err
	}
	return c.handleWithRetry(ctx, evt)
}

func (c *voiceTaskConsumer) handleWithRetry(ctx context.Context, evt voiceTaskRequestedEvent) error {
	done, err := c.isAlreadyDoneFn(ctx, evt.EventID)
	if err == nil && done {
		return nil
	}
	maxRetry := envIntOrDefault(mqConsumerMaxRetryEnv, 3)
	if maxRetry <= 0 {
		maxRetry = 1
	}
	var lastErr error
	for i := 1; i <= maxRetry; i++ {
		lastErr = c.processFn(ctx, evt)
		if lastErr == nil {
			_ = c.markDoneFn(ctx, evt.EventID)
			_ = c.publishFn(ctx, "voice.task.completed", map[string]interface{}{
				"event_id":    evt.EventID,
				"device_no":   evt.Payload.DeviceNo,
				"occurred_at": time.Now().Format(time.RFC3339Nano),
			})
			return nil
		}
		time.Sleep(time.Duration(i) * 100 * time.Millisecond)
	}
	_ = c.publishFn(ctx, "voice.task.failed", map[string]interface{}{
		"event_id":    evt.EventID,
		"device_no":   evt.Payload.DeviceNo,
		"reason":      lastErr.Error(),
		"occurred_at": time.Now().Format(time.RFC3339Nano),
	})
	return lastErr
}

func (c *voiceTaskConsumer) defaultProcess(ctx context.Context, evt voiceTaskRequestedEvent) error {
	if strings.TrimSpace(evt.Payload.DeviceNo) == "" || strings.TrimSpace(evt.Payload.Transcript) == "" {
		return fmt.Errorf("invalid payload")
	}
	return nil
}

func (c *voiceTaskConsumer) isAlreadyDone(ctx context.Context, eventID string) (bool, error) {
	if strings.TrimSpace(eventID) == "" {
		return false, nil
	}
	ret, err := g.Redis().Do(ctx, "EXISTS", mqConsumerDoneKeyPrefix+eventID)
	if err != nil {
		return false, err
	}
	return ret.Int() > 0, nil
}

func (c *voiceTaskConsumer) markDone(ctx context.Context, eventID string) error {
	if strings.TrimSpace(eventID) == "" {
		return nil
	}
	_, err := g.Redis().Do(ctx, "SET", mqConsumerDoneKeyPrefix+eventID, "1", "EX", 86400)
	return err
}

func (c *voiceTaskConsumer) pullOne(ctx context.Context) (voiceTaskRequestedEvent, bool, error) {
	base := strings.TrimSpace(os.Getenv(mqProducerAPIEnv))
	if base == "" {
		return voiceTaskRequestedEvent{}, false, fmt.Errorf("mq api base is empty")
	}
	u := strings.TrimRight(base, "/") + "/queues/%2F/" + mqConsumerQueueName + "/get"
	body := map[string]interface{}{
		"count":    1,
		"ackmode":  "ack_requeue_false",
		"encoding": "auto",
		"truncate": 50000,
	}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(data))
	if err != nil {
		return voiceTaskRequestedEvent{}, false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(defaultMQUser()+":"+defaultMQPass())))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return voiceTaskRequestedEvent{}, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return voiceTaskRequestedEvent{}, false, fmt.Errorf("mq get status=%d", resp.StatusCode)
	}
	var rows []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return voiceTaskRequestedEvent{}, false, err
	}
	if len(rows) == 0 {
		return voiceTaskRequestedEvent{}, false, nil
	}
	payload, _ := rows[0]["payload"].(string)
	if strings.TrimSpace(payload) == "" {
		return voiceTaskRequestedEvent{}, false, fmt.Errorf("empty payload")
	}
	var evt voiceTaskRequestedEvent
	if err := json.Unmarshal([]byte(payload), &evt); err != nil {
		return voiceTaskRequestedEvent{}, false, err
	}
	return evt, true, nil
}

func (c *voiceTaskConsumer) publishSimple(ctx context.Context, routingKey string, payload map[string]interface{}) error {
	base := strings.TrimSpace(os.Getenv(mqProducerAPIEnv))
	if base == "" {
		return nil
	}
	bodyObj := map[string]interface{}{
		"properties":       map[string]interface{}{},
		"routing_key":      routingKey,
		"payload":          mustJSON(payload),
		"payload_encoding": "string",
	}
	bodyBytes, _ := json.Marshal(bodyObj)
	u := strings.TrimRight(base, "/") + "/exchanges/%2F/" + mqProducerExchange + "/publish"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(defaultMQUser()+":"+defaultMQPass())))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("mq publish status=%d", resp.StatusCode)
	}
	return nil
}

func defaultMQUser() string {
	v := strings.TrimSpace(os.Getenv(mqProducerUserEnv))
	if v == "" {
		return "guest"
	}
	return v
}

func defaultMQPass() string {
	v := strings.TrimSpace(os.Getenv(mqProducerPassEnv))
	if v == "" {
		return "guest"
	}
	return v
}

func startVoiceTaskConsumer(ctx context.Context) {
	workers := envIntOrDefault(mqConsumerWorkersEnv, 1)
	if workers <= 0 {
		workers = 1
	}
	pollIntervalMs := envIntOrDefault(mqConsumerPollMsEnv, 2000)
	if pollIntervalMs < 100 {
		pollIntervalMs = 100
	}

	for i := 0; i < workers; i++ {
		consumer := newVoiceTaskConsumer()
		if !consumer.enabled() {
			return
		}
		workerID := i + 1
		go runVoiceTaskWorker(ctx, workerID, consumer, time.Duration(pollIntervalMs)*time.Millisecond)
	}
}

func runVoiceTaskWorker(ctx context.Context, workerID int, consumer *voiceTaskConsumer, pollInterval time.Duration) {
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
