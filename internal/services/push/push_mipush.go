package push

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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
		g.Log().Warningf(ctx, "[push] MiPush skipped: credentials not configured")
		return false, fmt.Errorf("credentials not configured")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return false, nil
	}
	form := url.Values{}
	form.Set("registration_id", token)
	form.Set("restricted_package_name", firstNonEmpty(os.Getenv("PUSH_MIPUSH_PACKAGE_NAME"), cfgStr(ctx, "push.mipush.packageName")))
	if form.Get("restricted_package_name") == "" {
		form.Set("restricted_package_name", "com.fzy.pangbao")
	}
	form.Set("notify_type", "-1")
	form.Set("extra.badge", fmt.Sprintf("%d", payload.Badge))
	// 点击回调只读 payload（getContent）。可见消息也写入，便于以后小米上架后解析 bizType。
	form.Set("payload", mipushClickPayload(payload))
	if payload.Silent {
		form.Set("pass_through", "1")
	} else {
		form.Set("pass_through", "0")
		// 1 = 打开 launcher；本变更不验收小米点击。
		form.Set("extra.notify_effect", "1")
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

// mipushClickPayload 把 bizType 等 data 放进小米 payload 字符串。
func mipushClickPayload(payload PushPayload) string {
	m := map[string]string{
		"badge": fmt.Sprintf("%d", payload.Badge),
	}
	if payload.Silent {
		m["silent"] = "true"
	}
	for k, v := range payload.Data {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		m[k] = v
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(b)
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
