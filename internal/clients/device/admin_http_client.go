// Package device 出站 device-service HTTP 客户端（含 DeviceAdmin 契约）。
package device

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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

func (c *httpDeviceAdminClient) Register(ctx context.Context, deviceNo string) (int64, error) {
	var out struct {
		ActiveTime int64 `json:"activeTime"`
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

func (c *httpDeviceAdminClient) ListUsersPage(ctx context.Context, page, pageSize int, q string) (contracts.UserPageResult, error) {
	var out contracts.UserPageResult
	err := c.doJSON(ctx, http.MethodGet, "/device/internal/api/user/list-page", map[string]string{
		"page":     fmt.Sprintf("%d", page),
		"pageSize": fmt.Sprintf("%d", pageSize),
		"q":        strings.TrimSpace(q),
	}, nil, &out)
	return out, err
}

// ListWxPage 经 device 内部 wx 分页契约拉取（含 babyName），无需管理口令。
func (c *httpDeviceAdminClient) ListWxPage(ctx context.Context, page, pageSize int, q string) (contracts.WxPageResult, error) {
	var out struct {
		List     []contracts.AdminWxListItem `json:"list"`
		Total    int                         `json:"total"`
		Page     int                         `json:"page"`
		PageSize int                         `json:"pageSize"`
	}
	err := c.doJSON(ctx, http.MethodGet, "/device/internal/api/wx/list-page", map[string]string{
		"page":     fmt.Sprintf("%d", page),
		"pageSize": fmt.Sprintf("%d", pageSize),
		"q":        strings.TrimSpace(q),
	}, nil, &out)
	if err != nil {
		return contracts.WxPageResult{}, err
	}
	return contracts.WxPageResult{List: out.List, Total: out.Total, Page: out.Page, PageSize: out.PageSize}, nil
}

func (c *httpDeviceAdminClient) TouchLastAPIAccess(ctx context.Context, deviceNo, apiPath string, atUnixSec int64) error {
	return c.doJSON(ctx, http.MethodPost, "/device/internal/api/user/touch-api-access", nil, map[string]interface{}{
		"deviceNo": strings.TrimSpace(deviceNo),
		"apiPath":  strings.TrimSpace(apiPath),
		"at":       atUnixSec,
	}, nil)
}

func (c *httpDeviceAdminClient) AddEvent(ctx context.Context, name string, eventType string, extraNames, color, unit, logoPath string, parentID int64) (int64, error) {
	err := c.doJSON(ctx, http.MethodPost, "/device/internal/api/event/add", nil, map[string]interface{}{
		"name":       strings.TrimSpace(name),
		"eventType":  normalizeEventType(eventType),
		"extraNames": strings.TrimSpace(extraNames),
		"color":      strings.TrimSpace(color),
		"unit":       strings.TrimSpace(unit),
		"logo":       strings.TrimSpace(logoPath),
		"parentId":   parentID,
	}, nil)
	return 0, err
}

func (c *httpDeviceAdminClient) ListEvents(ctx context.Context) ([]entity.Event, error) {
	var out struct {
		List []entity.Event `json:"list"`
	}
	err := c.doJSON(ctx, http.MethodGet, "/device/internal/api/event/options", nil, nil, &out)
	return out.List, err
}

func (c *httpDeviceAdminClient) UpdateEvent(ctx context.Context, id int64, name string, eventType string, extraNames, color, unit, logoPath string, parentID *int64) error {
	body := map[string]interface{}{
		"id":         id,
		"name":       strings.TrimSpace(name),
		"eventType":  normalizeEventType(eventType),
		"extraNames": strings.TrimSpace(extraNames),
		"color":      strings.TrimSpace(color),
		"unit":       strings.TrimSpace(unit),
		"logo":       strings.TrimSpace(logoPath),
	}
	if parentID != nil {
		body["parentId"] = normalizeEventParentID(*parentID)
	}
	return c.doJSON(ctx, http.MethodPost, "/device/internal/api/event/update", nil, body, nil)
}

func (c *httpDeviceAdminClient) DeleteEvent(ctx context.Context, id int64) error {
	return c.doJSON(ctx, http.MethodPost, "/device/internal/api/event/delete", nil, map[string]interface{}{
		"id": id,
	}, nil)
}

func (c *httpDeviceAdminClient) ListQAPage(ctx context.Context, page, pageSize int) (contracts.QaPageResult, error) {
	t := contracts.ResolveHTTPTargets()
	return fetchQaPageFromVoiceWithClient(ctx, c.client, t.VoiceInternalQaListURL(), page, pageSize)
}

func (c *httpDeviceAdminClient) DeleteQA(ctx context.Context, id int64) error {
	return deleteQaFromVoiceHTTP(ctx, id)
}

func (c *httpDeviceAdminClient) ListFeedbackByWxID(ctx context.Context, wxID int64) ([]entity.Feedback, error) {
	return nil, fmt.Errorf("device http admin: feedback APIs are device-service local only")
}

func (c *httpDeviceAdminClient) SubmitFeedback(ctx context.Context, wxID int64, question string) (entity.Feedback, error) {
	return entity.Feedback{}, fmt.Errorf("device http admin: feedback APIs are device-service local only")
}

func (c *httpDeviceAdminClient) ListFeedbackPage(ctx context.Context, page, pageSize int, unrepliedOnly bool) (contracts.FeedbackPageResult, error) {
	return contracts.FeedbackPageResult{}, fmt.Errorf("device http admin: feedback APIs are device-service local only")
}

func (c *httpDeviceAdminClient) ReplyFeedback(ctx context.Context, id int64, officialReply string) error {
	return fmt.Errorf("device http admin: feedback APIs are device-service local only")
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

func (c *httpDeviceAdminClient) SaveUserProfile(ctx context.Context, deviceNo string, babyName string, birthdayUnixSec int64, sex int) error {
	return c.doJSON(ctx, http.MethodPost, "/device/app/api/user/save", nil, map[string]interface{}{
		"deviceNo": strings.TrimSpace(deviceNo),
		"babyName": strings.TrimSpace(babyName),
		"birthday": birthdayUnixSec,
		"sex":      sex,
	}, nil)
}

func (c *httpDeviceAdminClient) InsertVoiceActionRecord(ctx context.Context, name, targetType string) error {
	return c.doJSON(ctx, http.MethodPost, "/device/internal/api/voice/action", nil, map[string]interface{}{
		"name":       strings.TrimSpace(name),
		"targetType": strings.TrimSpace(targetType),
	}, nil)
}

func (c *httpDeviceAdminClient) InsertOrGetEventByNeedle(ctx context.Context, needle string, eventType, unit string) (entity.Event, error) {
	var out struct {
		Item entity.Event `json:"item"`
	}
	err := c.doJSON(ctx, http.MethodPost, "/device/internal/api/voice/event/needle", nil, map[string]interface{}{
		"needle":    strings.TrimSpace(needle),
		"eventType": normalizeEventType(eventType),
		"unit":      strings.TrimSpace(unit),
	}, &out)
	return out.Item, err
}

func (c *httpDeviceAdminClient) ApplyDeepSeekEventExtractPersistence(ctx context.Context, out entity.Event) (entity.Event, string, error) {
	var resp struct {
		Item       entity.Event `json:"item"`
		TargetName string       `json:"targetName"`
	}
	err := c.doJSON(ctx, http.MethodPost, "/device/internal/api/voice/event/deepseek", nil, map[string]interface{}{
		"name":       strings.TrimSpace(out.Name),
		"extraNames": strings.TrimSpace(out.ExtraNames),
		"eventType":  normalizeEventType(out.EventType),
		"unit":       strings.TrimSpace(out.Unit),
	}, &resp)
	return resp.Item, strings.TrimSpace(resp.TargetName), err
}


// normalizeEventType 规范化事件类型（clients 侧副本，避免依赖 device 实现包）。
func normalizeEventType(eventType string) string {
	t := strings.TrimSpace(strings.ToLower(eventType))
	switch t {
	case "number", "time", "one":
		return t
	default:
		return "time"
	}
}

func normalizeEventParentID(parentID int64) int64 {
	if parentID < 0 {
		return 0
	}
	return parentID
}


// fetchQaPageFromVoiceHTTP device 域分页拉取问答库（qa 表仅在 voice 库）。
func fetchQaPageFromVoiceHTTP(ctx context.Context, page, pageSize int) (contracts.QaPageResult, error) {
	t := contracts.ResolveHTTPTargets()
	client := &http.Client{Timeout: 8 * time.Second}
	return fetchQaPageFromVoiceWithClient(ctx, client, t.VoiceInternalQaListURL(), page, pageSize)
}

func fetchQaPageFromVoiceWithClient(ctx context.Context, client *http.Client, baseURL string, page, pageSize int) (contracts.QaPageResult, error) {
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || u.String() == "" {
		return contracts.QaPageResult{}, fmt.Errorf("voice QA list: url empty, set VOICE_SERVICE_URL for device-service")
	}
	q := u.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("pageSize", strconv.Itoa(pageSize))
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return contracts.QaPageResult{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return contracts.QaPageResult{}, err
	}
	defer resp.Body.Close()
	var env gfEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return contracts.QaPageResult{}, err
	}
	if resp.StatusCode >= 400 || env.Code != 0 {
		if strings.TrimSpace(env.Message) == "" {
			return contracts.QaPageResult{}, fmt.Errorf("voice QA list failed: status=%d url=%s", resp.StatusCode, u.String())
		}
		return contracts.QaPageResult{}, fmt.Errorf("voice QA list failed: %s", strings.TrimSpace(env.Message))
	}
	var out contracts.QaPageResult
	if len(env.Data) == 0 || string(env.Data) == "null" {
		return out, nil
	}
	if err := json.Unmarshal(env.Data, &out); err != nil {
		return contracts.QaPageResult{}, err
	}
	return out, nil
}

// deleteQaFromVoiceHTTP 经 voice 内部接口删除问答库行。
func deleteQaFromVoiceHTTP(ctx context.Context, id int64) error {
	t := contracts.ResolveHTTPTargets()
	client := &http.Client{Timeout: 8 * time.Second}
	u := strings.TrimSpace(t.VoiceInternalQaDeleteURL())
	if u == "" {
		return fmt.Errorf("voice QA delete: url empty, set VOICE_SERVICE_URL for device-service")
	}
	body, _ := json.Marshal(map[string]interface{}{"id": id})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
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
			return fmt.Errorf("voice QA delete failed: status=%d", resp.StatusCode)
		}
		return fmt.Errorf("voice QA delete failed: %s", strings.TrimSpace(env.Message))
	}
	return nil
}

// fetchQaListFromVoiceHTTP 兼容旧调用：取第一页最多 100 条（id 倒序）。
func fetchQaListFromVoiceHTTP(ctx context.Context) ([]entity.Qa, error) {
	res, err := fetchQaPageFromVoiceHTTP(ctx, 1, 100)
	if err != nil {
		return nil, err
	}
	return res.List, nil
}

// fetchQaListFromVoiceWithClient 兼容 HTTP 客户端全量拉取（最多 100 条）。
func fetchQaListFromVoiceWithClient(ctx context.Context, client *http.Client, fullURL string) ([]entity.Qa, error) {
	res, err := fetchQaPageFromVoiceWithClient(ctx, client, fullURL, 1, 100)
	if err != nil {
		return nil, err
	}
	return res.List, nil
}
