package mcpbridge

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gogf/gf/v2/os/glog"
)

// serverName / serverVersion 暴露给小智平台的 MCP 服务端标识。
const (
	serverName    = "xiaozhi-mcp-service"
	serverVersion = "1.0.0"
)

// chatToolName 工具名；小智平台在 tools/list 中按此名称调用。
const chatToolName = "baby_feeding_advisor"

// chatToolDescription 工具描述，影响大模型是否选用本工具。
const chatToolDescription = "" +
	"宝宝喂养与成长顾问。当用户提到以下内容时，必须调用此工具：\n" +
	"1. 记录喂养信息：奶量、母乳/奶粉/辅食、喂养时间\n" +
	"2. 查询喂养历史：昨天喂了多少次、最近奶量趋势\n" +
	"3. 喂养建议：宝宝不爱喝奶怎么办、该加辅食了吗\n" +
	"4. 成长发育：身高体重是否达标、睡眠建议\n" +
	"5. 健康提醒：吐奶、胀气、过敏等常见问题\n" +
	"\n" +
	"调用时传入用户的原话，工具会自动理解并执行记录或查询，返回专业建议。" +
	"注意：只要用户提到宝宝、喂养、奶量、辅食、成长、睡眠、健康等相关话题，都必须调用此工具，不要自己回答。"

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
// 流程：校验 transcript → 经 /voice/chat/ws 文模式对话 → 返回 answer。
func (h *ChatHandler) Handle(ctx context.Context, arguments map[string]any) *ToolsCallResult {
	transcript := readTranscriptArg(arguments)
	if transcript == "" {
		return NewErrorCallResult("transcript 不能为空")
	}
	glog.Infof(ctx, "小智设备号：%s", arguments["xzDeviceNo"])
	reply, err := ChatViaVoiceWS(ctx, h.deviceNo, transcript)
	if err != nil {
		wsURL := os.Getenv("VOICE_CHAT_WS_URL")
		if strings.TrimSpace(wsURL) == "" {
			wsURL = os.Getenv("VOICE_SERVICE_URL")
		}
		glog.Errorf(ctx, "[mcp-bridge] chat WS failed deviceNo=%s voiceURL=%s err=%v",
			h.deviceNo, wsURL, err)
		return NewErrorCallResult(fmt.Sprintf("对话失败：%v", err))
	}
	return NewTextCallResult(reply)
}

// readTranscriptArg 从 arguments 提取 transcript 字符串并 trim。
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
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

// ListTools 返回本服务暴露的工具列表（用于 tools/list 响应）。
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
