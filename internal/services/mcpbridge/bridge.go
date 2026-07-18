package mcpbridge

// bridge.go 实现小智 MCP 接入点的 WebSocket 传输与 JSON-RPC 桥接。
//
// 连接方向：本服务（MCP Server）主动拨号连接小智接入点（MCP Client）。
// 协议：WebSocket + JSON-RPC 2.0；MCP 协议子集（initialize / notifications/initialized / tools/list / tools/call）。
// 容错：断线自动重连（指数退避，2s→60s），重连不退出进程。

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/os/glog"
	"github.com/gorilla/websocket"
)

// 默认重连参数；可被 Bridge 构造选项覆盖。
const (
	defaultReconnectMin = 2 * time.Second
	defaultReconnectMax = 60 * time.Second
	dialHandshakeTimeout = 30 * time.Second
	// WebSocket 读写超时：读不设超时（长连接等待），写单帧 10s 超时。
	writeWait = 10 * time.Second
	// ping 周期：维持中间件 NAT 不超时回收。
	pingPeriod = 30 * time.Second
)

// Bridge 维护与小智 MCP 接入点的 WebSocket 长连接，并在其上承载 MCP JSON-RPC。
type Bridge struct {
	baseURL       string // 接入点基址，如 wss://api.xiaozhi.me/mcp/
	token         string // 接入点 token（拼接到 baseURL 的 query）
	deviceNo      string // 绑定的设备号（注入 chat handler）
	chat          *ChatHandler
	reconnectMin  time.Duration
	reconnectMax  time.Duration
	dialer        *websocket.Dialer
	writeMu       sync.Mutex // 串行化 WebSocket 写，避免帧交错
}

// NewBridge 构造 Bridge。
// baseURL / token / deviceNo 由 main.go 从环境变量读取并校验后传入。
// reconnectMin / reconnectMax 为 0 时使用默认值。
func NewBridge(baseURL, token, deviceNo string, reconnectMin, reconnectMax time.Duration) *Bridge {
	if reconnectMin <= 0 {
		reconnectMin = defaultReconnectMin
	}
	if reconnectMax <= 0 {
		reconnectMax = defaultReconnectMax
	}
	return &Bridge{
		// 不 trim 末尾斜杠：小智接入点路径为 /mcp/（带斜杠），
		// 部分服务端对 /mcp 与 /mcp/ 路径匹配敏感（301/404），保留用户配置原样。
		baseURL:      baseURL,
		token:        strings.TrimSpace(token),
		deviceNo:     strings.TrimSpace(deviceNo),
		chat:         NewChatHandler(deviceNo),
		reconnectMin: reconnectMin,
		reconnectMax: reconnectMax,
		// 复用 gorilla/websocket 默认 Dialer，仅覆盖握手超时。
		dialer: &websocket.Dialer{
			HandshakeTimeout: dialHandshakeTimeout,
		},
	}
}

// endpointURL 拼接完整的接入点 URL：baseURL?token=<token>。
// 处理 baseURL 已带 query 的情况，确保 token 不被覆盖。
func (b *Bridge) endpointURL() string {
	u, err := url.Parse(b.baseURL)
	if err != nil {
		// baseURL 非法时直接拼接，由 Dial 阶段报错。
		return b.baseURL + "?token=" + b.token
	}
	q := u.Query()
	if strings.TrimSpace(q.Get("token")) == "" {
		q.Set("token", b.token)
		u.RawQuery = q.Encode()
	}
	return u.String()
}

