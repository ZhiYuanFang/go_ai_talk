package workeroutbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"hello/internal/platform/eventkit"
	"hello/internal/services/contracts"

	"github.com/gogf/gf/v2/os/glog"
)

type gfEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// EnqueueDomainOutbox 经 Worker HTTP 写入 domain_outbox（落库在 worker 专用库 ai_voice_worker 的 database.outbox 分组）。
func EnqueueDomainOutbox(ctx context.Context, routingKey eventkit.RouteKey, payload map[string]interface{}) error {
	if !routingKey.IsValid() {
		return fmt.Errorf("invalid routing key: %s", routingKey.String())
	}
	t := contracts.ResolveHTTPTargets()
	base := strings.TrimRight(strings.TrimSpace(t.WorkerBaseURL), "/")
	if base == "" {
		glog.Debugf(ctx, "[worker-outbox] WORKER_SERVICE_URL 未配置，跳过投递: key=%s", routingKey.String())
		return nil
	}
	u := base + t.WorkerOutboxEnqueuePath()
	body := map[string]interface{}{
		"routingKey": routingKey.String(),
		"payload":    payload,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("worker outbox enqueue: %w", err)
	}
	defer resp.Body.Close()
	var env gfEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return err
	}
	if resp.StatusCode >= 400 || env.Code != 0 {
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			return fmt.Errorf("worker outbox enqueue failed: status=%d", resp.StatusCode)
		}
		return fmt.Errorf("worker outbox enqueue failed: %s", msg)
	}
	return nil
}
