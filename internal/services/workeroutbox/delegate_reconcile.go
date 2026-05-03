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

	"github.com/gogf/gf/v2/os/glog"
)

// RunProjectionReconcileDelegating worker 无 history/device 库时，经 HTTP 触发各域投影修复。
func RunProjectionReconcileDelegating(ctx context.Context, limit int) error {
	t := contracts.ResolveHTTPTargets()
	client := &http.Client{Timeout: 30 * time.Second}
	hBase := strings.TrimRight(strings.TrimSpace(t.HistoryBaseURL), "/")
	dBase := strings.TrimRight(strings.TrimSpace(t.DeviceBaseURL), "/")
	if hBase == "" {
		return fmt.Errorf("projection reconcile: HISTORY_SERVICE_URL empty")
	}
	if dBase == "" {
		return fmt.Errorf("projection reconcile: DEVICE_SERVICE_URL empty")
	}
	// 1) history 进程内完成历史/生日/meta 修复并返回涉及的 device_no 列表
	hURL := hBase + t.HistoryInternalProjectionReconcilePath()
	raw, _ := json.Marshal(map[string]int{"limit": limit})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hURL, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("history projection reconcile: %w", err)
	}
	defer resp.Body.Close()
	var env gfEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return err
	}
	if resp.StatusCode >= 400 || env.Code != 0 {
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			return fmt.Errorf("history projection reconcile failed: status=%d", resp.StatusCode)
		}
		return fmt.Errorf("history projection reconcile failed: %s", msg)
	}
	var histData struct {
		DeviceNos []string `json:"deviceNos"`
	}
	if len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, &histData); err != nil {
			return err
		}
	}
	// 2) device 进程：按设备刷新画像缓存 + 全量事件/动作缓存
	dURL := dBase + t.DeviceInternalProjectionReconcilePath()
	raw2, _ := json.Marshal(map[string][]string{"deviceNos": histData.DeviceNos})
	req2, err := http.NewRequestWithContext(ctx, http.MethodPost, dURL, bytes.NewReader(raw2))
	if err != nil {
		return err
	}
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := client.Do(req2)
	if err != nil {
		return fmt.Errorf("device projection reconcile: %w", err)
	}
	defer resp2.Body.Close()
	var env2 gfEnvelope
	if err := json.NewDecoder(resp2.Body).Decode(&env2); err != nil {
		return err
	}
	if resp2.StatusCode >= 400 || env2.Code != 0 {
		msg := strings.TrimSpace(env2.Message)
		if msg == "" {
			glog.Warningf(ctx, "device projection reconcile failed: status=%d", resp2.StatusCode)
			return fmt.Errorf("device projection reconcile failed: status=%d", resp2.StatusCode)
		}
		glog.Warningf(ctx, "device projection reconcile failed: %s", msg)
		return fmt.Errorf("device projection reconcile failed: %s", msg)
	}
	return nil
}
