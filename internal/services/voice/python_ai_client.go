package voice

import (
	"bufio"
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

// PythonAIClient 封装对 python-ai-talk 微服务的 HTTP 调用
// 用于将 go 侧的意图分析和胖宝诊疗能力委托给 Python 服务
type PythonAIClient struct {
	baseURL    string       // Python 服务的基础地址，如 http://python-ai-talk:8000
	httpClient *http.Client // 复用的 HTTP 客户端
}

// NewPythonAIClient 创建 Python AI 服务客户端
// baseURL 参数为 Python 服务的完整地址（不含尾部斜杠）
func NewPythonAIClient(baseURL string) *PythonAIClient {
	return &PythonAIClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // 超时时间设为 60 秒，覆盖 LLM 调用的较长响应时间
		},
	}
}

// AnalyzeIntentRequest 意图分析请求体
// 与 Python 服务的 /v1/analyze/intent 接口对齐
type AnalyzeIntentRequest struct {
	Text     string         `json:"text"`      // 用户输入的自然语言文本
	DeviceNo string         `json:"device_no"` // 设备编号
	Model    PythonModelCfg `json:"model"`     // 模型配置
}

// PythonModelCfg 传递给 Python 服务的模型配置
// 由 go 侧从 llmLanes 配置中读取，不在 Python 中轮询
type PythonModelCfg struct {
	Provider    string `json:"provider"`      // 模型提供商，如 deepseek、zhipu
	Name        string `json:"name"`          // 模型名称，如 deepseek-chat
	MaxInFlight int    `json:"max_in_flight"` // 最大并发数
}

// AnalyzeIntentResponse 意图分析响应体
// 与 Python 服务的返回结构对齐
type AnalyzeIntentResponse struct {
	TargetType string   `json:"target_type"` // 目标类型：feeding|history|suggest|conversation|exit
	Action     string   `json:"action"`      // 动作类型：start|end|one|search|suggestion|reply|exit
	EventName  string   `json:"event_name"`  // 匹配到的事件名称（喂养场景）
	Keywords   []string `json:"keywords"`    // 匹配到的关键词列表
	Content    string   `json:"content"`     // 对话场景的回答内容
}

// AnalyzeIntent 调用 Python 服务进行意图分析
// ctx: 上下文，用于超时和取消
// req: 意图分析请求
// 返回：意图分析响应和错误
func (c *PythonAIClient) AnalyzeIntent(ctx context.Context, req *AnalyzeIntentRequest) (*AnalyzeIntentResponse, error) {
	// 将请求体序列化为 JSON
	body, _ := json.Marshal(req)

	// 创建 HTTP POST 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/analyze/intent", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("创建意图分析请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("调用 Python 意图分析服务失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Python 意图分析服务返回错误状态码 %d: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应体
	var result AnalyzeIntentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析意图分析响应失败: %w", err)
	}

	glog.Debugf(ctx, "[Python AI] 意图分析完成。deviceNo=%s target_type=%s action=%s", req.DeviceNo, result.TargetType, result.Action)
	return &result, nil
}

// ClinicStreamRequest 胖宝诊疗流式请求体
// 与 Python 服务的 /v1/clinic/stream 接口对齐
type ClinicStreamRequest struct {
	Question string         `json:"question"`  // 用户的诊疗问题
	DeviceNo string         `json:"device_no"` // 设备编号
	Model    PythonModelCfg `json:"model"`     // 模型配置
}

// ClinicStreamCallback 诊疗流式回调
// 用于将流式响应分块传递给调用方
type ClinicStreamCallback struct {
	OnThinking func(delta string) error // 收到思考过程片段时的回调
	OnAnswer   func(delta string) error // 收到回答内容片段时的回调
}

// ClinicStream 调用 Python 服务进行流式诊疗
// ctx: 上下文
// req: 诊疗请求
// cb: 流式回调
// 返回：完整的思考过程、完整的回答内容、错误
func (c *PythonAIClient) ClinicStream(ctx context.Context, req *ClinicStreamRequest, cb *ClinicStreamCallback) (thinking, answer string, err error) {
	// 将请求体序列化为 JSON
	body, _ := json.Marshal(req)

	// 创建 HTTP POST 请求，要求流式响应
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/clinic/stream", strings.NewReader(string(body)))
	if err != nil {
		return "", "", fmt.Errorf("创建诊疗流式请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("调用 Python 诊疗流式服务失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("Python 诊疗流式服务返回错误状态码 %d: %s", resp.StatusCode, string(respBody))
	}

	// 逐行解析 SSE 响应
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		// SSE 格式：每行以 "data: " 开头
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		// 提取 data 后面的 JSON 内容
		data := strings.TrimPrefix(line, "data: ")
		var event struct {
			Type    string `json:"type"`    // 消息类型：thinking 或 answer
			Content string `json:"content"` // 内容片段
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue // 跳过无法解析的行
		}
		// 根据类型分发到对应回调
		switch event.Type {
		case "thinking":
			// 累积思考过程
			thinking += event.Content
			if cb != nil && cb.OnThinking != nil {
				if cbErr := cb.OnThinking(event.Content); cbErr != nil {
					return thinking, answer, cbErr
				}
			}
		case "answer":
			// 累积回答内容
			answer += event.Content
			if cb != nil && cb.OnAnswer != nil {
				if cbErr := cb.OnAnswer(event.Content); cbErr != nil {
					return thinking, answer, cbErr
				}
			}
		}
	}

	return thinking, answer, scanner.Err()
}

// pythonAIClientFromCfg 从配置中读取 Python 服务地址并创建客户端
// 优先从配置文件 pythonAiTalk.url 读取，其次环境变量 PYTHON_AI_TALK_URL，默认 http://python-ai-talk:8000
func pythonAIClientFromCfg() *PythonAIClient {
	// 尝试从配置文件读取
	url := g.Cfg().MustGet(context.Background(), "pythonAiTalk.url", "").String()
	if url == "" {
		// 回退到环境变量
		url = os.Getenv("PYTHON_AI_TALK_URL")
	}
	if url == "" {
		// 默认值：Docker 网络中的服务名
		url = "http://python-ai-talk:8000"
	}
	return NewPythonAIClient(url)
}