// Run 启动桥接循环：拨号 → 读循环 → 断线后指数退避重连。
// 阻塞直到 ctx 被取消（SIGTERM/SIGINT）。
// 退避策略：拨号失败时倍增退避；拨号成功后连接断开时重置为 reconnectMin。
func (b *Bridge) Run(ctx context.Context) error {
	backoff := b.reconnectMin
	for {
		if ctx.Err() != nil {
			// 收到关闭信号，退出循环。
			return ctx.Err()
		}
		connected, err := b.dialAndServe(ctx)
		if err != nil {
			glog.Errorf(ctx, "[mcp-bridge] serve ended err=%v", err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if connected {
			// 上次拨号成功并维持了一段时间：重置退避，下次断线立即重连。
			backoff = b.reconnectMin
		}
		glog.Infof(ctx, "[mcp-bridge] reconnect in %v", backoff)
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return ctx.Err()
		}
		if !connected {
			// 拨号失败：指数退避倍增并封顶。
			backoff *= 2
			if backoff > b.reconnectMax {
				backoff = b.reconnectMax
			}
		}
	}
}

// dialAndServe 执行一次完整的「拨号 + 读循环」。
// 返回 (connected, err)：
//   - connected=true 表示拨号成功并进入了读循环（无论后续因何原因退出）；
//   - connected=false 表示拨号阶段就失败，未进入读循环；
//   - err 非 nil 时携带退出原因。
// 连接的 Close 由本函数 defer 处理，调用方无需关心。
func (b *Bridge) dialAndServe(ctx context.Context) (bool, error) {
	endpoint := b.endpointURL()
	glog.Infof(ctx, "[mcp-bridge] dialing %s", maskToken(endpoint))
	// 携带 User-Agent 便于小智服务端识别本接入服务，便于排障。
	headers := http.Header{}
	headers.Set("User-Agent", "xiaozhi-mcp-service/1.0")
	conn, resp, err := b.dialer.DialContext(ctx, endpoint, headers)
	if err != nil {
		// bad handshake 时 gorilla/websocket 仍会返回响应体（含服务端错误说明），
		// 读取后拼入错误信息，便于排查（如 token 失效、路径 404、协议不匹配）。
		detail := readHandshakeErrorDetail(resp)
		return false, fmt.Errorf("dial xiaozhi mcp endpoint: %w%s", err, detail)
	}
	// 成功路径下 resp.Body 由 Dial 内部处理，无需手动关闭。
	defer conn.Close()
	glog.Infof(ctx, "[mcp-bridge] connected, entering read loop")
	// 启动 ping 维活 goroutine：定时发 Ping，避免 NAT/中间件超时回收连接。
	// pingCtx 在 readLoop 退出后随 defer cancel 取消，避免 goroutine 泄漏。
	pingCtx, pingCancel := context.WithCancel(ctx)
	defer pingCancel()
	go b.runPing(pingCtx, conn)
	return true, b.readLoop(ctx, conn)
}

// readHandshakeErrorDetail 从 dial 失败的 HTTP 响应中提取状态码与 Body 摘要。
// 返回形如 " (status=401, body=token expired)" 的字符串；无响应体时返回空串。
// 该函数仅用于日志与错误信息增强，不参与协议判断。
func readHandshakeErrorDetail(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	detail := fmt.Sprintf(" (status=%d)", resp.StatusCode)
	if resp.Body == nil {
		return detail
	}
	// 限制读取量，避免服务端返回超大错误页时拖累日志。
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512))
	_ = resp.Body.Close()
	if err != nil || len(body) == 0 {
		return detail
	}
	return fmt.Sprintf("%s body=%s", detail, strings.TrimSpace(string(body)))
}

// readLoop 持续读取 WebSocket 文本帧并分发到 MCP handler。
// 遇到任何错误（读失败、ctx 取消）即返回，由上层决定重连。
func (b *Bridge) readLoop(ctx context.Context, conn *websocket.Conn) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		_, payload, err := conn.ReadMessage()
		if err != nil {
			// 正常关闭 / 异常断开均视为需要重连。
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return fmt.Errorf("websocket closed by peer: %v", err)
			}
			return fmt.Errorf("read message: %w", err)
		}
		// 解析 + 分发；单条消息处理失败不影响后续帧。
		if err := b.handleFrame(ctx, conn, payload); err != nil {
			glog.Errorf(ctx, "[mcp-bridge] handle frame err=%v", err)
		}
	}
}

