package device

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"hello/internal/model/entity"
	"hello/internal/services/contracts"
	sharedtypes "hello/internal/shared/types"
)

// httpDeviceAdminClient 通过 device-service HTTP 实现 DeviceAdminContract，
// 供 voice-service 等仅连接本域库的进程使用，避免跨进程直连他域 DAO。
type httpDeviceAdminClient struct {
	base   string
	client *http.Client
}

var (
	httpAdminOnce sync.Once
	httpAdminIns  contracts.DeviceAdminContract
)

// HTTPDeviceAdmin 返回单例 HTTP 版设备管理契约（默认走 DEVICE_SERVICE_URL）。
func HTTPDeviceAdmin() contracts.DeviceAdminContract {
	httpAdminOnce.Do(func() {
		t := contracts.ResolveHTTPTargets()
		httpAdminIns = &httpDeviceAdminClient{
			base:   strings.TrimRight(t.DeviceBaseURL, "/"),
			client: &http.Client{Timeout: 8 * time.Second},
		}
	})
	return httpAdminIns
}

type gfEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *httpDeviceAdminClient) doJSON(ctx context.Context, method, path string, query map[string]string, body interface{}, out interface{}) error {
	if c.base == "" {
		return fmt.Errorf("device http admin: base url empty, set %s", "DEVICE_SERVICE_URL")
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
	var bodyReader strings.Reader
	if body != nil {
		raw, mErr := json.Marshal(body)
		if mErr != nil {
			return mErr
		}
		bodyReader = *strings.NewReader(string(raw))
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), &bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var env gfEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return err
	}
	if resp.StatusCode >= 400 || env.Code != 0 {
		if strings.TrimSpace(env.Message) == "" {
			return fmt.Errorf("device http admin failed: status=%d path=%s", resp.StatusCode, path)
		}
		return fmt.Errorf("device http admin failed: %s", strings.TrimSpace(env.Message))
	}
	if out == nil || len(env.Data) == 0 || string(env.Data) == "null" {
		return nil
	}
	return json.Unmarshal(env.Data, out)
}

func (c *httpDeviceAdminClient) VerifyPassword(password string) bool {
	ctx := context.Background()
	var out struct {
		OK bool `json:"ok"`
	}
	err := c.doJSON(ctx, http.MethodPost, "/device/internal/api/admin/verify-password", nil, map[string]interface{}{
		"password": strings.TrimSpace(password),
	}, &out)
	return err == nil && out.OK
}

func (c *httpDeviceAdminClient) Register(ctx context.Context, deviceNo string) (string, error) {
	var out struct {
		ActiveTime string `json:"activeTime"`
	}
	err := c.doJSON(ctx, http.MethodPost, "/device/internal/api/register", nil, map[string]interface{}{
		"deviceNo": strings.TrimSpace(deviceNo),
	}, &out)
	return out.ActiveTime, err
}

func (c *httpDeviceAdminClient) EnsureRegistered(ctx context.Context, deviceNo string) error {
	return c.doJSON(ctx, http.MethodPost, "/device/internal/api/user/ensure", nil, map[string]interface{}{
		"deviceNo": strings.TrimSpace(deviceNo),
	}, nil)
}

func (c *httpDeviceAdminClient) UpdateLastTalk(ctx context.Context, deviceNo, ask, answer string) error {
	return c.doJSON(ctx, http.MethodPost, "/device/internal/api/user/last-talk", nil, map[string]interface{}{
		"deviceNo": strings.TrimSpace(deviceNo),
		"ask":      ask,
		"answer":   answer,
	}, nil)
}

func (c *httpDeviceAdminClient) List(ctx context.Context) ([]entity.User, error) {
	var out struct {
		List []entity.User `json:"list"`
	}
	err := c.doJSON(ctx, http.MethodGet, "/device/internal/api/user/list", nil, nil, &out)
	return out.List, err
}

func (c *httpDeviceAdminClient) AddEvent(ctx context.Context, name string, needQuantity int, extraNames string) error {
	return c.doJSON(ctx, http.MethodPost, "/device/internal/api/event/add", nil, map[string]interface{}{
		"name":         strings.TrimSpace(name),
		"needQuantity": needQuantity,
		"extraNames":   strings.TrimSpace(extraNames),
	}, nil)
}

