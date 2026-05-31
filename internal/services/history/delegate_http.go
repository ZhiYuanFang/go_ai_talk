package history

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"hello/internal/model/entity"
	"hello/internal/services/contracts"

	"github.com/gogf/gf/v2/os/glog"
)

// localHTTPEnvelope 与网关 MiddlewareHandlerResponse 对齐的通用响应壳。
type localHTTPEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func localHTTPDoJSON(ctx context.Context, method, baseURL, path string, query map[string]string, body interface{}, out interface{}) error {
	if strings.TrimSpace(baseURL) == "" {
		return fmt.Errorf("local delegate base url is empty for path %s", path)
	}
	reqURL, err := url.Parse(strings.TrimRight(baseURL, "/") + path)
	if err != nil {
		return err
	}
	if len(query) > 0 {
		q := reqURL.Query()
		for k, v := range query {
			if strings.TrimSpace(v) != "" {
				q.Set(k, strings.TrimSpace(v))
			}
		}
		reqURL.RawQuery = q.Encode()
	}
	var bodyReader strings.Reader
	if body != nil {
		raw, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			return marshalErr
		}
		bodyReader = *strings.NewReader(string(raw))
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), &bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var env localHTTPEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		if strings.TrimSpace(env.Message) == "" {
			return fmt.Errorf("local delegate failed: status=%d path=%s", resp.StatusCode, path)
		}
		return fmt.Errorf("local delegate failed: %s", strings.TrimSpace(env.Message))
	}
	if env.Code != 0 {
		return fmt.Errorf("local delegate business failed: %s", strings.TrimSpace(env.Message))
	}
	if out == nil || len(env.Data) == 0 || string(env.Data) == "null" {
		return nil
	}
	return json.Unmarshal(env.Data, out)
}

func delegateListSuggest(ctx context.Context, deviceNo string) ([]entity.Suggest, error) {
	t := contracts.ResolveHTTPTargets()
	var resp struct {
		List []entity.Suggest `json:"list"`
	}
	if err := localHTTPDoJSON(ctx, http.MethodGet, t.VoiceBaseURL, t.VoiceSuggestListPath(), map[string]string{"deviceNo": strings.TrimSpace(deviceNo)}, nil, &resp); err != nil {
		return nil, err
	}
	return resp.List, nil
}

func delegateDeleteSuggest(ctx context.Context, id int64, deviceNo string) error {
	t := contracts.ResolveHTTPTargets()
	return localHTTPDoJSON(ctx, http.MethodPost, t.VoiceBaseURL, t.VoiceSuggestDeletePath(), nil, map[string]interface{}{
		"id":       id,
		"deviceNo": strings.TrimSpace(deviceNo),
	}, nil)
}

func delegateListEventOptions(ctx context.Context) ([]entity.Event, error) {
	t := contracts.ResolveHTTPTargets()
	var resp struct {
		List []entity.Event `json:"list"`
	}
	if err := localHTTPDoJSON(ctx, http.MethodGet, t.DeviceBaseURL, t.DeviceInternalEventOptionsPath(), nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp.List, nil
}

func delegateGetProfile(ctx context.Context, deviceNo string) (babyName string, birthdayUnixSec int64, sex int, err error) {
	t := contracts.ResolveHTTPTargets()
	var resp struct {
		BabyName string `json:"babyName"`
		Birthday int64  `json:"birthday"`
		Sex      int    `json:"sex"`
	}
	if err := localHTTPDoJSON(ctx, http.MethodGet, t.DeviceBaseURL, t.DeviceProfileGetPath(), map[string]string{"deviceNo": strings.TrimSpace(deviceNo)}, nil, &resp); err != nil {
		return "", 0, 0, err
	}
	return strings.TrimSpace(resp.BabyName), resp.Birthday, resp.Sex, nil
}

func delegateSaveProfile(ctx context.Context, deviceNo string, babyName string, birthdayUnixSec int64, sex int) error {
	t := contracts.ResolveHTTPTargets()
	return localHTTPDoJSON(ctx, http.MethodPost, t.DeviceBaseURL, t.DeviceProfileSavePath(), nil, map[string]interface{}{
		"deviceNo": strings.TrimSpace(deviceNo),
		"babyName": strings.TrimSpace(babyName),
		"birthday": birthdayUnixSec,
		"sex":      sex,
	}, nil)
}

func logDelegateFailure(ctx context.Context, stage string, err error) {
	if err != nil {
		glog.Warningf(ctx, "[history-local-delegate] %s err=%v", stage, err)
	}
}
