package eventkit

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type HTTPQueueConsumerConfig struct {
	APIBase  string
	User     string
	Password string
	Timeout  time.Duration
}

type HTTPQueueConsumer struct {
	cfg        HTTPQueueConsumerConfig
	httpClient *http.Client
}

func NewHTTPQueueConsumer(cfg HTTPQueueConsumerConfig) (*HTTPQueueConsumer, error) {
	cfg.APIBase = strings.TrimSpace(cfg.APIBase)
	cfg.User = strings.TrimSpace(cfg.User)
	cfg.Password = strings.TrimSpace(cfg.Password)
	if cfg.APIBase == "" {
		return nil, ErrEmptyAPIBase
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.User == "" {
		cfg.User = "guest"
	}
	if cfg.Password == "" {
		cfg.Password = "guest"
	}
	return &HTTPQueueConsumer{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}, nil
}

func (c *HTTPQueueConsumer) PullOne(ctx context.Context, queueName string) (string, bool, error) {
	queueName = strings.TrimSpace(queueName)
	if queueName == "" {
		return "", false, fmt.Errorf("%w: empty queue name", ErrUnavailable)
	}
	u := strings.TrimRight(c.cfg.APIBase, "/") + "/queues/%2F/" + queueName + "/get"
	body := map[string]any{
		"count":    1,
		// 拉取后立即 ack，消费端失败由上层重试与失败事件处理。
		"ackmode":  "ack_requeue_false",
		"encoding": "auto",
		"truncate": 50000,
	}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(data))
	if err != nil {
		return "", false, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.basicAuth())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", false, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", false, fmt.Errorf("%w: status=%d", ErrUnavailable, resp.StatusCode)
	}
	var rows []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return "", false, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if len(rows) == 0 {
		// 空队列不是错误，交由轮询层继续下一次拉取。
		return "", false, nil
	}
	payload, _ := rows[0]["payload"].(string)
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return "", false, fmt.Errorf("%w: empty payload", ErrUnavailable)
	}
	return payload, true, nil
}

func (c *HTTPQueueConsumer) basicAuth() string {
	raw := c.cfg.User + ":" + c.cfg.Password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(raw))
}

