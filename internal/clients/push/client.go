// Package push 出站调用 push-service 的中立 HTTP 客户端。
package push

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"hello/internal/platform/httpmeta"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
)

// 与 push-service 业务类型对齐。
const (
	BizPredictImminent = "predict_imminent"
	BizUcgAlert        = "ucg_alert"
	BizUcgSilentBadge  = "ucg_silent_badge"
)

// PushByBizType 经 push internal API 发送；badge 由调用方提供。
func PushByBizType(ctx context.Context, wxID int64, bizType, alert string, badge int, silent bool, data map[string]string) error {
	if wxID <= 0 {
		return nil
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("PUSH_SERVICE_URL")), "/")
	if base == "" {
		return gerror.NewCode(gcode.CodeInternalError, "未配置 PUSH_SERVICE_URL")
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_GATEWAY_INTERNAL_SECRET")
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"wxId":    wxID,
		"bizType": strings.TrimSpace(bizType),
		"alert":   strings.TrimSpace(alert),
		"badge":   badge,
		"silent":  silent,
		"data":    data,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/push/internal/api/by-biz-type", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(httpmeta.HeaderDeviceGatewayInternalSecret, secret)
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		glog.Warningf(ctx, "[clients/push] by-biz-type 失败 wxId=%d err=%v", wxID, err)
		return gerror.WrapCode(gcode.CodeInternalError, err, "push 发送失败")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return gerror.NewCode(gcode.CodeInternalError, "push 发送失败")
	}
	var env struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return err
	}
	if env.Code != 0 {
		return gerror.NewCode(gcode.CodeInternalError, "push 发送失败: "+env.Message)
	}
	return nil
}

// LiveActivityOp 出站 Live Activity 编排参数（与 push-service internal API 对齐）。
type LiveActivityOp struct {
	WxID           int64
	Action         string // start | update | end
	AttributesKind string // imminent | timing
	ActivityID     string
	Attributes     map[string]interface{}
	ContentState   map[string]interface{}
	AlertTitle     string
	AlertBody      string
	DismissalDate  int64
}

// FanoutLiveActivity 经 push internal API 异步扇出 Live Activity；无 token / 开关关由 push-service 跳过。
func FanoutLiveActivity(ctx context.Context, op LiveActivityOp) error {
	if op.WxID <= 0 {
		return nil
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("PUSH_SERVICE_URL")), "/")
	if base == "" {
		return gerror.NewCode(gcode.CodeInternalError, "未配置 PUSH_SERVICE_URL")
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_GATEWAY_INTERNAL_SECRET")
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"wxId":           op.WxID,
		"action":         strings.TrimSpace(op.Action),
		"attributesKind": strings.TrimSpace(op.AttributesKind),
		"activityId":     strings.TrimSpace(op.ActivityID),
		"attributes":     op.Attributes,
		"contentState":   op.ContentState,
		"alertTitle":     strings.TrimSpace(op.AlertTitle),
		"alertBody":      strings.TrimSpace(op.AlertBody),
		"dismissalDate":  op.DismissalDate,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/push/internal/api/live-activity", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(httpmeta.HeaderDeviceGatewayInternalSecret, secret)
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		glog.Warningf(ctx, "[clients/push] live-activity 失败 wxId=%d err=%v", op.WxID, err)
		return gerror.WrapCode(gcode.CodeInternalError, err, "live-activity 发送失败")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return gerror.NewCode(gcode.CodeInternalError, "live-activity 发送失败")
	}
	var env struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return err
	}
	if env.Code != 0 {
		return gerror.NewCode(gcode.CodeInternalError, "live-activity 发送失败: "+env.Message)
	}
	return nil
}
