package device

// xiaozhi_mcp.go：出站拉取 device-service 小智绑定全量（供 xiaozhi-mcp-service 启动/reconcile）。

import (
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

// XiaozhiMcpBindingItem 内部全量绑定项（与 device 域字段对齐）。
type XiaozhiMcpBindingItem struct {
	Id       int64  `json:"id"`
	WxId     int64  `json:"wxId"`
	DeviceNo string `json:"deviceNo"`
	McpToken string `json:"mcpToken"`
	Alias    string `json:"alias"`
	Status   int    `json:"status"`
}

// ListActiveXiaozhiMcpBindings 经 DEVICE_SERVICE_URL 拉取全部 active 小智绑定（含完整 token）。
// Side Effects: 出站 HTTP；须配置 DEVICE_GATEWAY_INTERNAL_SECRET。
func ListActiveXiaozhiMcpBindings(ctx context.Context) ([]XiaozhiMcpBindingItem, error) {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("DEVICE_SERVICE_URL")), "/")
	if base == "" {
		return nil, gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_SERVICE_URL")
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return nil, gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_GATEWAY_INTERNAL_SECRET")
	}
	url := base + "/device/internal/api/xiaozhi-mcp/bindings"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(httpmeta.HeaderDeviceGatewayInternalSecret, secret)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		glog.Warningf(ctx, "[clients/device] list xiaozhi bindings 失败 err=%v", err)
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "拉取小智绑定失败")
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, gerror.NewCode(gcode.CodeInternalError, "拉取小智绑定失败")
	}
	var env struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			List []XiaozhiMcpBindingItem `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, err
	}
	if env.Code != 0 {
		return nil, gerror.NewCode(gcode.CodeInternalError, "拉取小智绑定失败: "+env.Message)
	}
	return env.Data.List, nil
}
