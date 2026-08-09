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

	contracts "hello/internal/services/contracts"

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
// 与 Python 服务的 /v1/analyze/intent 及 /v1/analyze/intent/stream 对齐（流式复用本结构）。
// 同一输入框续聊：上一轮 need_confirm 返回的 conversation_id 在本轮可选带回，供 Python 解析自由文本澄清。
type AnalyzeIntentRequest struct {
	Text           string          `json:"text"`                      // 用户输入的自然语言文本
	DeviceNo       string          `json:"device_no"`                 // 设备编号
	Model          *PythonModelCfg `json:"model,omitempty"`           // 模型配置；nil=omit，由 Python 自选免费模
	ConversationID string          `json:"conversation_id,omitempty"` // 可选；有本地澄清 pending 时带回以续聊
}

// PythonModelCfg 传递给 Python 服务的模型配置
// 由 go 侧经 ResolveLaneModel 填入；非 premium 且 free 为空时整段 omit。
type PythonModelCfg struct {
	Provider    string `json:"provider"`      // 模型提供商，如 deepseek、zhipu
	Name        string `json:"name"`          // 模型名称，如 deepseek-v4-flash
	MaxInFlight int    `json:"max_in_flight"` // 最大并发数
}

// IntentEvent 单个事件结构
// 用于描述多事件场景中的单个事件
type IntentEvent struct {
	Action    string `json:"action"`
	EventName string `json:"event_name"`
	EventId   string `json:"event_id"`
	Quantity  *int   `json:"quantity,omitempty"` // 从用户输入中提取的数量值
}

// AnalyzeIntentResponse 意图分析响应体
// 与 Python 服务的返回结构对齐
type AnalyzeIntentResponse struct {
	TargetType     string        `json:"target_type"`            // 目标类型：feeding|history|suggest|conversation|exit
	Action         string        `json:"action"`                 // 动作类型：start|end|one|search|suggestion|reply|exit|multi
	EventName      string        `json:"event_name"`             // 匹配到的事件名称（喂养场景，单事件时使用）
	EventId        string        `json:"event_id"`               // 事件ID（单事件时使用）
	Quantity       *int          `json:"quantity,omitempty"`     // 从用户输入中提取的数量值（Python 前置提取）
	EventType      string        `json:"event_type,omitempty"`   // 事件类型：number|time|one（新事件时 Python 返回）
	EventUnit      string        `json:"event_unit,omitempty"`   // 事件单位：ml、次、分钟（新事件时 Python 返回）
	IsNewEvent     bool          `json:"is_new_event,omitempty"` // 是否为新事件
	Keywords       []string      `json:"keywords"`               // 匹配到的关键词列表
	Content        string        `json:"content"`                // 对话场景的回答内容
	Events         []IntentEvent `json:"events"`                 // 多事件列表（当 action 为 multi 时使用）
	NeedConfirm    bool          `json:"need_confirm"`           // 是否需要用户澄清（同一 /intent 输入框续聊，非独立 confirm 通道）
	ConfirmMessage string        `json:"confirm_message"`        // 澄清话术（由 Python 生成，Go 侧原样透传）
	ConversationID string        `json:"conversation_id"`        // 会话 ID；need_confirm 时返回，下一轮请求带回同一 /intent 续聊
}

// AnalyzeIntent 调用 Python 服务进行意图分析
// ctx: 上下文，用于超时和取消
// req: 意图分析请求（可含 conversation_id 续聊）
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

	glog.Debugf(ctx, "[Python AI] 意图分析完成。deviceNo=%s target_type=%s action=%s conversation_id=%s need_confirm=%v",
		req.DeviceNo, result.TargetType, result.Action, result.ConversationID, result.NeedConfirm)
	return &result, nil
}

// ClinicStreamRequest 胖宝诊疗流式请求体
// 与 Python 服务的 /v1/clinic/stream 接口对齐
type ClinicStreamRequest struct {
	Question string          `json:"question"`        // 用户的诊疗问题
	DeviceNo string          `json:"device_no"`       // 设备编号
	Model    *PythonModelCfg `json:"model,omitempty"` // 模型配置；nil=omit
}

