package mcpbridge

import (
	"context"
	"fmt"
	"os"
	"strings"

	cashclient "hello/internal/clients/cash"

	"github.com/gogf/gf/v2/os/glog"
)

// serverName / serverVersion 暴露给小智平台的 MCP 服务端标识。
const (
	serverName    = "xiaozhi-mcp-service"
	serverVersion = "1.0.0"
)

// chatToolName 工具名；小智平台在 tools/list 中按此名称调用。
const chatToolName = "baby_feeding_advisor"

// chatToolDescription 工具描述，供小智选路：谈到宝宝时调用本工具，不限于喂养记录。
const chatToolDescription = "" +
	"宝宝顾问。用户谈到宝宝时调用此工具，不限于喂养记录。包括：\n" +
	"1. 喂养：记录或查询奶量、母乳/奶粉/辅食、时间与历史\n" +
	"2. 宝宝信息：名字、生日、月龄、性别、身高体重、睡眠、发育与日常照护\n" +
	"3. 对上一条本工具操作的确认或取消（是的、确定、好的、取消、不要、算了等），原话放入 transcript\n" +
	"\n" +
	"调用时传入用户原话，由工具理解并回答。"

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
type ChatHandler struct {
	deviceNo             string
	wxId                 int64
	onEntitlementDenied  func() // 明确未开通时懒停桥
}

// NewChatHandler 构造 chat 工具 handler。
func NewChatHandler(deviceNo string, wxId int64) *ChatHandler {
	return &ChatHandler{deviceNo: deviceNo, wxId: wxId}
}

// Handle 执行 chat 工具调用。
// 流程：校验开通 → transcript → /voice/chat/ws 文模式 → answer。
func (h *ChatHandler) Handle(ctx context.Context, arguments map[string]any) *ToolsCallResult {
	// 硬闸门：wx 须有效权益；缺 wxId 或未开通 fail-closed。
	if h.wxId <= 0 {
		glog.Warningf(ctx, "[mcp-bridge] tools/call missing wxId deviceNo=%s", h.deviceNo)
		return NewErrorCallResult("绑定缺少账号信息，无法校验开通状态")
	}
	unlocked, err := cashclient.RemoteXiaozhiMcpUnlocked(ctx, h.wxId)
	if err != nil {
		glog.Warningf(ctx, "[mcp-bridge] entitlement check failed wxId=%d err=%v", h.wxId, err)
		return NewErrorCallResult("暂时无法校验小智 MCP 开通状态，请稍后重试")
	}
	if !unlocked {
		glog.Infof(ctx, "[mcp-bridge] entitlement denied wxId=%d deviceNo=%s", h.wxId, h.deviceNo)
		if h.onEntitlementDenied != nil {
			go h.onEntitlementDenied()
		}
		return NewErrorCallResult("小智 MCP 未开通或试用已过期，请先开通后再使用")
	}

	transcript := readTranscriptArg(arguments)
	if transcript == "" {
		return NewErrorCallResult("transcript 不能为空")
	}
	// xzDeviceNo 仅观测日志；喂养落点 MUST 使用 Bridge 绑定的 h.deviceNo，禁止被工具入参覆盖。
	if raw, ok := arguments["xzDeviceNo"]; ok && raw != nil {
		glog.Infof(ctx, "[mcp-bridge] tools/call xzDeviceNo=%v boundDeviceNo=%s", raw, h.deviceNo)
	}
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