func (c *httpDeviceAdminClient) ListEvents(ctx context.Context) ([]entity.Event, error) {
	var out struct {
		List []entity.Event `json:"list"`
	}
	err := c.doJSON(ctx, http.MethodGet, "/device/internal/api/event/options", nil, nil, &out)
	return out.List, err
}

func (c *httpDeviceAdminClient) UpdateEvent(ctx context.Context, id int64, name string, needQuantity int, extraNames string) error {
	return c.doJSON(ctx, http.MethodPost, "/device/internal/api/event/update", nil, map[string]interface{}{
		"id":           id,
		"name":         strings.TrimSpace(name),
		"needQuantity": needQuantity,
		"extraNames":   strings.TrimSpace(extraNames),
	}, nil)
}

func (c *httpDeviceAdminClient) DeleteEvent(ctx context.Context, id int64) error {
	return c.doJSON(ctx, http.MethodPost, "/device/internal/api/event/delete", nil, map[string]interface{}{
		"id": id,
	}, nil)
}

func (c *httpDeviceAdminClient) ListQA(ctx context.Context) ([]entity.Qa, error) {
	// qa 权威在 voice 库；若再请求 device 会形成 device↔voice 循环，故直连 voice 内部接口。
	t := contracts.ResolveHTTPTargets()
	return fetchQaListFromVoiceWithClient(ctx, c.client, t.VoiceInternalQaListURL())
}

func (c *httpDeviceAdminClient) ListActionsForAdmin(ctx context.Context) ([]sharedtypes.AdminActionItem, error) {
	var out struct {
		List []sharedtypes.AdminActionItem `json:"list"`
	}
	err := c.doJSON(ctx, http.MethodGet, "/device/internal/api/action/list", nil, nil, &out)
	return out.List, err
}

func (c *httpDeviceAdminClient) UpdateAction(ctx context.Context, id int64, name, targetType string) error {
	return c.doJSON(ctx, http.MethodPost, "/device/internal/api/action/update", nil, map[string]interface{}{
		"id":         id,
		"name":       strings.TrimSpace(name),
		"targetType": strings.TrimSpace(targetType),
	}, nil)
}

func (c *httpDeviceAdminClient) DeleteAction(ctx context.Context, id int64) error {
	return c.doJSON(ctx, http.MethodPost, "/device/internal/api/action/delete", nil, map[string]interface{}{
		"id": id,
	}, nil)
}

func (c *httpDeviceAdminClient) SaveUserProfile(ctx context.Context, deviceNo, birthday string, sex int) error {
	return c.doJSON(ctx, http.MethodPost, "/device/profile/api/save", nil, map[string]interface{}{
		"deviceNo": strings.TrimSpace(deviceNo),
		"birthday": strings.TrimSpace(birthday),
		"sex":      sex,
	}, nil)
}

func (c *httpDeviceAdminClient) InsertVoiceActionRecord(ctx context.Context, name, targetType string) error {
	return c.doJSON(ctx, http.MethodPost, "/device/internal/api/voice/action", nil, map[string]interface{}{
		"name":       strings.TrimSpace(name),
		"targetType": strings.TrimSpace(targetType),
	}, nil)
}

func (c *httpDeviceAdminClient) InsertOrGetEventByNeedle(ctx context.Context, needle string, needQuantity bool) (entity.Event, error) {
	var out struct {
		Item entity.Event `json:"item"`
	}
	err := c.doJSON(ctx, http.MethodPost, "/device/internal/api/voice/event/needle", nil, map[string]interface{}{
		"needle":       strings.TrimSpace(needle),
		"needQuantity": needQuantity,
	}, &out)
	return out.Item, err
}

func (c *httpDeviceAdminClient) ApplyDeepSeekEventExtractPersistence(ctx context.Context, out entity.Event) (entity.Event, string, error) {
	var resp struct {
		Item       entity.Event `json:"item"`
		TargetName string       `json:"targetName"`
	}
	err := c.doJSON(ctx, http.MethodPost, "/device/internal/api/voice/event/deepseek", nil, map[string]interface{}{
		"name":         strings.TrimSpace(out.Name),
		"extraNames":   strings.TrimSpace(out.ExtraNames),
		"needQuantity": out.NeedQuantity,
	}, &resp)
	return resp.Item, strings.TrimSpace(resp.TargetName), err
}
