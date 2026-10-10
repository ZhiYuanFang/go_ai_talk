// Package xiaozhimcp 出站调用 xiaozhi-mcp-service 内部 HTTP 的中立客户端。
//
// 业务：device-service 在小智绑定增删成功后通知 Manager Upsert/Remove；
// 断连时 BindingReconnect 强制 cancel+redial；列表时批量查询 token 连接态。
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
// Args: id 绑定主键；token；deviceNo 喂养落点；wxId 开通主体（tools 校验用）。
// Returns: 网络/鉴权/业务错误；调用方失败时不应回滚 DB。
func BindingUpsert(ctx context.Context, id int64, token, deviceNo string, wxId int64) error {
	token = strings.TrimSpace(token)
	deviceNo = strings.TrimSpace(deviceNo)
	if token == "" || deviceNo == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "token/deviceNo 不能为空")
	}
	if wxId <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	_, err := postJSON(ctx, "/xiaozhi-mcp/internal/api/bindings/upsert", map[string]interface{}{
		"id":       id,
		"mcpToken": token,
		"deviceNo": deviceNo,
		"wxId":     wxId,
	})
	return err
}

// BindingReconnect 通知 mcp Manager 强制重连（cancel + startLocked，即使 key 未变）。
// 业务：App/Hub「重连」；失败语义与 Upsert/Remove 一致，调用方不应改 DB。
//
// Args: id 绑定主键；token；deviceNo；wxId 开通主体。
// Returns: 参数/网络/鉴权/业务错误。
func BindingReconnect(ctx context.Context, id int64, token, deviceNo string, wxId int64) error {
	token = strings.TrimSpace(token)
	deviceNo = strings.TrimSpace(deviceNo)
	if token == "" || deviceNo == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "token/deviceNo 不能为空")
	}
	if wxId <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	_, err := postJSON(ctx, "/xiaozhi-mcp/internal/api/bindings/reconnect", map[string]interface{}{
		"id":       id,
		"mcpToken": token,
		"deviceNo": deviceNo,
		"wxId":     wxId,
	})
	return err
}

// BindingRemove 通知 mcp Manager 停止对应 Bridge。
func BindingRemove(ctx context.Context, id int64, token string) error {
	token = strings.TrimSpace(token)
	if token == "" && id <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "id 或 token 至少提供一个")
	}
	_, err := postJSON(ctx, "/xiaozhi-mcp/internal/api/bindings/remove", map[string]interface{}{
		"id":       id,
		"mcpToken": token,
	})
	return err
}

// ConnectionStatus 批量查询 token 是否已连通小智 MCP WebSocket。
// Args: tokens 规范化后的 mcp token 列表。
// Returns: token→connected；网络/鉴权错误（调用方应降级为全 false）。
func ConnectionStatus(ctx context.Context, tokens []string) (map[string]bool, error) {
	cleaned := make([]string, 0, len(tokens))
	for _, t := range tokens {
		t = strings.TrimSpace(t)
		if t != "" {
			cleaned = append(cleaned, t)
		}
	}
	if len(cleaned) == 0 {
		return map[string]bool{}, nil
	}
	raw, err := postJSON(ctx, "/xiaozhi-mcp/internal/api/bindings/connection-status", map[string]interface{}{
		"tokens": cleaned,
	})
	if err != nil {
		return nil, err
	}
	var env struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Statuses map[string]bool `json:"statuses"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, err
	}
	if env.Data.Statuses == nil {
		return map[string]bool{}, nil
	}
	return env.Data.Statuses, nil
}

// postJSON 发送内部 POST；成功时返回响应 body（含 code=0）。
func postJSON(ctx context.Context, path string, body map[string]interface{}) ([]byte, error) {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("XIAOZHI_MCP_SERVICE_URL")), "/")
	if base == "" {
		return nil, gerror.NewCode(gcode.CodeInternalError, "未配置 XIAOZHI_MCP_SERVICE_URL")
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return nil, gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_GATEWAY_INTERNAL_SECRET")
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+path, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(httpmeta.HeaderDeviceGatewayInternalSecret, secret)
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		glog.Warningf(ctx, "[clients/xiaozhimcp] %s 失败 err=%v", path, err)
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "通知 xiaozhi-mcp 失败")
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		glog.Warningf(ctx, "[clients/xiaozhimcp] %s status=%d body=%s", path, resp.StatusCode, string(raw))
		return nil, gerror.NewCode(gcode.CodeInternalError, "通知 xiaozhi-mcp 失败")
	}
	var env struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, err
	}
	if env.Code != 0 {
		return nil, gerror.NewCode(gcode.CodeInternalError, "通知 xiaozhi-mcp 失败: "+env.Message)
	}
	return raw, nil
}
