package device

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"hello/internal/platform/httpmeta"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// RemoteMerchantBlockHistoryCreate 经 device 探活：商户欠费是否禁 create。
// 失败时返回 block=true（fail-closed）。
func RemoteMerchantBlockHistoryCreate(ctx context.Context, deviceNo string) (bool, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return false, nil
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("DEVICE_SERVICE_URL")), "/")
	if base == "" {
		return true, gerror.NewCode(gcode.CodeInternalError, "DEVICE_SERVICE_URL 未配置")
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return true, gerror.NewCode(gcode.CodeInternalError, "DEVICE_GATEWAY_INTERNAL_SECRET 未配置")
	}
	u, err := url.Parse(base + "/device/internal/api/merchant/history-write-gate")
	if err != nil {
		return true, err
	}
	q := u.Query()
	q.Set("deviceNo", deviceNo)
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return true, err
	}
	req.Header.Set(httpmeta.HeaderDeviceGatewayInternalSecret, secret)
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return true, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return true, err
	}
	var env struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &env); err != nil {
			return true, fmt.Errorf("write-gate 响应非 JSON")
		}
	}
	if resp.StatusCode >= 400 || env.Code != 0 {
		return true, fmt.Errorf("write-gate failed: %s", env.Message)
	}
	var data struct {
		BlockCreate bool `json:"blockCreate"`
	}
	if len(env.Data) > 0 {
		_ = json.Unmarshal(env.Data, &data)
	}
	return data.BlockCreate, nil
}
