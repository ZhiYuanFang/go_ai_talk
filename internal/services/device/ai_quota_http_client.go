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
	"sync"
	"time"

	"hello/internal/services/contracts"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// AIQuotaHTTPClient 跨服务调用 device internal ai-quota API。
type AIQuotaHTTPClient struct {
	base   string
	secret string
	client *http.Client
}

var (
	aiQuotaHTTPOnce sync.Once
	aiQuotaHTTPIns  *AIQuotaHTTPClient
)

// AIQuotaHTTP 返回单例客户端。
func AIQuotaHTTP() *AIQuotaHTTPClient {
	aiQuotaHTTPOnce.Do(func() {
		base := strings.TrimRight(contracts.ResolveHTTPTargets().DeviceBaseURL, "/")
		if v := strings.TrimSpace(os.Getenv("DEVICE_SERVICE_URL")); v != "" {
			base = strings.TrimRight(v, "/")
		}
		secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
		if secret == "" {
			secret = strings.TrimSpace(g.Cfg().MustGet(context.Background(), "deviceInternalSecret").String())
		}
		aiQuotaHTTPIns = &AIQuotaHTTPClient{
			base:   base,
			secret: secret,
			client: &http.Client{Timeout: 8 * time.Second},
		}
	})
	return aiQuotaHTTPIns
}

type aiQuotaGFEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *AIQuotaHTTPClient) doJSON(ctx context.Context, method, path string, query map[string]string, body interface{}, out interface{}) error {
	if c.base == "" {
		return gerror.NewCode(gcode.CodeInternalError, "DEVICE_SERVICE_URL 未配置")
	}
	if strings.TrimSpace(c.secret) == "" {
		return gerror.NewCode(gcode.CodeInternalError, "DEVICE_GATEWAY_INTERNAL_SECRET 未配置")
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
	req.Header.Set(HeaderDeviceGatewayInternalSecret, c.secret)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return gerror.NewCode(gcode.CodeOperationFailed, fmt.Sprintf("device ai-quota 不可达: %v", err))
	}
	defer resp.Body.Close()
	rawBody, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return err
	}
	var env aiQuotaGFEnvelope
	if len(rawBody) > 0 {
		if err = json.Unmarshal(rawBody, &env); err != nil {
			return gerror.NewCode(gcode.CodeInternalError, "device ai-quota 响应非 JSON")
		}
	}
	if env.Code == CodeAINotLoggedIn {
		return ErrAINotLoggedIn
	}
	if env.Code == CodeAIQuotaExhausted {
		return ErrAIQuotaExhausted
	}
	if resp.StatusCode >= 400 || env.Code != 0 {
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			msg = fmt.Sprintf("device ai-quota HTTP %d", resp.StatusCode)
		}
		return gerror.NewCode(gcode.CodeOperationFailed, msg)
	}
	if out != nil && len(env.Data) > 0 && string(env.Data) != "null" {
		if err = json.Unmarshal(env.Data, out); err != nil {
			return gerror.NewCode(gcode.CodeInternalError, "解析 device ai-quota 响应失败")
		}
	}
	return nil
}

// RemoteCheck 远程预检额度。
func (c *AIQuotaHTTPClient) RemoteCheck(ctx context.Context, wxID int64, feature AIQuotaFeature) (AIQuotaSnapshot, error) {
	var data struct {
		Allowed bool `json:"allowed"`
		Used    int  `json:"used"`
		Limit   int  `json:"limit"`
	}
	err := c.doJSON(ctx, http.MethodPost, "/device/internal/api/ai-quota/check", nil, map[string]interface{}{
		"wxId":    wxID,
		"feature": string(feature),
	}, &data)
	if err != nil {
		return AIQuotaSnapshot{}, err
	}
	return AIQuotaSnapshot{Used: data.Used, Limit: data.Limit, Allowed: data.Allowed}, nil
}

// RemoteConsume 远程扣减额度。
func (c *AIQuotaHTTPClient) RemoteConsume(ctx context.Context, wxID int64, feature AIQuotaFeature) (AIQuotaSnapshot, error) {
	var data struct {
		Used  int `json:"used"`
		Limit int `json:"limit"`
	}
	err := c.doJSON(ctx, http.MethodPost, "/device/internal/api/ai-quota/consume", nil, map[string]interface{}{
		"wxId":    wxID,
		"feature": string(feature),
	}, &data)
	if err != nil {
		return AIQuotaSnapshot{}, err
	}
	return AIQuotaSnapshot{Used: data.Used, Limit: data.Limit, Allowed: data.Used <= data.Limit}, nil
}

// RemoteWxIDByDeviceNo 远程按 deviceNo 查 wxId。
func (c *AIQuotaHTTPClient) RemoteWxIDByDeviceNo(ctx context.Context, deviceNo string) (int64, error) {
	var data struct {
		WxId int64 `json:"wxId"`
	}
	err := c.doJSON(ctx, http.MethodGet, "/device/internal/api/ai-quota/wx-id-by-device-no", map[string]string{
		"deviceNo": strings.TrimSpace(deviceNo),
	}, nil, &data)
	return data.WxId, err
}

// RemoteGetDefaultAdmin 读取全局默认（ucg admin 代理）。
func (c *AIQuotaHTTPClient) RemoteGetDefaultAdmin(ctx context.Context) (AIQuotaDefaultDTO, error) {
	var data AIQuotaDefaultDTO
	err := c.doJSON(ctx, http.MethodGet, "/device/internal/api/ai-quota/default", nil, nil, &data)
	return data, err
}

// RemotePutDefaultAdmin 更新全局默认。
func (c *AIQuotaHTTPClient) RemotePutDefaultAdmin(ctx context.Context, polish, voice int) (AIQuotaDefaultDTO, error) {
	var data AIQuotaDefaultDTO
	err := c.doJSON(ctx, http.MethodPut, "/device/internal/api/ai-quota/default", nil, map[string]interface{}{
		"polishMonthlyLimit":  polish,
		"voiceAiMonthlyLimit": voice,
	}, &data)
	return data, err
}

// RemoteGetUserOverrideAdmin 读取 wxId override。
func (c *AIQuotaHTTPClient) RemoteGetUserOverrideAdmin(ctx context.Context, wxID int64) (AIQuotaUserOverrideDTO, error) {
	var data AIQuotaUserOverrideDTO
	err := c.doJSON(ctx, http.MethodGet, "/device/internal/api/ai-quota/user", map[string]string{
		"wxId": fmt.Sprintf("%d", wxID),
	}, nil, &data)
	return data, err
}

// RemotePutUserOverrideAdmin 更新 wxId override。
func (c *AIQuotaHTTPClient) RemotePutUserOverrideAdmin(ctx context.Context, wxID int64, polish, voice *int, clearPolish, clearVoice bool) (AIQuotaUserOverrideDTO, error) {
	var data AIQuotaUserOverrideDTO
	err := c.doJSON(ctx, http.MethodPut, "/device/internal/api/ai-quota/user", nil, map[string]interface{}{
		"wxId":                wxID,
		"polishMonthlyLimit":  polish,
		"voiceAiMonthlyLimit": voice,
		"clearPolish":         clearPolish,
		"clearVoiceAi":        clearVoice,
	}, &data)
	return data, err
}
