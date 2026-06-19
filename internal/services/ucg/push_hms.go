package ucg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// HmsSender sends via Huawei Push Kit REST API.
type HmsSender struct{}

func NewHmsSender() *HmsSender { return &HmsSender{} }

func (s *HmsSender) Channel() string { return PushChannelHMS }

var (
	hmsTokenMu    sync.Mutex
	hmsTokenCache string
	hmsTokenExp   time.Time
)

func (s *HmsSender) Send(ctx context.Context, token string, payload PushPayload) (invalidToken bool, err error) {
	cfg := loadPushConfig(ctx)
	if !hmsConfigured(cfg) {
		g.Log().Debug(ctx, "[ucg-push] HMS skipped: credentials not configured")
		return false, nil
	}
	accessToken, err := hmsAccessToken(ctx, cfg)
	if err != nil {
		return false, err
	}
	msgBody, err := buildHmsMessage(payload)
	if err != nil {
		return false, err
	}
	endpoint := fmt.Sprintf("https://push-api.cloud.huawei.com/v1/%s/messages:send", url.PathEscape(cfg.HmsAppID))
	reqBody := map[string]interface{}{
		"message": map[string]interface{}{
			"token": []string{strings.TrimSpace(token)},
			"android": map[string]interface{}{
				"notification": map[string]interface{}{},
				"data":         msgBody,
			},
		},
	}
	if !payload.Silent && strings.TrimSpace(payload.Alert) != "" {
		reqBody["message"].(map[string]interface{})["android"].(map[string]interface{})["notification"] = map[string]interface{}{
			"title": "胖宝",
			"body":  payload.Alert,
			"click_action": map[string]interface{}{
				"type": 1,
			},
			"badge": map[string]interface{}{
				"add_num": 0,
				"class":   "com.fzy.pangbao.MainActivity",
				"set_num": payload.Badge,
			},
		}
	} else {
		reqBody["message"].(map[string]interface{})["android"].(map[string]interface{})["notification"] = map[string]interface{}{
			"badge": map[string]interface{}{
				"add_num": 0,
				"class":   "com.fzy.pangbao.MainActivity",
				"set_num": payload.Badge,
			},
		}
	}
	raw, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(raw)))
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return false, nil
	}
	invalid, parseErr := parseHmsInvalidToken(body)
	if invalid {
		return true, parseErr
	}
	return false, fmt.Errorf("hms status=%d body=%s", resp.StatusCode, string(body))
}

func buildHmsMessage(payload PushPayload) (string, error) {
	m := map[string]interface{}{
		"badge": fmt.Sprintf("%d", payload.Badge),
		"silent": payload.Silent,
	}
	if payload.Alert != "" {
		m["alert"] = payload.Alert
	}
	b, err := json.Marshal(m)
	return string(b), err
}

func hmsAccessToken(ctx context.Context, cfg pushConfig) (string, error) {
	hmsTokenMu.Lock()
	defer hmsTokenMu.Unlock()
	if hmsTokenCache != "" && time.Now().Before(hmsTokenExp.Add(-60*time.Second)) {
		return hmsTokenCache, nil
	}
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", cfg.HmsAppID)
	form.Set("client_secret", cfg.HmsAppSecret)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://oauth-login.cloud.huawei.com/oauth2/v3/token",
		strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var parsed struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
	}
	if json.Unmarshal(body, &parsed) != nil || parsed.AccessToken == "" {
		return "", fmt.Errorf("hms oauth failed: %s", string(body))
	}
	hmsTokenCache = parsed.AccessToken
	expSec := parsed.ExpiresIn
	if expSec <= 0 {
		expSec = 3600
	}
	hmsTokenExp = time.Now().Add(time.Duration(expSec) * time.Second)
	return hmsTokenCache, nil
}

func parseHmsInvalidToken(body []byte) (bool, error) {
	var m map[string]interface{}
	if json.Unmarshal(body, &m) != nil {
		return false, nil
	}
	code := fmt.Sprint(m["code"])
	msg := fmt.Sprint(m["msg"])
	upper := strings.ToUpper(code + " " + msg)
	if strings.Contains(upper, "INVALID") && strings.Contains(upper, "TOKEN") {
		return true, fmt.Errorf("hms invalid token: %s", msg)
	}
	if strings.Contains(upper, "NOT_REGISTERED") {
		return true, fmt.Errorf("hms not registered: %s", msg)
	}
	return false, nil
}
