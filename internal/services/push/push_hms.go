package push

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// HmsSender 经华为 Push Kit REST 下发通知/静默消息。
type HmsSender struct{}

func NewHmsSender() *HmsSender { return &HmsSender{} }

func (s *HmsSender) Channel() string { return PushChannelHMS }

const (
	// hmsBizCodeOK 华为下行消息业务成功码（字符串形式）。
	hmsBizCodeOK = "80000000"
	// envHmsClickIntent 非空时 click_action 使用 type=1 + intent（深链）；否则 type=3 打开应用。
	envHmsClickIntent = "PUSH_HMS_CLICK_INTENT"
	hmsDefaultTitle   = "胖宝"
	hmsBadgeClass     = "com.fzy.pangbao.MainActivity"
)

var (
	hmsTokenMu    sync.Mutex
	hmsTokenCache string
	hmsTokenExp   time.Time
	// hmsBizTypeToken 只允许业务类型常量，避免拼进 intent 时注入分号。
	hmsBizTypeToken = regexp.MustCompile(`^[a-z0-9_]+$`)
)

// Send 向单个 HMS token 发送推送。
//
// 业务：非静默+有 alert 走「通知消息」形态（顶层 message.notification + android.click_action），
// 对齐华为控制台常见样例，避免仅 android 内嵌 title 导致托盘不展示。
// 静默/无 alert：不伪造可见正文，仅 badge + data。
//
// Returns: invalidToken 表示应删除本地 token；err 非 nil 时 dispatcher 记 send_failed。
func (s *HmsSender) Send(ctx context.Context, token string, payload PushPayload) (invalidToken bool, err error) {
	cfg := loadPushConfig(ctx)
	if !hmsConfigured(cfg) {
		g.Log().Warningf(ctx, "[push] HMS skipped: credentials not configured")
		return false, fmt.Errorf("credentials not configured")
	}
	accessToken, err := hmsAccessToken(ctx, cfg)
	if err != nil {
		return false, err
	}
	dataStr, err := buildHmsMessage(payload)
	if err != nil {
		return false, err
	}

	// 华为：若带 add_num 则必须在 1–99；predict_imminent 等场景 badge=0，禁止写 add_num:0（会 80100003）。
	// badge<=0：整段省略角标，仅发通知正文；badge>0：只传 set_num+class，不传 add_num。
	// androidNotif := map[string]interface{}{
	// 	"click_action": hmsClickAction(),
	// }
	bizType := ""
	if payload.Data != nil {
		bizType = strings.TrimSpace(payload.Data["bizType"])
	}
	androidNotif := map[string]interface{}{
		"click_action": hmsClickAction(bizType),
		"priority":     "HIGH",
		"importance":   "HIGH",
		// "channelId":    "push_default", // 和Flutter端创建的通知渠道ID保持一致！ 部分老机子不支持
	}
	if payload.Badge > 0 {
		androidNotif["badge"] = map[string]interface{}{
			"class":   hmsBadgeClass,
			"set_num": payload.Badge,
		}
	}
	android := map[string]interface{}{
		"notification": androidNotif,
		"data":         dataStr,
	}
	message := map[string]interface{}{
		"token":   []string{strings.TrimSpace(token)},
		"android": android,
	}

	// 可见通知：必须有顶层 notification，系统托盘才按「通知消息」展示。
	visible := !payload.Silent && strings.TrimSpace(payload.Alert) != ""
	if visible {
		message["notification"] = map[string]interface{}{
			"title": hmsDefaultTitle,
			"body":  strings.TrimSpace(payload.Alert),
		}
	}
	// 静默/无 alert：不设置顶层 notification，避免空正文骚扰；仅依赖 android badge/data。

	reqBody := map[string]interface{}{"message": message}
	raw, _ := json.Marshal(reqBody)
	endpoint := fmt.Sprintf("https://push-api.cloud.huawei.com/v1/%s/messages:send", url.PathEscape(cfg.HmsAppID))
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		invalid, parseErr := parseHmsInvalidToken(body)
		if invalid {
			return true, parseErr
		}
		return false, fmt.Errorf("hms status=%d body=%s", resp.StatusCode, truncateHmsLog(string(body), 512))
	}

	// HTTP 2xx 仍须校验业务 code；否则会出现假 send_ok。
	code, msg := parseHmsResponseMeta(body)
	if code != hmsBizCodeOK {
		invalid, parseErr := parseHmsInvalidToken(body)
		if invalid {
			return true, parseErr
		}
		if msg == "" {
			msg = truncateHmsLog(string(body), 256)
		}
		return false, fmt.Errorf("hms biz code=%s msg=%s", code, truncateHmsLog(msg, 256))
	}
	return false, nil
}

