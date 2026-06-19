package ucg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// MipushSender sends via Xiaomi Push REST API.
type MipushSender struct{}

func NewMipushSender() *MipushSender { return &MipushSender{} }

func (s *MipushSender) Channel() string { return PushChannelMiPush }

func (s *MipushSender) Send(ctx context.Context, token string, payload PushPayload) (invalidToken bool, err error) {
	cfg := loadPushConfig(ctx)
	if !mipushConfigured(cfg) {
		g.Log().Debug(ctx, "[ucg-push] MiPush skipped: credentials not configured")
		return false, nil
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return false, nil
	}
	form := url.Values{}
	form.Set("registration_id", token)
	form.Set("restricted_package_name", cfgStr(ctx, "ucg.push.mipush.packageName"))
	if form.Get("restricted_package_name") == "" {
		form.Set("restricted_package_name", "com.fzy.pangbao")
	}
	form.Set("notify_type", "-1")
	form.Set("extra.badge", fmt.Sprintf("%d", payload.Badge))
	if payload.Silent {
		form.Set("pass_through", "1")
		form.Set("payload", fmt.Sprintf(`{"badge":%d,"silent":true}`, payload.Badge))
	} else {
		form.Set("pass_through", "0")
		title := "胖宝"
		body := strings.TrimSpace(payload.Alert)
		if body == "" {
			body = "您有一条新消息"
		}
		form.Set("title", title)
		form.Set("description", body)
		form.Set("notify_id", fmt.Sprintf("%d", payload.Badge))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.xmpush.xiaomi.com/v3/message/regid",
		strings.NewReader(form.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "key="+cfg.MipushAppSecret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		invalid, invErr := parseMipushInvalidToken(respBody)
		return invalid, invErr
	}
	return false, fmt.Errorf("mipush status=%d body=%s", resp.StatusCode, string(respBody))
}

func parseMipushInvalidToken(body []byte) (bool, error) {
	var m struct {
		Code        int    `json:"code"`
		Description string `json:"description"`
		Reason      string `json:"reason"`
	}
	if json.Unmarshal(body, &m) != nil {
		return false, nil
	}
	text := strings.ToUpper(fmt.Sprintf("%d %s %s", m.Code, m.Description, m.Reason))
	if m.Code == 20301 || strings.Contains(text, "INVALID") && strings.Contains(text, "REG") {
		return true, fmt.Errorf("mipush invalid token: %s", m.Description)
	}
	if m.Code != 0 {
		return false, fmt.Errorf("mipush error code=%d desc=%s", m.Code, m.Description)
	}
	return false, nil
}