// ClinicStreamCallback 诊疗流式回调
// 用于将流式响应分块传递给调用方
type ClinicStreamCallback struct {
	OnThinking func(delta string) error    // 收到思考过程片段时的回调
	OnAnswer   func(delta string) error    // 收到回答内容片段时的回调
	OnDone     func(answerID string) error // 收到完成事件时的回调（包含 answer_id 用于反馈）
}

// ClinicStream 调用 Python 服务进行流式诊疗
// ctx: 上下文
// req: 诊疗请求
// cb: 流式回调
// 返回：完整的思考过程、完整的回答内容、错误
// ClinicStreamResponse 诊疗流式响应结果
type ClinicStreamResponse struct {
	Thinking string // 完整的思考过程
	Answer   string // 完整的回答内容
	AnswerID string // 回答 ID（用于提交反馈）
}

func (c *PythonAIClient) ClinicStream(ctx context.Context, req *ClinicStreamRequest, cb *ClinicStreamCallback) (*ClinicStreamResponse, error) {
	// 将请求体序列化为 JSON
	body, _ := json.Marshal(req)

	// 创建 HTTP POST 请求，要求流式响应
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/clinic/stream", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("创建诊疗流式请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("调用 Python 诊疗流式服务失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Python 诊疗流式服务返回错误状态码 %d: %s", resp.StatusCode, string(respBody))
	}

	// 逐行解析 SSE 响应
	var thinking, answer, answerID string
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
			Type     string `json:"type"`      // 消息类型：thinking、answer、done
			Content  string `json:"content"`   // 内容片段
			AnswerID string `json:"answer_id"` // 回答 ID（done 事件时返回）
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
					return &ClinicStreamResponse{Thinking: thinking, Answer: answer, AnswerID: answerID}, cbErr
				}
			}
		case "answer":
			// 累积回答内容
			answer += event.Content
			if cb != nil && cb.OnAnswer != nil {
				if cbErr := cb.OnAnswer(event.Content); cbErr != nil {
					return &ClinicStreamResponse{Thinking: thinking, Answer: answer, AnswerID: answerID}, cbErr
				}
			}
		case "done":
			// 记录回答 ID，用于后续反馈
			answerID = event.AnswerID
			if cb != nil && cb.OnDone != nil {
				if cbErr := cb.OnDone(event.AnswerID); cbErr != nil {
					return &ClinicStreamResponse{Thinking: thinking, Answer: answer, AnswerID: answerID}, cbErr
				}
			}
		}
	}

	return &ClinicStreamResponse{Thinking: thinking, Answer: answer, AnswerID: answerID}, scanner.Err()
}

// PythonAIClientFromCfg 从配置中读取 Python 服务地址并创建客户端
// 优先从配置文件 pythonAiTalk.url 读取，其次环境变量 PYTHON_AI_TALK_URL，默认 http://python-ai-talk:8000
func PythonAIClientFromCfg() *PythonAIClient {
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

// TipStreamRequest 小贴士流式请求体
// 与 Python 服务的 /v1/tip/stream 接口对齐（snake_case）。
// 月龄与当前时间由 Python 派生，本结构不再携带 baby_age_months / current_time。
type TipStreamRequest struct {
	EventID   int64           `json:"event_id"`         // 触发事件 ID
	EventName string          `json:"event_name"`       // 触发事件名称
	DeviceNo  string          `json:"device_no"`        // 设备编号
	Model     *PythonModelCfg `json:"model,omitempty"`  // 模型配置；nil=omit
}

// TipStream 调用 Python 服务进行流式小贴士生成
// ctx: 上下文
// req: 小贴士请求
// cb: 流式回调
// 返回：小贴士流式响应结果和错误
func (c *PythonAIClient) TipStream(ctx context.Context, req *TipStreamRequest, cb *contracts.TipStreamCallback) (*contracts.TipStreamResponse, error) {
	// 将请求体序列化为 JSON
	body, _ := json.Marshal(req)

	// 创建 HTTP POST 请求，要求流式响应
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/tip/stream", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("创建小贴士流式请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("调用 Python 小贴士流式服务失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Python 小贴士流式服务返回错误状态码 %d: %s", resp.StatusCode, string(respBody))
	}

	// 逐行解析 SSE 响应
	var thinking, answer, answerID string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var event struct {
			Type     string `json:"type"`      // 消息类型：thinking、answer、done
			Content  string `json:"content"`   // 内容片段
			AnswerID string `json:"answer_id"` // 回答 ID（done 事件时返回）
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		switch event.Type {
		case "thinking":
			thinking += event.Content
			if cb != nil && cb.OnThinking != nil {
				if cbErr := cb.OnThinking(event.Content); cbErr != nil {
					return &contracts.TipStreamResponse{Thinking: thinking, Answer: answer, AnswerID: answerID}, cbErr
				}
			}
		case "answer":
			answer += event.Content
			if cb != nil && cb.OnAnswer != nil {
				if cbErr := cb.OnAnswer(event.Content); cbErr != nil {
					return &contracts.TipStreamResponse{Thinking: thinking, Answer: answer, AnswerID: answerID}, cbErr
				}
			}
		case "done":
			answerID = event.AnswerID
			if cb != nil && cb.OnDone != nil {
				if cbErr := cb.OnDone(event.AnswerID); cbErr != nil {
					return &contracts.TipStreamResponse{Thinking: thinking, Answer: answer, AnswerID: answerID}, cbErr
				}
			}
		}
	}

	return &contracts.TipStreamResponse{Thinking: thinking, Answer: answer, AnswerID: answerID}, scanner.Err()
}

