package device

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"hello/internal/model/entity"
	"hello/internal/services/contracts"
)

// fetchQaListFromVoiceHTTP device 域拉取问答库列表：qa 表仅在 voice 库，禁止 device 进程内直连 dao.Qa。
func fetchQaListFromVoiceHTTP(ctx context.Context) ([]entity.Qa, error) {
	t := contracts.ResolveHTTPTargets()
	client := &http.Client{Timeout: 8 * time.Second}
	return fetchQaListFromVoiceWithClient(ctx, client, t.VoiceInternalQaListURL())
}

// fetchQaListFromVoiceWithClient 使用指定 HTTP 客户端请求 voice 内部 QA 列表并解析 GoFrame 标准封套。
func fetchQaListFromVoiceWithClient(ctx context.Context, client *http.Client, fullURL string) ([]entity.Qa, error) {
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	u := strings.TrimSpace(fullURL)
	if u == "" {
		return nil, fmt.Errorf("voice QA list: url empty, set VOICE_SERVICE_URL for device-service")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var env gfEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 || env.Code != 0 {
		if strings.TrimSpace(env.Message) == "" {
			return nil, fmt.Errorf("voice QA list failed: status=%d url=%s", resp.StatusCode, u)
		}
		return nil, fmt.Errorf("voice QA list failed: %s", strings.TrimSpace(env.Message))
	}
	var out struct {
		List []entity.Qa `json:"list"`
	}
	if len(env.Data) == 0 || string(env.Data) == "null" {
		return out.List, nil
	}
	if err := json.Unmarshal(env.Data, &out); err != nil {
		return nil, err
	}
	return out.List, nil
}
