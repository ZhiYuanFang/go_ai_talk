package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestVoiceTaskProducerPublishTaskRequested(t *testing.T) {
	var gotPath string
	var gotRouting string
	var gotPayload string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if v, ok := body["routing_key"].(string); ok {
			gotRouting = v
		}
		if v, ok := body["payload"].(string); ok {
			gotPayload = v
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"routed":true}`))
	}))
	defer srv.Close()

	oldEnabled := os.Getenv(mqProducerEnabledEnv)
	oldAPI := os.Getenv(mqProducerAPIEnv)
	oldUser := os.Getenv(mqProducerUserEnv)
	oldPass := os.Getenv(mqProducerPassEnv)
	defer func() {
		_ = os.Setenv(mqProducerEnabledEnv, oldEnabled)
		_ = os.Setenv(mqProducerAPIEnv, oldAPI)
		_ = os.Setenv(mqProducerUserEnv, oldUser)
		_ = os.Setenv(mqProducerPassEnv, oldPass)
	}()

	_ = os.Setenv(mqProducerEnabledEnv, "true")
	_ = os.Setenv(mqProducerAPIEnv, srv.URL+"/api")
	_ = os.Setenv(mqProducerUserEnv, "guest")
	_ = os.Setenv(mqProducerPassEnv, "guest")

	p := newVoiceTaskProducer()
	if err := p.publishTaskRequested(context.Background(), "device-1", "你好", "text-chat"); err != nil {
		t.Fatalf("unexpected publish error: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/exchanges///voice.events/publish") &&
		!strings.HasSuffix(gotPath, "/exchanges/%2F/voice.events/publish") {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotRouting != mqProducerRoutingKey {
		t.Fatalf("unexpected routing key: %s", gotRouting)
	}
	if gotPayload == "" {
		t.Fatal("expected non-empty payload")
	}
}
