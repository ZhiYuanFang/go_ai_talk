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
	"sync"
	"time"

	"hello/internal/dao"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	outboxEnabledEnv      = "OUTBOX_RELAY_ENABLED"
	outboxWorkersEnv      = "OUTBOX_RELAY_WORKERS"
	outboxPollMsEnv       = "OUTBOX_RELAY_POLL_INTERVAL_MS"
	outboxMaxAttemptsEnv  = "OUTBOX_RELAY_MAX_ATTEMPTS"
	outboxTableName       = "domain_outbox"
	outboxStatusPending   = "pending"
	outboxStatusPublished = "published"
	outboxStatusFailed    = "failed"
)

var (
	outboxRelayOnce sync.Once
)

type domainOutboxRow struct {
	Id         int64  `json:"id"`
	RoutingKey string `json:"routing_key"`
	Payload    string `json:"payload"`
	Attempts   int    `json:"attempts"`
}

func isOutboxRelayEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(outboxEnabledEnv)))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func startDomainOutboxRelay(ctx context.Context) {
	if !isOutboxRelayEnabled() {
		return
	}
	outboxRelayOnce.Do(func() {
		workers := envIntOrDefault(outboxWorkersEnv, 1)
		if workers <= 0 {
			workers = 1
		}
		pollMs := envIntOrDefault(outboxPollMsEnv, 1500)
		if pollMs < 100 {
			pollMs = 100
		}
		for i := 0; i < workers; i++ {
			workerID := i + 1
			go runOutboxRelayWorker(ctx, workerID, time.Duration(pollMs)*time.Millisecond)
		}
	})
}

func runOutboxRelayWorker(ctx context.Context, workerID int, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := relayOneOutboxEvent(ctx); err != nil {
				glog.Warningf(ctx, "outbox relay worker %d failed: %v", workerID, err)
			}
		}
	}
}

func relayOneOutboxEvent(ctx context.Context) error {
	group := dao.History.Group()
	maxAttempts := envIntOrDefault(outboxMaxAttemptsEnv, 10)
	if maxAttempts <= 0 {
		maxAttempts = 10
	}
	return g.DB(group).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		row, err := tx.Model(outboxTableName).
			WhereIn("status", []string{outboxStatusPending, outboxStatusFailed}).
			Where("attempts<?", maxAttempts).
			OrderAsc("id").
			Limit(1).
			One()
		if err != nil {
			return err
		}
		if row == nil || row.IsEmpty() {
			return nil
		}
		item := domainOutboxRow{
			Id:         row["id"].Int64(),
			RoutingKey: row["routing_key"].String(),
			Payload:    row["payload"].String(),
			Attempts:   row["attempts"].Int(),
		}
		if strings.TrimSpace(item.RoutingKey) == "" || strings.TrimSpace(item.Payload) == "" {
			_, _ = tx.Model(outboxTableName).Where("id", item.Id).Data(g.Map{
				"status":     outboxStatusFailed,
				"attempts":   item.Attempts + 1,
				"last_error": "invalid outbox payload",
			}).Update()
			return nil
		}
		if err := publishMQEvent(ctx, item.RoutingKey, item.Payload); err != nil {
			nextStatus := outboxStatusFailed
			_, _ = tx.Model(outboxTableName).Where("id", item.Id).Data(g.Map{
				"status":      nextStatus,
				"attempts":    item.Attempts + 1,
				"last_error":  truncateError(err.Error(), 512),
				"published_at": nil,
			}).Update()
			return nil
		}
		_, _ = tx.Model(outboxTableName).Where("id", item.Id).Data(g.Map{
			"status":       outboxStatusPublished,
			"attempts":     item.Attempts + 1,
			"last_error":   "",
			"published_at": gdb.Raw("NOW()"),
		}).Update()
		return nil
	})
}

func insertOutboxEventTx(tx gdb.TX, routingKey string, payload map[string]interface{}) error {
	eventID := fmt.Sprintf("outbox-%d", time.Now().UnixNano())
	body := mustJSON(payload)
	_, err := tx.Model(outboxTableName).Data(g.Map{
		"event_id":    eventID,
		"event_type":  routingKey,
		"routing_key": routingKey,
		"payload":     body,
		"status":      outboxStatusPending,
		"attempts":    0,
		"last_error":  "",
	}).Insert()
	return err
}

func publishMQEvent(ctx context.Context, routingKey, payload string) error {
	base := strings.TrimSpace(os.Getenv(mqProducerAPIEnv))
	if base == "" {
		return fmt.Errorf("mq api base is empty")
	}
	bodyObj := map[string]interface{}{
		"properties":       map[string]interface{}{},
		"routing_key":      routingKey,
		"payload":          payload,
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
	resp, err := (&http.Client{Timeout: 3 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("mq publish status=%d", resp.StatusCode)
	}
	return nil
}

func truncateError(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max]
}