// FeedbackRequest 反馈请求体
// 与 Python 服务的 /v1/clinic/feedback 和 /v1/tip/feedback 接口对齐
type FeedbackRequest struct {
	AnswerID string `json:"answer_id"` // 回答 ID（来自流式响应的 done 事件）
	Feedback int    `json:"feedback"`  // 反馈值：1=thumbs up, -1=thumbs down
}

// FeedbackResponse 反馈响应体
type FeedbackResponse struct {
	Code    int    `json:"code"`    // 状态码：0=成功
	Message string `json:"message"` // 提示信息
	Data    struct {
		AnswerID string `json:"answer_id"` // 回答 ID
		Feedback int    `json:"feedback"`  // 反馈值
	} `json:"data"`
}

// ClinicFeedback 提交诊疗反馈
// ctx: 上下文
// req: 反馈请求
// 返回：反馈响应和错误
func (c *PythonAIClient) ClinicFeedback(ctx context.Context, req *FeedbackRequest) (*FeedbackResponse, error) {
	return c.submitFeedback(ctx, "/v1/clinic/feedback", req)
}

// TipFeedback 提交小贴士反馈
// ctx: 上下文
// req: 反馈请求
// 返回：反馈响应和错误
func (c *PythonAIClient) TipFeedback(ctx context.Context, req *FeedbackRequest) (*FeedbackResponse, error) {
	return c.submitFeedback(ctx, "/v1/tip/feedback", req)
}

// submitFeedback 提交反馈的通用方法
func (c *PythonAIClient) submitFeedback(ctx context.Context, path string, req *FeedbackRequest) (*FeedbackResponse, error) {
	// 将请求体序列化为 JSON
	body, _ := json.Marshal(req)

	// 创建 HTTP POST 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("创建反馈请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("调用 Python 反馈服务失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Python 反馈服务返回错误状态码 %d: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应体
	var result FeedbackResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析反馈响应失败: %w", err)
	}

	glog.Debugf(ctx, "[Python AI] 反馈提交完成。path=%s answer_id=%s feedback=%d", path, req.AnswerID, req.Feedback)
	return &result, nil
}

// AnalyzeIntentStreamRequest 流式意图分析请求体
// 与非流式 AnalyzeIntentRequest 字段完全一致，使用类型别名以减少重复定义
type AnalyzeIntentStreamRequest = AnalyzeIntentRequest

// AnalyzeIntentStreamCallback 流式意图分析回调
// 用于将 SSE 流式响应分块传递给调用方
type AnalyzeIntentStreamCallback struct {
	OnThinking func(delta string) error                  // 收到思考过程片段时的回调
	OnAnswer   func(delta string) error                  // 收到回答内容片段时的回调（answer 内容为 JSON 格式的意图结果）
	OnDone     func(result *AnalyzeIntentResponse) error // 收到完成事件时的回调（包含完整的意图分析结果）
}

