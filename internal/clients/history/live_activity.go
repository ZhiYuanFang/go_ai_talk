package history

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

// RefreshBabyLiveActivity 经 HISTORY_SERVICE_URL 触发统一 Live Activity Refresh。
//
// 业务：voice pending PUT / MQ 到点调用；history 本进程应直接调 services/history.RefreshBabyLiveActivity。
// Side Effects: 出站 HTTP；失败仅返回错误，调用方通常 WARN 不阻断主路径。
func RefreshBabyLiveActivity(ctx context.Context, deviceNo string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("HISTORY_SERVICE_URL")), "/")
	if base == "" {
		return gerror.NewCode(gcode.CodeInternalError, "未配置 HISTORY_SERVICE_URL")
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_GATEWAY_INTERNAL_SECRET")
	}
	payload, _ := json.Marshal(map[string]string{"deviceNo": deviceNo})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/history/internal/api/live-activity/refresh", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(httpmeta.HeaderDeviceGatewayInternalSecret, secret)

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		glog.Warningf(ctx, "[clients/history] live-activity refresh 失败 deviceNoLen=%d err=%v", len(deviceNo), err)
		return gerror.WrapCode(gcode.CodeInternalError, err, "history live-activity refresh 失败")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		glog.Warningf(ctx, "[clients/history] live-activity refresh HTTP=%d body=%s", resp.StatusCode, string(body))
		return gerror.NewCode(gcode.CodeInternalError, "history live-activity refresh 失败")
	}
	var env struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "history 响应解析失败")
	}
	if env.Code != 0 {
		return gerror.NewCode(gcode.CodeInternalError, "history live-activity refresh 失败: "+env.Message)
	}
	return nil
}
