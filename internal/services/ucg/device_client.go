package ucg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"hello/internal/services/contracts"
	"hello/internal/services/device"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// DeviceClient ucg-service 调用 device internal UCG API 的 HTTP 客户端（禁止直连 wx DAO）。
type DeviceClient struct {
	base   string
	secret string
	client *http.Client
}

var (
	deviceClientOnce sync.Once
	deviceClientIns  *DeviceClient
)

// Device 返回单例 device internal 客户端。
func Device() *DeviceClient {
	deviceClientOnce.Do(func() {
		deviceClientIns = &DeviceClient{
			base:   resolveDeviceServiceBaseURL(),
			secret: resolveDeviceInternalSecret(),
			client: &http.Client{Timeout: 8 * time.Second},
		}
	})
	return deviceClientIns
}

func resolveDeviceServiceBaseURL() string {
	if v := strings.TrimSpace(os.Getenv("DEVICE_SERVICE_URL")); v != "" {
		return strings.TrimRight(v, "/")
	}
	ctx := context.Background()
	if v := strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.deviceServiceUrl").String()); v != "" {
		return strings.TrimRight(v, "/")
	}
	return strings.TrimRight(contracts.ResolveHTTPTargets().DeviceBaseURL, "/")
}

func resolveDeviceInternalSecret() string {
	if v := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET")); v != "" {
		return v
	}
	ctx := context.Background()
	return strings.TrimSpace(g.Cfg().MustGet(ctx, "ucg.deviceInternalSecret").String())
}

type gfEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// ValidateWx 校验 wxId 是否存在。
func (c *DeviceClient) ValidateWx(ctx context.Context, wxID int64) (exists bool, babyName string, err error) {
	var data struct {
		Exists   bool   `json:"exists"`
		BabyName string `json:"babyName"`
	}
	if err = c.doJSON(ctx, http.MethodPost, "/device/internal/api/ucg/wx/validate", nil, map[string]interface{}{
		"wxId": wxID,
	}, &data); err != nil {
		return false, "", err
	}
	return data.Exists, data.BabyName, nil
}

// BabyName 获取默认昵称所需的 baby_name。
func (c *DeviceClient) BabyName(ctx context.Context, wxID int64) (string, error) {
	path := fmt.Sprintf("/device/internal/api/ucg/wx/%d/baby-name", wxID)
	var data struct {
		BabyName string `json:"babyName"`
	}
	if err := c.doJSON(ctx, http.MethodGet, path, nil, nil, &data); err != nil {
		return "", err
	}
	return data.BabyName, nil
}

// UpdateIpLocation 写入 wx IP 属地。
func (c *DeviceClient) UpdateIpLocation(ctx context.Context, wxID int64, ipLocation string) error {
	path := fmt.Sprintf("/device/internal/api/ucg/wx/%d/ip-location", wxID)
	return c.doJSON(ctx, http.MethodPut, path, nil, map[string]interface{}{
		"ipLocation": ipLocation,
	}, nil)
}

// BatchWx 批量拉取展示字段。
func (c *DeviceClient) BatchWx(ctx context.Context, wxIDs []int64) (map[int64]device.UcgWxDisplay, error) {
	var data struct {
		List []device.UcgWxDisplay `json:"list"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/device/internal/api/ucg/wx/batch", nil, map[string]interface{}{
		"wxIds": wxIDs,
	}, &data); err != nil {
		return nil, err
	}
	out := make(map[int64]device.UcgWxDisplay, len(data.List))
	for _, item := range data.List {
		out[item.WxId] = item
	}
	return out, nil
}

func (c *DeviceClient) doJSON(ctx context.Context, method, path string, query map[string]string, body interface{}, out interface{}) error {
	if c.base == "" {
		return deviceClientErr("DEVICE_SERVICE_URL 未配置")
	}
	if strings.TrimSpace(c.secret) == "" {
		return deviceClientErr("DEVICE_GATEWAY_INTERNAL_SECRET 未配置")
	}
	u, err := url.Parse(c.base + path)
	if err != nil {
		return err
	}
	if len(query) > 0 {
		q := u.Query()
		for k, v := range query {
			if strings.TrimSpace(v) != "" {
				q.Set(k, strings.TrimSpace(v))
			}
		}
		u.RawQuery = q.Encode()
	}
	var bodyReader *strings.Reader
	if body != nil {
		raw, mErr := json.Marshal(body)
		if mErr != nil {
			return mErr
		}
		bodyReader = strings.NewReader(string(raw))
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set(device.HeaderDeviceGatewayInternalSecret, c.secret)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return deviceClientErr(fmt.Sprintf("device-service 不可达: %v", err))
	}
	defer resp.Body.Close()
	rawBody, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return deviceClientErr(fmt.Sprintf("读取 device 响应失败: %v", err))
	}
	var env gfEnvelope
	if len(rawBody) > 0 {
		if err = json.Unmarshal(rawBody, &env); err != nil {
			snippet := strings.TrimSpace(string(rawBody))
			if len(snippet) > 160 {
				snippet = snippet[:160] + "..."
			}
			return deviceClientErr(fmt.Sprintf("device 响应非 JSON（HTTP %d）: %s", resp.StatusCode, snippet))
		}
	}
	if resp.StatusCode == http.StatusForbidden || env.Code == 403 {
		return deviceClientErr("device 内部接口鉴权失败，请检查 DEVICE_GATEWAY_INTERNAL_SECRET")
	}
	if resp.StatusCode >= 400 {
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return deviceClientErr(msg)
	}
	if env.Code != 0 {
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			msg = fmt.Sprintf("code=%d", env.Code)
		}
		return deviceClientErr(msg)
	}
	if out != nil && len(env.Data) > 0 && string(env.Data) != "null" {
		if err = json.Unmarshal(env.Data, out); err != nil {
			return deviceClientErr(fmt.Sprintf("解析 device 响应失败: %v", err))
		}
	}
	return nil
}

func deviceClientErr(detail string) error {
	return gerror.NewCode(gcode.CodeInternalError, "device internal ucg: "+detail)
}
