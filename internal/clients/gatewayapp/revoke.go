// Package gatewayapp 出站 gateway-app-server HTTP 客户端（会话吊销等内部契约）。
package gatewayapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"hello/internal/platform/httpmeta"
)

// RevokeSessions 调用网关内部接口 bump session epoch，吊销指定 wx 与/或纯设备 refresh。
// Args: wxIDs 可为空；deviceNo 非空时吊销 wxId=0 上该号会话。
// Side Effects: HTTP POST GATEWAY_APP_SERVICE_URL；须配置 DEVICE_GATEWAY_INTERNAL_SECRET。
func RevokeSessions(ctx context.Context, wxIDs []int64, deviceNo string) error {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("GATEWAY_APP_SERVICE_URL")), "/")
	if base == "" {
		return fmt.Errorf("GATEWAY_APP_SERVICE_URL 未配置")
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return fmt.Errorf("DEVICE_GATEWAY_INTERNAL_SECRET 未配置")
	}
	body, err := json.Marshal(map[string]interface{}{
		"wxIds":    wxIDs,
		"deviceNo": strings.TrimSpace(deviceNo),
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/device/app/api/internal/sessions/revoke", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(httpmeta.HeaderDeviceGatewayInternalSecret, secret)
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var env struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return err
	}
	if resp.StatusCode >= 400 || env.Code != 0 {
		if strings.TrimSpace(env.Message) == "" {
			return fmt.Errorf("gateway revoke sessions failed: status=%d", resp.StatusCode)
		}
		return fmt.Errorf("gateway revoke sessions failed: %s", strings.TrimSpace(env.Message))
	}
	return nil
}