// hmsClickAction 有 bizType 时用 type=1，把字段放进点击 Intent extras。
// type=3 只打开应用，china_push 读不到业务字段。无 bizType 时仍 type=3（或环境变量里的自定义 intent）。
func hmsClickAction(bizType string) map[string]interface{} {
	bizType = strings.TrimSpace(bizType)
	if intent := hmsBizTypeClickIntent(bizType); intent != "" {
		return map[string]interface{}{
			"type":   1,
			"intent": intent,
		}
	}
	if intent := strings.TrimSpace(os.Getenv(envHmsClickIntent)); intent != "" {
		return map[string]interface{}{
			"type":   1,
			"intent": intent,
		}
	}
	return map[string]interface{}{"type": 3}
}

// hmsBizTypeClickIntent 生成带 S.bizType 的显式打开 MainActivity 的 intent。
// scheme/host/path 必须与 AndroidManifest 的 pangbaopush 过滤器一致。
func hmsBizTypeClickIntent(bizType string) string {
	if bizType == "" || !hmsBizTypeToken.MatchString(bizType) {
		return ""
	}
	return "intent://click/open#Intent;scheme=pangbaopush;launchFlags=0x14000000;package=com.fzy.pangbao;component=com.fzy.pangbao/com.fzy.pangbao.MainActivity;S.bizType=" + bizType + ";end"
}

func buildHmsMessage(payload PushPayload) (string, error) {
	m := map[string]interface{}{
		"badge":  fmt.Sprintf("%d", payload.Badge),
		"silent": payload.Silent,
	}
	if payload.Alert != "" {
		m["alert"] = payload.Alert
	}
	// 透传给端上的业务 data（如 bizType），由调用方 payload.Data 扩展时在此合并更清晰；
	// 当前 build 保持轻量，biz 字段已在 APNs 等路径由 data map 承载；HMS android.data 至少含 badge/silent。
	if payload.Data != nil {
		for k, v := range payload.Data {
			k = strings.TrimSpace(k)
			if k == "" {
				continue
			}
			// 不覆盖已有核心键。
			if _, exists := m[k]; exists {
				continue
			}
			m[k] = v
		}
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

// parseHmsResponseMeta 提取华为下行响应 code/msg（code 统一成字符串便于比较）。
func parseHmsResponseMeta(body []byte) (code, msg string) {
	var m map[string]interface{}
	if json.Unmarshal(body, &m) != nil {
		return "", ""
	}
	code = strings.TrimSpace(fmt.Sprint(m["code"]))
	if code == "<nil>" {
		code = ""
	}
	msg = strings.TrimSpace(fmt.Sprint(m["msg"]))
	if msg == "<nil>" {
		msg = ""
	}
	return code, msg
}

func parseHmsInvalidToken(body []byte) (bool, error) {
	code, msg := parseHmsResponseMeta(body)
	upper := strings.ToUpper(code + " " + msg)
	// 常见：token 无效/未注册；部分错误码文档写作 80300007 等，文案含 TOKEN。
	if strings.Contains(upper, "INVALID") && strings.Contains(upper, "TOKEN") {
		return true, fmt.Errorf("hms invalid token: code=%s msg=%s", code, truncateHmsLog(msg, 128))
	}
	if strings.Contains(upper, "NOT_REGISTERED") {
		return true, fmt.Errorf("hms not registered: code=%s msg=%s", code, truncateHmsLog(msg, 128))
	}
	// 80300007：部分文档标明为 token 无效类。
	if code == "80300007" {
		return true, fmt.Errorf("hms invalid token: code=%s msg=%s", code, truncateHmsLog(msg, 128))
	}
	return false, nil
}

func truncateHmsLog(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