// AnalyzeIntentStreamResponse 流式意图分析响应结果
type AnalyzeIntentStreamResponse struct {
	Thinking string                 // 完整的思考过程
	Answer   string                 // 完整的回答内容（JSON 格式的意图结果）
	Result   *AnalyzeIntentResponse // 解析后的意图分析结果
}

// AnalyzeIntentStream 调用 Python 服务进行流式意图分析
// 以 SSE 流式方式调用 /v1/analyze/intent/stream 接口，逐块返回思考过程和意图结果
// ctx: 上下文，用于超时和取消
// req: 流式意图分析请求（与非流式请求字段一致）
// cb: 流式回调（OnThinking/OnAnswer/OnDone）
// 返回：流式意图分析响应结果（完整思考内容 + 完整意图 JSON + 解析后的意图结果）和错误
func (c *PythonAIClient) AnalyzeIntentStream(ctx context.Context, req *AnalyzeIntentStreamRequest, cb *AnalyzeIntentStreamCallback) (*AnalyzeIntentStreamResponse, error) {
	// 将请求体序列化为 JSON
	body, _ := json.Marshal(req)

	// 创建 HTTP POST 请求，要求流式响应（SSE）
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/analyze/intent/stream", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("创建流式意图分析请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("调用 Python 流式意图分析服务失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Python 流式意图分析服务返回错误状态码 %d: %s", resp.StatusCode, string(respBody))
	}

	// 逐行解析 SSE 响应
	var thinking, answer string
	var result *AnalyzeIntentResponse
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		// SSE 格式：每行以 "data: " 开头
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		// 提取 data 后面的 JSON 内容
		data := strings.TrimPrefix(line, "data: ")
		// 检查是否为结束标记
		if data == "[DONE]" {
			break
		}
		// 解析事件 JSON
		var event struct {
			Type    string `json:"type"`    // 消息类型：thinking、answer
			Content string `json:"content"` // 内容片段
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue // 跳过无法解析的行
		}
		// 根据类型分发到对应回调
		switch event.Type {
		case "thinking":
			// 累积思考过程并触发回调
			thinking += event.Content
			if cb != nil && cb.OnThinking != nil {
				if cbErr := cb.OnThinking(event.Content); cbErr != nil {
					return &AnalyzeIntentStreamResponse{Thinking: thinking, Answer: answer, Result: result}, cbErr
				}
			}
		case "answer":
			// 累积回答内容（answer 为完整的意图分析 JSON）并触发回调
			answer += event.Content
			if cb != nil && cb.OnAnswer != nil {
				if cbErr := cb.OnAnswer(event.Content); cbErr != nil {
					return &AnalyzeIntentStreamResponse{Thinking: thinking, Answer: answer, Result: result}, cbErr
				}
			}
		}
	}

	// 流式结束后，解析完整的 answer 为意图分析结果
	if answer != "" {
		var parsed AnalyzeIntentResponse
		if err := json.Unmarshal([]byte(answer), &parsed); err != nil {
			return &AnalyzeIntentStreamResponse{Thinking: thinking, Answer: answer}, fmt.Errorf("解析流式意图分析结果失败: %w", err)
		}
		result = &parsed
	}

	// 触发 OnDone 回调
	if cb != nil && cb.OnDone != nil {
		if cbErr := cb.OnDone(result); cbErr != nil {
			return &AnalyzeIntentStreamResponse{Thinking: thinking, Answer: answer, Result: result}, cbErr
		}
	}

	glog.Debugf(ctx, "[Python AI] 流式意图分析完成。deviceNo=%s target_type=%s action=%s", req.DeviceNo, result.TargetType, result.Action)
	return &AnalyzeIntentStreamResponse{Thinking: thinking, Answer: answer, Result: result}, scanner.Err()
}

// ---------- 护理留意（care-alert）Go → Python 内部契约 ----------

