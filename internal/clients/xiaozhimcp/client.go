// Package xiaozhimcp 出站调用 xiaozhi-mcp-service 内部 HTTP 的中立客户端。
//
// 业务：device-service 在小智绑定增删成功后通知 Manager Upsert/Remove。
// 调用方 import clients，禁止 import services/mcpbridge。
package xiaozhimcp

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

// BindingUpsert 通知 mcp Manager 启动或更新一条 Bridge。
// Args: id 绑定主键；token 小智 MCP token；deviceNo 喂养落点。
// Returns: 网络/鉴权/业务错误；调用方失败时不应回滚 DB。
func BindingUpsert(ctx context.Context, id int64, token, deviceNo string) error {
	token = strings.TrimSpace(token)
	deviceNo = strings.TrimSpace(deviceNo)
	if token == "" || deviceNo == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "token/deviceNo 不能为空")
	}
	return postJSON(ctx, "/xiaozhi-mcp/internal/api/bindings/upsert", map[string]interface{}{
		"id":       id,
		"mcpToken": token,
		"deviceNo": deviceNo,
	})
}

// BindingRemove 通知 mcp Manager 停止对应 Bridge。
func BindingRemove(ctx context.Context, id int64, token string) error {
	token = strings.TrimSpace(token)
	if token == "" && id <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "id 或 token 至少提供一个")
	}
	return postJSON(ctx, "/xiaozhi-mcp/internal/api/bindings/remove", map[string]interface{}{
		"id":       id,
		"mcpToken": token,
	})
}

func postJSON(ctx context.Context, path string, body map[string]interface{}) error {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("XIAOZHI_MCP_SERVICE_URL")), "/")
	if base == "" {
		return gerror.NewCode(gcode.CodeInternalError, "未配置 XIAOZHI_MCP_SERVICE_URL")
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_GATEWAY_INTERNAL_SECRET")
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(httpmeta.HeaderDeviceGatewayInternalSecret, secret)
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		glog.Warningf(ctx, "[clients/xiaozhimcp] %s 失败 err=%v", path, err)
		return gerror.WrapCode(gcode.CodeInternalError, err, "通知 xiaozhi-mcp 失败")
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		glog.Warningf(ctx, "[clients/xiaozhimcp] %s status=%d body=%s", path, resp.StatusCode, string(raw))
		return gerror.NewCode(gcode.CodeInternalError, "通知 xiaozhi-mcp 失败")
	}
	var env struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return err
	}
	if env.Code != 0 {
		return gerror.NewCode(gcode.CodeInternalError, "通知 xiaozhi-mcp 失败: "+env.Message)
	}
	return nil
}
