package ucg

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

// PushBizPredictImminent 与 ucg PushByBizType 常量对齐。
const PushBizPredictImminent = "predict_imminent"

// PushByBizType 经 ucg internal API 按业务类型发送可见推送。
// 业务：预测临近等；接收方须已在 App 调用 /ucg/app/api/push/register 注册 token。
// Side Effects: 出站 HTTP；失败返回 error（调用方一期应 Ack MQ 不重试）。
func PushByBizType(ctx context.Context, wxID int64, bizType, alert string, data map[string]string) error {
	if wxID <= 0 || strings.TrimSpace(alert) == "" {
		return nil
	}
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("UCG_SERVICE_URL")), "/")
	if base == "" {
		return gerror.NewCode(gcode.CodeInternalError, "未配置 UCG_SERVICE_URL")
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_GATEWAY_INTERNAL_SECRET")
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"wxId":    wxID,
		"bizType": strings.TrimSpace(bizType),
		"alert":   strings.TrimSpace(alert),
		"data":    data,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/ucg/internal/api/push/by-biz-type", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(httpmeta.HeaderDeviceGatewayInternalSecret, secret)
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		glog.Warningf(ctx, "[clients/ucg] push/by-biz-type 失败 wxId=%d err=%v", wxID, err)
		return gerror.WrapCode(gcode.CodeInternalError, err, "ucg 推送失败")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return gerror.NewCode(gcode.CodeInternalError, "ucg 推送失败")
	}
	var env struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return err
	}
	if env.Code != 0 {
		return gerror.NewCode(gcode.CodeInternalError, "ucg 推送失败: "+env.Message)
	}
	return nil
}
