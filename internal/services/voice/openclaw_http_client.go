package voice

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// OpenClawHTTPClient 经 OpenAI 兼容 HTTP 调用 OpenClaw 门禁/Gateway。
//
// 业务说明：
// Intent/Clinic/Care 编排权威在 Gateway；本客户端只传话术、session、Go 已选型的 model。
// G Token（gatewayToken）走 Authorization；A Token（apiToken）走 x-pangbao-api-token 供 Python tools 路由。
// 不解析结构化意图信封；不调用飞轮。
type OpenClawHTTPClient struct {
	baseURL    string
	token      string // G：调智能体门禁
	apiToken   string // A：落库/读史路由
	httpClient *http.Client
}

// OpenClawChatRequest OpenAI Chat Completions 子集。
type OpenClawChatRequest struct {
	Model    string              `json:"model"` // openclaw/intent|clinic|care_alert|default
	Messages []OpenClawChatMessage `json:"messages"`
	Stream   bool                `json:"stream,omitempty"`
	User     string              `json:"user,omitempty"`
}

// OpenClawChatMessage 单条消息。
type OpenClawChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenClawChatResponse 非流式响应。
type OpenClawChatResponse struct {
	Choices []struct {
		Message struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// OpenClawFromCfg 从配置/环境创建 Gateway/门禁客户端。
//
// 配置键：openclaw.gatewayUrl / openclaw.gatewayToken / openclaw.apiToken；
// 环境变量：OPENCLAW_GATEWAY_URL（建议指向 Python /agent-gate）、OPENCLAW_GATEWAY_TOKEN（G）、PANGBAO_API_TOKEN（A）。
func OpenClawFromCfg() *OpenClawHTTPClient {
	ctx := context.Background()
	url := strings.TrimSpace(g.Cfg().MustGet(ctx, "openclaw.gatewayUrl", "").String())
	if url == "" {
		url = strings.TrimSpace(os.Getenv("OPENCLAW_GATEWAY_URL"))
	}
	if url == "" {
		url = "http://127.0.0.1:8000/agent-gate"
	}
	token := strings.TrimSpace(g.Cfg().MustGet(ctx, "openclaw.gatewayToken", "").String())
	if token == "" {
		token = strings.TrimSpace(os.Getenv("OPENCLAW_GATEWAY_TOKEN"))
	}
	apiTok := strings.TrimSpace(g.Cfg().MustGet(ctx, "openclaw.apiToken", "").String())
	if apiTok == "" {
		apiTok = strings.TrimSpace(os.Getenv("PANGBAO_API_TOKEN"))
	}
	return &OpenClawHTTPClient{
		baseURL:  strings.TrimRight(url, "/"),
		token:    token,
		apiToken: apiTok,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// setOpenClawHeaders 写入 G/A 与注模、会话头。
func (c *OpenClawHTTPClient) setOpenClawHeaders(req *http.Request, backendModel, sessionKey string) {
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if c.apiToken != "" {
		req.Header.Set("x-pangbao-api-token", c.apiToken)
	}
	if strings.TrimSpace(backendModel) != "" {
		req.Header.Set("x-openclaw-model", strings.TrimSpace(backendModel))
	}
	if strings.TrimSpace(sessionKey) != "" {
		req.Header.Set("x-openclaw-session-key", strings.TrimSpace(sessionKey))
	}
}

// Chat 非流式 agent 回合。
//
// agentModel：openclaw/intent 等；backendModel：x-openclaw-model（Go 注模）；sessionKey：稳定会话键。
func (c *OpenClawHTTPClient) Chat(
	ctx context.Context,
	agentModel string,
	backendModel string,
	sessionKey string,
	userText string,
) (*OpenClawChatResponse, error) {
	body, _ := json.Marshal(OpenClawChatRequest{
		Model: agentModel,
		Messages: []OpenClawChatMessage{
			{Role: "user", Content: userText},
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setOpenClawHeaders(req, backendModel, sessionKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openclaw chat 失败: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openclaw chat HTTP %d: %s", resp.StatusCode, string(raw))
	}
	var out OpenClawChatResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("解析 openclaw 响应失败: %w", err)
	}
	glog.Debugf(ctx, "[OpenClaw] chat ok agent=%s session=%s", agentModel, sessionKey)
	return &out, nil
}

// ReplyText 取首条 assistant content。
func (r *OpenClawChatResponse) ReplyText() string {
	if r == nil || len(r.Choices) == 0 {
		return ""
	}
	return strings.TrimSpace(r.Choices[0].Message.Content)
}

// openClawBackendModel 将 Go 选模结果格式化为 x-openclaw-model：provider/name。
func openClawBackendModel(cfg *PythonModelCfg) string {
	if cfg == nil {
		return ""
	}
	p := strings.TrimSpace(cfg.Provider)
	n := strings.TrimSpace(cfg.Name)
	if p == "" || n == "" {
		return ""
	}
	return p + "/" + n
}

// EmitCareCardsItems 从 tool_calls 中解析 emit_care_cards 的 items（Care 出卡权威）。
//
// 业务逻辑：优先读 function.arguments JSON 的 items；无 tool_calls 时尝试 content 整段 JSON。
func (r *OpenClawChatResponse) EmitCareCardsItems() []map[string]interface{} {
	if r == nil || len(r.Choices) == 0 {
		return nil
	}
	msg := r.Choices[0].Message
	for _, tc := range msg.ToolCalls {
		name := strings.TrimSpace(tc.Function.Name)
		if name != "emit_care_cards" && name != "emit_care_alert_cards" {
			continue
		}
		var args struct {
			Items []map[string]interface{} `json:"items"`
		}
		if json.Unmarshal([]byte(tc.Function.Arguments), &args) == nil && len(args.Items) > 0 {
			return args.Items
		}
	}
	content := strings.TrimSpace(msg.Content)
	if content == "" {
		return nil
	}
	var asObj struct {
		Items []map[string]interface{} `json:"items"`
	}
	if json.Unmarshal([]byte(content), &asObj) == nil && len(asObj.Items) > 0 {
		return asObj.Items
	}
	var asArr []map[string]interface{}
	if json.Unmarshal([]byte(content), &asArr) == nil && len(asArr) > 0 {
		return asArr
	}
	return nil
}

// StreamChat 流式 SSE；onDelta 收到 content 增量。
func (c *OpenClawHTTPClient) StreamChat(
	ctx context.Context,
	agentModel string,
	backendModel string,
	sessionKey string,
	userText string,
	onDelta func(delta string) error,
) (full string, err error) {
	body, _ := json.Marshal(OpenClawChatRequest{
		Model:  agentModel,
		Stream: true,
		Messages: []OpenClawChatMessage{
			{Role: "user", Content: userText},
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	c.setOpenClawHeaders(req, backendModel, sessionKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openclaw stream 失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openclaw stream HTTP %d: %s", resp.StatusCode, string(raw))
	}
	var b strings.Builder
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if json.Unmarshal([]byte(data), &chunk) != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		b.WriteString(delta)
		if onDelta != nil {
			if cbErr := onDelta(delta); cbErr != nil {
				return b.String(), cbErr
			}
		}
	}
	return b.String(), sc.Err()
}
