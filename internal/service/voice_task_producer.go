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
)

const (
	mqProducerEnabledEnv = "MQ_PRODUCER_ENABLED"
	mqProducerAPIEnv     = "MQ_HTTP_API_BASE" // e.g. http://127.0.0.1:15672/api
	mqProducerUserEnv    = "MQ_USER"
	mqProducerPassEnv    = "MQ_PASSWORD"
	mqProducerExchange   = "voice.events"
	mqProducerRoutingKey = "voice.task.requested"
)

type voiceTaskProducer struct {
	httpClient *http.Client
}

type voiceTaskRequestedEvent struct {
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

func newVoiceTaskProducer() *voiceTaskProducer {
	return &voiceTaskProducer{
		httpClient: &http.Client{Timeout: 3 * time.Second},
	}
}

func (p *voiceTaskProducer) enabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(mqProducerEnabledEnv)))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func (p *voiceTaskProducer) publishTaskRequested(ctx context.Context, deviceNo, transcript, source string) error {
	if !p.enabled() {
		return nil
	}
	base := strings.TrimSpace(os.Getenv(mqProducerAPIEnv))
	if base == "" {
		return fmt.Errorf("mq producer enabled but %s is empty", mqProducerAPIEnv)
	}

	event := voiceTaskRequestedEvent{
		EventID:      fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		EventType:    mqProducerRoutingKey,
		EventVersion: "v1",
		Producer:     "voice-service",
		OccurredAt:   time.Now().Format(time.RFC3339Nano),
		TraceID:      fmt.Sprintf("trace-%d", time.Now().UnixNano()),
	}
	event.Payload.DeviceNo = strings.TrimSpace(deviceNo)
	event.Payload.Transcript = strings.TrimSpace(transcript)
	event.Payload.Source = strings.TrimSpace(source)

	bodyObj := map[string]interface{}{
		"properties":       map[string]interface{}{},
		"routing_key":      mqProducerRoutingKey,
		"payload":          mustJSON(event),
		"payload_encoding": "string",
	}
	bodyBytes, _ := json.Marshal(bodyObj)

	u := strings.TrimRight(base, "/") + "/exchanges/%2F/" + mqProducerExchange + "/publish"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	user := strings.TrimSpace(os.Getenv(mqProducerUserEnv))
	pass := strings.TrimSpace(os.Getenv(mqProducerPassEnv))
	if user == "" {
		user = "guest"
	}
	if pass == "" {
		pass = "guest"
	}
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(user+":"+pass)))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("mq publish http status=%d", resp.StatusCode)
	}
	return nil
}

func mustJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
