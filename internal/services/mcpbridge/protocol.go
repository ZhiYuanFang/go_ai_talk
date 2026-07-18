package mcpbridge

import (
	"encoding/json"
	"fmt"
)

// JSONRPCRequest 是 JSON-RPC 2.0 请求的通用结构。
// 小智 MCP 接入点作为 Client，向本服务发送请求/通知。
// 字段 Params 用 RawMessage 保留原始 JSON，由各 method 的 handler 自行解析。
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"` // 通知（notification）时为 nil，无需响应
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse 是 JSON-RPC 2.0 响应结构。
// 成功时填 Result，失败时填 Error；二者互斥。
// 通知请求不产生响应。
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError JSON-RPC 错误对象。
// code 遵循 MCP 约定：-32600 invalid request、-32601 method not found、-32602 invalid params、-32603 internal error。
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewErrorResponse 构造一个错误响应。
// 用于 method 不支持、参数校验失败等场景。
func NewErrorResponse(id json.RawMessage, code int, msg string) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCError{Code: code, Message: msg},
	}
}

// NewSuccessResponse 构造一个成功响应。
// result 由调用方提供具体 MCP 结构（InitializeResult / ToolsListResult / ToolsCallResult）。
func NewSuccessResponse(id json.RawMessage, result interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// ---- initialize 响应 ----

// InitializeParams initialize 请求参数（小智作为 Client 发送）。
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

// ClientInfo MCP 客户端标识。
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult initialize 响应的 result 字段。
// 声明本服务为 MCP Server，仅暴露 tools 能力。
type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      ServerInfo             `json:"serverInfo"`
}

// ServerInfo MCP 服务端标识。
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// NewInitializeResult 构造 initialize 成功响应体。
// protocolVersion 与小智平台对齐使用 "2024-11-05"。
func NewInitializeResult(serverName, serverVersion string) *InitializeResult {
	return &InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]interface{}{
			// 仅声明 tools 能力，不暴露 resources/prompts/logging。
			"tools": map[string]interface{}{},
		},
		ServerInfo: ServerInfo{Name: serverName, Version: serverVersion},
	}
}

// ---- tools/list 响应 ----

// ToolsListResult tools/list 响应的 result 字段。
type ToolsListResult struct {
	Tools []ToolDefinition `json:"tools"`
}

// ToolDefinition 单个工具的元数据与 inputSchema。
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// ---- tools/call 请求与响应 ----

// ToolsCallParams tools/call 请求参数。
type ToolsCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// ToolsCallResult tools/call 响应的 result 字段。
// 失败时 IsError=true 且 Content 携带错误说明文本；成功时 IsError 省略或 false。
type ToolsCallResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ContentItem MCP content 数组元素；当前仅支持 text 类型。
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NewTextCallResult 成功的工具调用结果，content 为单条 text。
func NewTextCallResult(text string) *ToolsCallResult {
	return &ToolsCallResult{
		Content: []ContentItem{{Type: "text", Text: text}},
	}
}

// NewErrorCallResult 失败的工具调用结果（isError=true），content 携带错误说明。
// 与 JSON-RPC level error 不同：这是 tool 执行层错误，仍以 200 result 返回，
// 由 MCP client 根据 isError 判定。
func NewErrorCallResult(msg string) *ToolsCallResult {
	return &ToolsCallResult{
		Content: []ContentItem{{Type: "text", Text: msg}},
		IsError: true,
	}
}

// ParseRequest 从原始 JSON 字节解析一条 JSON-RPC 请求。
// 返回的 req 可能为 notification（ID 为 nil），调用方据此决定是否回写响应。
func ParseRequest(raw []byte) (*JSONRPCRequest, error) {
	var req JSONRPCRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, fmt.Errorf("parse jsonrpc request: %w", err)
	}
	if req.JSONRPC == "" {
		// 兼容部分 client 未填 jsonrpc 字段的情况，默认补齐。
		req.JSONRPC = "2.0"
	}
	return &req, nil
}

// MarshalResponse 将响应序列化为 JSON 字节，用于写回 WebSocket。
func MarshalResponse(resp *JSONRPCResponse) ([]byte, error) {
	return json.Marshal(resp)
}
