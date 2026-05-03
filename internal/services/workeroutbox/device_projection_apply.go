package workeroutbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"hello/internal/services/contracts"
)

// ApplyDeviceProjectionHTTP 在 device-service 进程内执行缓存投影，避免 worker 直连 device 库。
func ApplyDeviceProjectionHTTP(ctx context.Context, routingKey, payload string) error {
	t := contracts.ResolveHTTPTargets()
	base := strings.TrimRight(strings.TrimSpace(t.DeviceBaseURL), "/")
	if base == "" {
		return fmt.Errorf("ApplyDeviceProjectionHTTP: DEVICE_SERVICE_URL empty")
	}
	u := base + t.DeviceInternalProjectionApplyPath()
	raw, err := json.Marshal(map[string]string{
		"routingKey": strings.TrimSpace(routingKey),
		"payload":    payload,
	})
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
		return fmt.Errorf("device projection apply: %w", err)
	}
	defer resp.Body.Close()
	var env gfEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return err
	}
	if resp.StatusCode >= 400 || env.Code != 0 {
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			return fmt.Errorf("device projection apply failed: status=%d", resp.StatusCode)
		}
		return fmt.Errorf("device projection apply failed: %s", msg)
	}
	return nil
}
