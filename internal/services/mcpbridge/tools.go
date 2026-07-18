package mcpbridge

import (
	"context"
	"fmt"
	histsvc "hello/internal/services/history"
	"strings"

	"github.com/gogf/gf/v2/os/glog"
)

// serverName / serverVersion 暴露给小智平台的 MCP 服务端标识。
const (
	serverName    = "xiaozhi-mcp-service"
	serverVersion = "1.0.0"
)

// chatToolName 工具名；小智平台在 tools/list 中按此名称调用。
const chatToolName = "chat"

// chatToolDescription 工具描述，影响大模型是否选用本工具。
const chatToolDescription = "与设备进行文本对话：接收用户输入的文本，返回 AI 回复内容。"

// chatToolInputSchema 工具入参 schema，遵循 JSON Schema 子集。
// 仅暴露 transcript 一个必填 string 参数。
func chatToolInputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"transcript": map[string]any{
				"type":        "string",
				"description": "用户输入的文本内容",
			},
			"xzDeviceNo": map[string]any{
				"type":        "string",
				"description": "设备号",
			},
		},
		"required": []string{"transcript"},
	}
}

// ChatHandler 封装 chat 工具的执行逻辑。
// deviceNo 在构造时从环境变量注入（单设备绑定场景）。
type ChatHandler struct {
	deviceNo string
}

// NewChatHandler 构造 chat 工具 handler。
// deviceNo 不能为空，由调用方（main.go）在启动阶段校验后传入。
func NewChatHandler(deviceNo string) *ChatHandler {
	return &ChatHandler{deviceNo: deviceNo}
}

// Handle 执行 chat 工具调用。
// 流程：
//  1. 从 arguments 取 transcript 并 trim；
//  2. 空串直接返回 tool error，不发起下游调用；
//  3. 调用 histsvc.DelegateTextChat（进程内函数，内部经 HTTP 委派 voice-service）；
//  4. 成功返回 reply 文本；失败返回 tool error 并记录日志。
//
// wxID 固定传 0：MCP 接入点绑定单设备，不参与 wx 维度 AI 额度校验。
func (h *ChatHandler) Handle(ctx context.Context, arguments map[string]any) *ToolsCallResult {
	transcript := readTranscriptArg(arguments)
	if transcript == "" {
		// 参数校验失败：不发起下游调用，直接返回 tool error。
		return NewErrorCallResult("transcript 不能为空")
	}
	glog.Infof(ctx, "小智设备号：%s", arguments["xzDeviceNo"])
	reply, err := histsvc.DelegateTextChat(ctx, h.deviceNo, transcript, 0)
	if err != nil {
		// 下游失败：记录原始错误并返回 tool error，避免暴露内部细节给小智。
		glog.Errorf(ctx, "[mcp-bridge] chat failed deviceNo=%s err=%v", h.deviceNo, err)
		return NewErrorCallResult(fmt.Sprintf("对话失败：%v", err))
	}
	return NewTextCallResult(reply)
}

// readTranscriptArg 从 arguments 提取 transcript 字符串并 trim。
// 兼容 number / bool 等类型被错误传入的场景，统一转字符串后 trim。
func readTranscriptArg(arguments map[string]any) string {
	if arguments == nil {
		return ""
	}
	raw, ok := arguments["transcript"]
	if !ok || raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		// 非字符串类型尝试 fmt 转字符串后 trim，避免 panic。
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

// ListTools 返回本服务暴露的工具列表（用于 tools/list 响应）。
// 当前仅 chat 一个工具。
func ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        chatToolName,
			Description: chatToolDescription,
			InputSchema: chatToolInputSchema(),
		},
	}
}

// ServerName / ServerVersion 暴露给 bridge 构造 initialize 响应。
func ServerName() string    { return serverName }
func ServerVersion() string { return serverVersion }
