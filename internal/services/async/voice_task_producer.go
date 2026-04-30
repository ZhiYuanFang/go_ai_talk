package async

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hello/internal/platform/eventkit"
	"hello/internal/shared/mq"
)

const (
	MQProducerRoutingKey = "voice.task.requested"
)

type Publisher interface {
	Publish(ctx context.Context, routingKey string, payload any) error
}

type VoiceTaskRequestedEvent struct {
	EventID      string `json:"event_id"`
	EventType    string `json:"event_type"`
	EventVersion string `json:"event_version"`
	Producer     string `json:"producer"`
	OccurredAt   string `json:"occurred_at"`
	TraceID      string `json:"trace_id"`
	Payload      struct {
		DeviceNo   string `json:"device_no"`
		Transcript string `json:"transcript"`
		Source     string `json:"source"`
	} `json:"payload"`
}

type VoiceTaskProducer struct {
	publisherFactory func() (Publisher, error)
}

func NewVoiceTaskProducer() *VoiceTaskProducer {
	return &VoiceTaskProducer{
		publisherFactory: func() (Publisher, error) {
			return newObservedEventPublisher()
		},
	}
}

func (p *VoiceTaskProducer) PublishTaskRequested(ctx context.Context, deviceNo, transcript, source string) error {
	pub, err := p.publisherFactory()
	if err != nil {
		return err
	}
	event := VoiceTaskRequestedEvent{
		EventID:      fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		EventType:    MQProducerRoutingKey,
		EventVersion: "v1",
		Producer:     "voice-service",
		OccurredAt:   time.Now().Format(time.RFC3339Nano),
		TraceID:      fmt.Sprintf("trace-%d", time.Now().UnixNano()),
	}
	event.Payload.DeviceNo = strings.TrimSpace(deviceNo)
	event.Payload.Transcript = strings.TrimSpace(transcript)
	event.Payload.Source = strings.TrimSpace(source)
	return pub.Publish(ctx, MQProducerRoutingKey, event)
}

func newObservedEventPublisher() (eventkit.Publisher, error) {
	return mq.NewObservedEventPublisher()
}
