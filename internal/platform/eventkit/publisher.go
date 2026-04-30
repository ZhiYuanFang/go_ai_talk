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

type Publisher interface {
	// CheckDependency 用于启动前探活 RabbitMQ 管理 API。
	CheckDependency(ctx context.Context) error
	// Publish 为强语义发布：调用方应将错误视为可阻断失败。
	Publish(ctx context.Context, routingKey string, payload any) error
}

type HTTPPublisherConfig struct {
	APIBase  string
	User     string
	Password string
	Exchange string
	Timeout  time.Duration
}

type HTTPPublisher struct {
	cfg        HTTPPublisherConfig
	httpClient *http.Client
}

func NewHTTPPublisher(cfg HTTPPublisherConfig) (*HTTPPublisher, error) {
	cfg.APIBase = strings.TrimSpace(cfg.APIBase)
	cfg.User = strings.TrimSpace(cfg.User)
	cfg.Password = strings.TrimSpace(cfg.Password)
	cfg.Exchange = strings.TrimSpace(cfg.Exchange)
	if cfg.APIBase == "" {
		return nil, ErrEmptyAPIBase
	}
	if cfg.Exchange == "" {
		cfg.Exchange = DefaultExchange
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 3 * time.Second
	}
	if cfg.User == "" {
		cfg.User = "guest"
	}
	if cfg.Password == "" {
		cfg.Password = "guest"
	}
	return &HTTPPublisher{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}, nil
}

func (p *HTTPPublisher) CheckDependency(ctx context.Context) error {
	// 通过 /overview 做轻量健康探测，避免启动后首次发布才暴露依赖故障。
	u := strings.TrimRight(p.cfg.APIBase, "/") + "/overview"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDependencyFailed, err)
	}
	req.Header.Set("Authorization", p.basicAuth())
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDependencyFailed, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("%w: status=%d", ErrDependencyFailed, resp.StatusCode)
	}
	return nil
}

func (p *HTTPPublisher) Publish(ctx context.Context, routingKey string, payload any) error {
	routingKey = strings.TrimSpace(routingKey)
	if routingKey == "" {
		return ErrEmptyRoutingKey
	}
	if _, ok := ParseRoutingKey(routingKey); !ok {
		return ErrInvalidRoutingKey
	}
	bodyObj := map[string]any{
		"properties":       map[string]any{},
		"routing_key":      routingKey,
		"payload":          mustJSON(payload),
		"payload_encoding": "string",
	}
	bodyBytes, _ := json.Marshal(bodyObj)
	u := strings.TrimRight(p.cfg.APIBase, "/") + "/exchanges/%2F/" + p.cfg.Exchange + "/publish"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", p.basicAuth())
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		// 管理 API 返回非 2xx 视为发布被拒绝，需要上层走失败语义。
		return fmt.Errorf("%w: status=%d", ErrPublishRejected, resp.StatusCode)
	}
	return nil
}

func (p *HTTPPublisher) basicAuth() string {
	// 管理 API 使用 Basic Auth，不引入额外签名流程。
	raw := p.cfg.User + ":" + p.cfg.Password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(raw))
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

