package history

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"hello/internal/model/entity"
	"hello/internal/services/contracts"
	"hello/internal/services/device"
	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
)

// localHTTPEnvelope 与网关 MiddlewareHandlerResponse 对齐的通用响应壳。
type localHTTPEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type delegateHTTPOptions struct {
	headers map[string]string
	timeout time.Duration
}

func localHTTPDoJSON(ctx context.Context, method, baseURL, path string, query map[string]string, body interface{}, out interface{}) error {
	return localHTTPDoJSONWithOpts(ctx, method, baseURL, path, query, body, out, delegateHTTPOptions{})
}

func localHTTPDoJSONWithOpts(ctx context.Context, method, baseURL, path string, query map[string]string, body interface{}, out interface{}, opts delegateHTTPOptions) error {
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
	var reqBody io.Reader = http.NoBody
	if body != nil {
		raw, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			return marshalErr
		}
		reqBody = strings.NewReader(string(raw))
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), reqBody)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range opts.headers {
		if strings.TrimSpace(k) != "" && strings.TrimSpace(v) != "" {
			req.Header.Set(k, strings.TrimSpace(v))
		}
	}
	timeout := opts.timeout
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	client := &http.Client{Timeout: timeout}
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
		msg := strings.TrimSpace(env.Message)
		if env.Code == contracts.CodeAINotLoggedIn {
			return gerror.NewCode(contracts.GCodeAINotLoggedIn(), msg)
		}
		if env.Code == contracts.CodeAIQuotaExhausted {
			return gerror.NewCode(contracts.GCodeAIQuotaExhausted(), msg)
		}
		return fmt.Errorf("local delegate business failed: %s", msg)
	}
	if out == nil || len(env.Data) == 0 || string(env.Data) == "null" {
		return nil
	}
	return json.Unmarshal(env.Data, out)
}

// DelegateTextChat 经 voice-service internal HTTP 执行文本对话，禁止 history 直查 voice 库。
func DelegateTextChat(ctx context.Context, deviceNo, transcript string, wxID int64) (string, error) {
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return "", fmt.Errorf("DEVICE_GATEWAY_INTERNAL_SECRET 未配置")
	}
	headers := map[string]string{
		device.HeaderDeviceGatewayInternalSecret: secret,
	}
	if wxID > 0 {
		headers[gatewayapp.HeaderInternalWxId] = strconv.FormatInt(wxID, 10)
	}
	t := contracts.ResolveHTTPTargets()
	var resp struct {
		Reply string `json:"reply"`
	}
	err := localHTTPDoJSONWithOpts(ctx, http.MethodPost, t.VoiceBaseURL, t.VoiceInternalTextChatPath(), nil, map[string]interface{}{
		"deviceNo":   strings.TrimSpace(deviceNo),
		"transcript": strings.TrimSpace(transcript),
	}, &resp, delegateHTTPOptions{
		headers: headers,
		timeout: 30 * time.Second,
	})
	if err != nil {
		return "", err
	}
	return resp.Reply, nil
}

// DelegateTextChatStream 经 voice-service internal HTTP 执行流式文本对话（SSE）。
// cb.OnThinking 用于实时转发思考过程；event: answer 仅为业务话术，累积后作为返回值（不经回调）。
func DelegateTextChatStream(ctx context.Context, deviceNo, transcript string, wxID int64, cb *contracts.IntentStreamCallback) (string, error) {
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return "", fmt.Errorf("DEVICE_GATEWAY_INTERNAL_SECRET 未配置")
	}
	t := contracts.ResolveHTTPTargets()
	// 1. 构造请求体
	reqBody, _ := json.Marshal(map[string]interface{}{
		"deviceNo":   strings.TrimSpace(deviceNo),
		"transcript": strings.TrimSpace(transcript),
	})
	// 2. 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(t.VoiceBaseURL, "/")+t.VoiceInternalTextChatStreamPath(), strings.NewReader(string(reqBody)))
	if err != nil {
		return "", fmt.Errorf("创建流式对话请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set(device.HeaderDeviceGatewayInternalSecret, secret)
	if wxID > 0 {
		httpReq.Header.Set(gatewayapp.HeaderInternalWxId, strconv.FormatInt(wxID, 10))
	}
	// 3. 发送请求
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("调用流式对话服务失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("流式对话服务返回错误状态码 %d: %s", resp.StatusCode, string(respBody))
	}
	// 4. 逐行解析 SSE：thinking 实时回调；answer 累积为业务 Reply；error 记下来但不中断，以便仍能读到降级话术
	var reply string
	var streamErrMsg string
	scanner := bufio.NewScanner(resp.Body)
	var currentEvent string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		switch currentEvent {
		case "thinking":
			if cb != nil && cb.OnThinking != nil {
				if cbErr := cb.OnThinking(data); cbErr != nil {
					return reply, cbErr
				}
			}
		case "answer":
			// 业务话术：只累积（voice 保证 answer=Reply，非意图 JSON）
			reply += data
		case "error":
			streamErrMsg = data
		}
	}
	if err := scanner.Err(); err != nil {
		return reply, err
	}
	if streamErrMsg != "" {
		return reply, fmt.Errorf("%s", streamErrMsg)
	}
	return reply, nil
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

// resolveEventUnitFromDevice 经 device-service HTTP 契约按 eventId 解析单位；禁止 history 直查 event 表。
func resolveEventUnitFromDevice(ctx context.Context, eventID int64) string {
	if eventID <= 0 {
		return ""
	}
	rows, err := delegateListEventOptions(ctx)
	if err != nil {
		logDelegateFailure(ctx, "list_event_options_for_unit", err)
		return ""
	}
	for i := range rows {
		if rows[i].Id == eventID {
			return strings.TrimSpace(rows[i].Unit)
		}
	}
	return ""
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
