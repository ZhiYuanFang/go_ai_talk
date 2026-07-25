## ADDED Requirements

### Requirement: Chat streaming API
Go 服务端 SHALL 暴露 `POST /device/history/api/chat/stream` 接口，以 SSE 方式流式返回文本对话的思考过程与最终回复。

#### Scenario: Flutter client requests chat via SSE
- **WHEN** Flutter 客户端向 `/device/history/api/chat/stream` 发送 POST 请求（body 包含 deviceNo、transcript）
- **THEN** Go 服务端 SHALL 立即返回 SSE 响应头（Content-Type: text/event-stream、Cache-Control: no-cache）
- **AND** 按序推送 `thinking`、`answer` 两类 SSE 事件
- **AND** 最终以 `data: [DONE]` 结束流式响应

### Requirement: Chat thinking events streamed
Go 服务端 SHALL 将 AI 思考过程增量实时推送给客户端。

#### Scenario: Chat thinking delta pushed to client
- **WHEN** Go 服务端从 voice 服务流式意图分析收到 thinking delta 回调
- **THEN** Go 服务端 SHALL 立即将该 thinking delta 以 SSE `event: thinking` 帧转发给客户端
- **AND** 不得缓冲或延迟超过 100ms

### Requirement: Chat answer events streamed
Go 服务端 SHALL 将 AI 最终回复内容增量实时推送给客户端。

#### Scenario: Chat answer delta pushed to client
- **WHEN** Go 服务端从 voice 服务流式意图分析收到 answer delta 回调
- **THEN** Go 服务端 SHALL 立即将该 answer delta 以 SSE `event: answer` 帧转发给客户端

### Requirement: Chat streaming via history delegate
history 服务 SHALL 以 HTTP SSE 客户端方式委派流式对话请求到 voice 服务。

#### Scenario: History service delegates streaming chat to voice
- **WHEN** history 服务的 `DelegateTextChatStream()` 被调用
- **THEN** history 服务 SHALL 向 voice 服务的 internal 接口 `/voice/internal/api/text/chat/stream` 发起 SSE 请求
- **AND** 将 voice 服务返回的 SSE 事件逐帧透传给调用方

### Requirement: Voice internal streaming chat API
voice 服务 SHALL 暴露 internal SSE 接口 `/voice/internal/api/text/chat/stream`，供跨进程委派使用。

#### Scenario: Voice internal chat streaming endpoint
- **WHEN** 经内部认证的请求到达 `/voice/internal/api/text/chat/stream`
- **THEN** voice 服务 SHALL 调用 `HandleTranscriptForIntentStream()` 执行流式意图分析
- **AND** 将 OnThinking/OnAnswer 回调以 SSE 帧写入 HTTP 响应
- **AND** 流式结束后推送 `data: [DONE]`

### Requirement: Chat stream preserves sync API compatibility
原同步接口 `/device/history/api/chat` SHALL 保持完全不变，继续提供一次性返回 reply 的能力。

#### Scenario: Sync chat API still works
- **WHEN** 客户端调用原同步接口 `/device/history/api/chat`
- **THEN** Go 服务端 SHALL 以与变更前完全一致的同步方式返回 `{reply: "..."}`

### Requirement: Chat API error handling
当 AI 服务不可用时，Go 服务端 SHALL 返回固定错误提示。

#### Scenario: AI unavailable during streaming chat
- **WHEN** 流式对话过程中 AI 服务调用失败
- **THEN** Go 服务端 SHALL 以 SSE `event: error` 帧推送错误消息："AI服务暂时不可用，请稍后再试"
- **AND** 紧接着推送 `data: [DONE]` 结束连接

### Requirement: Flutter sendCommand uses streaming
Flutter 端 `sendCommand()` SHALL 改为 SSE 流式调用 `/chat/stream` 接口。

#### Scenario: Flutter streaming chat display
- **WHEN** Flutter 调用 `sendCommand(text)`
- **THEN** Flutter SHALL 建立 SSE 连接到 `/device/history/api/chat/stream`
- **AND** 收到 thinking 事件时更新 `_chatThinking` 状态并在消息条展示
- **AND** 收到 answer 事件时累积 `_chatReply` 并切换展示最终回复
- **AND** 收到 `data: [DONE]` 时关闭 SSE 连接