// handleFrame 处理一条 JSON-RPC 请求/通知。
// 通知（无 ID）不回写响应；请求回写响应帧。
func (b *Bridge) handleFrame(ctx context.Context, conn *websocket.Conn, payload []byte) error {
	req, err := ParseRequest(payload)
	if err != nil {
		// 无法解析的帧：记录后忽略，不回写（无法确定 ID）。
		glog.Errorf(ctx, "[mcp-bridge] parse frame failed raw=%s err=%v", string(payload), err)
		return err
	}
	// 通知：无 ID，不响应。
	if len(req.ID) == 0 || string(req.ID) == "null" {
		// 仅处理已知通知，未知通知忽略。
		switch req.Method {
		case "notifications/initialized":
			glog.Infof(ctx, "[mcp-bridge] received initialized notification")
		default:
			glog.Debugf(ctx, "[mcp-bridge] ignore notification method=%s", req.Method)
		}
		return nil
	}
	// 请求：分发到对应 handler 并回写响应。
	resp := b.dispatch(ctx, req)
	if resp == nil {
		// handler 已自行处理（理论上不会发生），保守跳过。
		return nil
	}
	raw, err := MarshalResponse(resp)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}
	return b.writeFrame(conn, raw)
}

// dispatch 根据 method 分发到对应 MCP handler，返回响应。
// 未知 method 返回 -32601 method not found。
func (b *Bridge) dispatch(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return b.handleInitialize(ctx, req)
	case "tools/list":
		return b.handleToolsList(ctx, req)
	case "tools/call":
		return b.handleToolsCall(ctx, req)
	case "ping":
		// MCP ping method：返回空 result。
		return NewSuccessResponse(req.ID, map[string]any{})
	default:
		return NewErrorResponse(req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
	}
}

// handleInitialize 响应 initialize 请求。
// 不解析 params（小智 clientInfo 不影响本服务行为），直接返回 server 声明。
func (b *Bridge) handleInitialize(_ context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	return NewSuccessResponse(req.ID, NewInitializeResult(ServerName(), ServerVersion()))
}

// handleToolsList 响应 tools/list 请求。
// 静态返回 chat 工具定义。
func (b *Bridge) handleToolsList(_ context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	return NewSuccessResponse(req.ID, &ToolsListResult{Tools: ListTools()})
}

// handleToolsCall 响应 tools/call 请求。
// 仅支持 chat 工具；其他工具名返回 -32601。
func (b *Bridge) handleToolsCall(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	var params ToolsCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(req.ID, -32602, fmt.Sprintf("invalid tools/call params: %v", err))
	}
	if params.Name != chatToolName {
		return NewErrorResponse(req.ID, -32601, fmt.Sprintf("tool not found: %s", params.Name))
	}
	result := b.chat.Handle(ctx, params.Arguments)
	return NewSuccessResponse(req.ID, result)
}

// writeFrame 串行化写一帧文本，设置写超时避免阻塞读循环。
func (b *Bridge) writeFrame(conn *websocket.Conn, payload []byte) error {
	b.writeMu.Lock()
	defer b.writeMu.Unlock()
	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, payload)
}

// runPing 周期性发送 WebSocket Ping 帧维持连接。
// 收到 ctx 取消时退出；写失败时让上层读循环感知断开。
func (b *Bridge) runPing(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.writeMu.Lock()
			err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait))
			b.writeMu.Unlock()
			if err != nil {
				// ping 写失败：连接已不可用，退出 ping goroutine；读循环会随后感知断开。
				return
			}
		}
	}
}

// maskToken 脱敏 endpoint URL 中的 token，仅保留前 6 位用于日志识别。
func maskToken(s string) string {
	idx := strings.Index(s, "token=")
	if idx < 0 {
		return s
	}
	prefix := s[:idx]
	rest := s[idx+len("token="):]
	if len(rest) <= 6 {
		return prefix + "token=***"
	}
	return prefix + "token=" + rest[:6] + "***"
}
