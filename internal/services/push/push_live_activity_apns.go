package push

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// LiveActivityAPNsPayload ActivityKit 远程推送载荷（start/update/end）。
type LiveActivityAPNsPayload struct {
	Event          string                 // start | update | end
	AttributesType string                 // start 必填：ActivityAttributes 类型名
	Attributes     map[string]interface{} // start 必填
	ContentState   map[string]interface{} // 各事件 content-state
	AlertTitle     string
	AlertBody      string
	// DismissalDate 为 APNs dismissal-date（unix 秒）。end 时 <=0 由组装层填「现在-1 秒」（过去时刻=立即从锁屏摘条）；
	// 禁止省略该键（Apple 省略后 ended 活动可在锁屏再挂数小时）。>0 则原样写入。
	DismissalDate int64
}

// SendLiveActivityAPNs 使用现有 .p8 JWT 向 APNs 发送 liveactivity。
//
// 业务：apns-push-type=liveactivity；topic={bundleId}.push-type.liveactivity（后缀代码拼接，勿写入 env Bundle ID）。
// Returns: invalidToken 表示应删除该 LA token 行。
func SendLiveActivityAPNs(ctx context.Context, deviceToken string, payload LiveActivityAPNsPayload) (invalidToken bool, err error) {
	cfg := loadPushConfig(ctx)
	if !apnsConfigured(cfg) {
		g.Log().Warningf(ctx, "[push-la] APNs skipped: credentials not configured")
		return false, fmt.Errorf("credentials not configured")
	}
	deviceToken = strings.TrimSpace(deviceToken)
	if deviceToken == "" {
		return false, nil
	}
	event := strings.TrimSpace(strings.ToLower(payload.Event))
	if event != "start" && event != "update" && event != "end" {
		return false, fmt.Errorf("invalid liveactivity event")
	}
	body, err := buildLiveActivityAPNsBody(payload)
	if err != nil {
		return false, err
	}
	host := "https://api.sandbox.push.apple.com"
	if cfg.ApnsProduction {
		host = "https://api.push.apple.com"
	}
	url := host + "/3/device/" + deviceToken
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	bearer, err := apnsBearerToken(cfg)
	if err != nil {
		return false, err
	}
	topic := strings.TrimSpace(cfg.ApnsBundleID) + ".push-type.liveactivity"
	req.Header.Set("authorization", "bearer "+bearer)
	req.Header.Set("apns-topic", topic)
	req.Header.Set("apns-push-type", "liveactivity")
	req.Header.Set("apns-priority", "10")
	req.Header.Set("content-type", "application/json")
	client := apnsHTTPClientFor(cfg)
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode == http.StatusOK {
		return false, nil
	}
	reason := parseApnsReason(respBody)
	if isApnsInvalidToken(resp.StatusCode, reason) {
		return true, fmt.Errorf("apns invalid token: %s", reason)
	}
	return false, fmt.Errorf("apns status=%d reason=%s body=%s", resp.StatusCode, reason, string(respBody))
}

func buildLiveActivityAPNsBody(payload LiveActivityAPNsPayload) ([]byte, error) {
	event := strings.TrimSpace(strings.ToLower(payload.Event))
	aps := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"event":     event,
	}
	cs := payload.ContentState
	if cs == nil {
		cs = map[string]interface{}{}
	}
	aps["content-state"] = cs
	if event == "start" {
		aps["attributes-type"] = strings.TrimSpace(payload.AttributesType)
		attrs := payload.Attributes
		if attrs == nil {
			attrs = map[string]interface{}{}
		}
		aps["attributes"] = attrs
	}
	if event == "end" {
		// Apple：过去的 dismissal-date 才会立刻从锁屏移除；省略则可能再挂数小时。
		dismissAt := payload.DismissalDate
		if dismissAt <= 0 {
			dismissAt = time.Now().Unix() - 1
		}
		aps["dismissal-date"] = dismissAt
	}
	title := strings.TrimSpace(payload.AlertTitle)
	body := strings.TrimSpace(payload.AlertBody)
	if title != "" || body != "" {
		alert := map[string]string{}
		if title != "" {
			alert["title"] = title
		}
		if body != "" {
			alert["body"] = body
		}
		aps["alert"] = alert
	}
	root := map[string]interface{}{"aps": aps}
	return json.Marshal(root)
}