// CareAlertAnalyzeRequest 护理留意日分析请求（路径见 Python CONTRACT：/v1/care-alert/analyze）。
// Model/ModelCfg 均可 omit：非 premium 且 free 为空时由 Python 自选免费模。
type CareAlertAnalyzeRequest struct {
	DeviceNo       string                 `json:"device_no"`
	Day            string                 `json:"day"`
	Model          string                 `json:"model,omitempty"`     // 可选简写；优先以 ModelCfg 为准
	ModelCfg       *PythonModelCfg        `json:"model_cfg,omitempty"` // 完整模型配置；nil=omit
	AgeMonths      int                    `json:"age_months"`
	HistorySummary map[string]interface{} `json:"history_summary"`
	KgContext      map[string]interface{} `json:"kg_context"`
}

// CareAlertAnalyzeReason Python 返回的原因片段。
type CareAlertAnalyzeReason struct {
	Type            string   `json:"type"`
	Score           float64  `json:"score"`
	ExpectationUsed bool     `json:"expectationUsed"`
	AgeMonths       int      `json:"ageMonths"`
	MedianGapMs     int64    `json:"medianGapMs"`
	LastGapMs       int64    `json:"lastGapMs"`
	ExpectGapMaxMs  int64    `json:"expectGapMaxMs"`
	P75DurMs        int64    `json:"p75DurMs"`
	ElapsedMs       int64    `json:"elapsedMs"`
	ExpectDurMaxMs  int64    `json:"expectDurMaxMs"`
	DailyAvg        float64  `json:"dailyAvg"`
	Recent48hCount  int      `json:"recent48hCount"`
	StillExpected   bool     `json:"stillExpected"`
	DetailLines     []string `json:"detailLines"`
}

// CareAlertAnalyzeItem Python 返回的单条留意（suggestionId 可省略，由 Go 补齐）。
type CareAlertAnalyzeItem struct {
	SuggestionID   string                    `json:"suggestionId"`
	EventID        string                    `json:"eventId"`
	EventName      string                    `json:"eventName"`
	SummaryLine    string                    `json:"summaryLine"`
	FollowUpPrompt string                    `json:"followUpPrompt"`
	Reasons        []CareAlertAnalyzeReason  `json:"reasons"`
}

// CareAlertAnalyzeResponse Python 分析响应。
type CareAlertAnalyzeResponse struct {
	Items []CareAlertAnalyzeItem `json:"items"`
}

// CareAlertFeedbackRequest 固定意图飞轮（无 NLP）。
type CareAlertFeedbackRequest struct {
	DeviceNo     string `json:"device_no"`
	SuggestionID string `json:"suggestion_id"`
	Intent       string `json:"intent"` // ignore|follow_up
	Day          string `json:"day"`
}

// CareAlertAnalyze 调用 Python KG+LLM 日分析；超时单独放宽至 100s。
func (c *PythonAIClient) CareAlertAnalyze(ctx context.Context, req *CareAlertAnalyzeRequest) (*CareAlertAnalyzeResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化护理留意分析请求失败: %w", err)
	}
	// 与 tip/clinic 同前缀 /v1/...（勿用 /internal/，Python 未挂该前缀）
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/care-alert/analyze", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("创建护理留意分析请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 100 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("调用 Python 护理留意分析失败: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Python 护理留意分析返回 %d: %s", resp.StatusCode, string(respBody))
	}
	var result CareAlertAnalyzeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		// 兼容外层 envelope {code,data:{items}}
		var env struct {
			Code int                      `json:"code"`
			Data CareAlertAnalyzeResponse `json:"data"`
		}
		if err2 := json.Unmarshal(respBody, &env); err2 != nil || (env.Code != 0 && len(env.Data.Items) == 0) {
			return nil, fmt.Errorf("解析护理留意分析响应失败: %w", err)
		}
		result = env.Data
	}
	if result.Items == nil {
		result.Items = []CareAlertAnalyzeItem{}
	}
	glog.Debugf(ctx, "[Python AI] 护理留意分析完成。deviceNo=%s day=%s count=%d", req.DeviceNo, req.Day, len(result.Items))
	return &result, nil
}

// CareAlertFeedback 转发固定意图飞轮至 Python（/v1/care-alert/feedback；无 NLP，Python 侧 ACK）。
func (c *PythonAIClient) CareAlertFeedback(ctx context.Context, req *CareAlertFeedbackRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化护理留意飞轮请求失败: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/care-alert/feedback", strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("创建护理留意飞轮请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("调用 Python 护理留意飞轮失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Python 护理留意飞轮返回 %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
