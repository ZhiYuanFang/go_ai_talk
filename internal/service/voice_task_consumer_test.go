package service

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

func TestVoiceTaskConsumerHandleWithRetrySuccess(t *testing.T) {
	c := newVoiceTaskConsumer()
	attempt := 0
	c.processFn = func(ctx context.Context, evt voiceTaskRequestedEvent) error {
		attempt++
		if attempt < 2 {
			return errors.New("temporary error")
		}
		return nil
	}
	c.isAlreadyDoneFn = func(ctx context.Context, eventID string) (bool, error) { return false, nil }
	markCalled := false
	c.markDoneFn = func(ctx context.Context, eventID string) error {
		markCalled = true
		return nil
	}
	c.publishFn = func(ctx context.Context, routingKey string, payload map[string]interface{}) error {
		return nil
	}
	evt := voiceTaskRequestedEvent{EventID: "evt-1"}
	if err := c.handleWithRetry(context.Background(), evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempt != 2 {
		t.Fatalf("unexpected attempt count: %d", attempt)
	}
	if !markCalled {
		t.Fatal("expected markDone called")
	}
}

func TestVoiceTaskConsumerHandleWithRetryFailedToDLQ(t *testing.T) {
	c := newVoiceTaskConsumer()
	c.processFn = func(ctx context.Context, evt voiceTaskRequestedEvent) error { return errors.New("always fail") }
	c.isAlreadyDoneFn = func(ctx context.Context, eventID string) (bool, error) { return false, nil }
	c.markDoneFn = func(ctx context.Context, eventID string) error { return nil }
	dlqCalled := false
	c.publishFn = func(ctx context.Context, routingKey string, payload map[string]interface{}) error {
		if routingKey == "voice.task.failed" {
			dlqCalled = true
		}
		return nil
	}
	t.Setenv(mqConsumerMaxRetryEnv, "2")
	evt := voiceTaskRequestedEvent{EventID: "evt-2"}
	err := c.handleWithRetry(context.Background(), evt)
	if err == nil {
		t.Fatal("expected error")
	}
	if !dlqCalled {
		t.Fatal("expected failed event publish")
	}
}

func TestStartVoiceTaskConsumerWorkerPool(t *testing.T) {
	oldEnabled := os.Getenv(mqConsumerEnabledEnv)
	oldWorkers := os.Getenv(mqConsumerWorkersEnv)
	oldPoll := os.Getenv(mqConsumerPollMsEnv)
	defer func() {
		_ = os.Setenv(mqConsumerEnabledEnv, oldEnabled)
		_ = os.Setenv(mqConsumerWorkersEnv, oldWorkers)
		_ = os.Setenv(mqConsumerPollMsEnv, oldPoll)
	}()

	_ = os.Setenv(mqConsumerEnabledEnv, "true")
	_ = os.Setenv(mqConsumerWorkersEnv, "2")
	_ = os.Setenv(mqConsumerPollMsEnv, "100")

	ctx, cancel := context.WithCancel(context.Background())
	startVoiceTaskConsumer(ctx)
	time.Sleep(250 * time.Millisecond)
	cancel()
}
