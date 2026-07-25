## ADDED Requirements

### Requirement: Tip streaming generate API
Go 服务端 SHALL 暴露 `POST /device/tip/generate` 接口，以 SSE 方式流式返回小贴士生成的思考过程与最终建议。

#### Scenario: Flutter client requests tip via SSE
- **WHEN** Flutter 客户端向 `/device/tip/generate` 发送 POST 请求（body 包含 deviceNo、eventId、eventName；月龄与当前时间由服务端派生，不再要求 babyAgeMonths / currentTime）
- **THEN** Go 服务端 SHALL 立即返回 SSE 响应头（Content-Type: text/event-stream、Cache-Control: no-cache）
- **AND** 按序推送 `thinking`、`answer`、`done` 三类 SSE 事件
- **AND** 最终以 `data: [DONE]` 结束流式响应

### Requirement: Tip thinking events streamed
Go 服务端 SHALL 将 Python AI 返回的思考过程增量实时推送给客户端，不得等全部完成后一次性返回。

#### Scenario: Thinking delta pushed to client
- **WHEN** Go 服务端从 Python `/v1/tip/stream` 收到 `thinking` 类型的 SSE 事件
- **THEN** Go 服务端 SHALL 立即将该 thinking delta 以 SSE `event: thinking` 帧转发给 Flutter 客户端
- **AND** 不得缓冲或延迟超过 100ms

### Requirement: Tip answer events streamed
Go 服务端 SHALL 将 Python AI 返回的建议内容增量实时推送给客户端。

#### Scenario: Answer delta pushed to client
- **WHEN** Go 服务端从 Python `/v1/tip/stream` 收到 `answer` 类型的 SSE 事件
- **THEN** Go 服务端 SHALL 立即将该 answer delta 以 SSE `event: answer` 帧转发给 Flutter 客户端

### Requirement: Tip done event with answer_id
Go 服务端 SHALL 在流式结束时返回 answer_id，供后续反馈使用。

#### Scenario: Done event includes answer_id
- **WHEN** Go 服务端从 Python `/v1/tip/stream` 收到 `done` 事件（含 answer_id）
- **THEN** Go 服务端 SHALL 向客户端推送 SSE `event: done` 帧，body 含 answer_id
- **AND** 紧接着推送 `data: [DONE]` 结束整个 SSE 连接

### Requirement: Tip API error handling
当 Python AI 服务不可用时，Go 服务端 SHALL 返回固定错误提示给客户端。

#### Scenario: Python AI unavailable during tip generate
- **WHEN** Go 服务端调用 Python `/v1/tip/stream` 失败（连接超时、5xx 错误等）
- **THEN** Go 服务端 SHALL 以 SSE `event: error` 帧向客户端推送错误消息："AI服务暂时不可用，请稍后再试"
- **AND** 紧接着推送 `data: [DONE]` 结束连接
