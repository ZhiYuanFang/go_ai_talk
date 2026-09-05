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
	"time"

	"hello/internal/model/entity"
	"hello/internal/services/contracts"
)

// gfEnvelope GoFrame 标准 JSON 响应外壳（voice QA HTTP 客户端解析用）。
type gfEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
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
