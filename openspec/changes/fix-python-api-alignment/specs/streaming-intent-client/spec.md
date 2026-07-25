## ADDED Requirements

### Requirement: Go 侧 PythonAIClient 支持流式意图分析
PythonAIClient SHALL 提供 `AnalyzeIntentStream` 方法，以 SSE 流式方式调用 Python 服务 `/v1/analyze/intent/stream` 接口。

#### Scenario: 流式意图分析请求格式
- **WHEN** Go 侧调用 `AnalyzeIntentStream`
- **THEN** 请求 SHALL 以 `POST /v1/analyze/intent/stream` 发送，Content-Type 为 `application/json`，Accept 为 `text/event-stream`
- **AND** 请求体 SHALL 包含 `text`、`device_no`、`model` 字段，与非流式接口一致

#### Scenario: 流式响应 SSE 解析
- **WHEN** Python 服务返回 SSE 流式响应
- **THEN** Go 侧 SHALL 逐行解析以 `data: ` 开头的 SSE 事件
- **AND** 事件类型为 `thinking` 时 SHALL 触发 `OnThinking` 回调并累积思考内容
- **AND** 事件类型为 `answer` 时 SHALL 触发 `OnAnswer` 回调并累积回答内容（answer 事件的 content 是 JSON 格式的意图结果）
- **AND** 事件类型为 `done` 或 `[DONE]` 时 SHALL 结束解析
