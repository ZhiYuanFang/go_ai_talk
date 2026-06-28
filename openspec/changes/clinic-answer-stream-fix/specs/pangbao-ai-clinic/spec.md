## MODIFIED Requirements

### Requirement: Clinic LLM stream SHALL emit answer_delta after thinking phase

当 Clinic 调用上游且 `ThinkingEnabled=true` 时，aimodel 流式层在已接收至少一个 reasoning/thinking 分片后，对上游单独到达的 `content` 分片 MUST 路由为 `answer`（触发 `OnAnswerDelta` 与 `answer_delta` WS 帧），MUST NOT 仅写入未订阅的 content 通道。

#### Scenario: reasoning 结束后 content 分片到达

- **WHEN** 上游先流式 `reasoning_content` 再流式纯 `content` 分片
- **THEN** voice-service MUST 向客户端发送 `answer_delta` 帧
- **AND** `answer_done` 的 `answer` 字段 MUST 非空（与上游正文一致）

#### Scenario: 非 thinking 闲聊路径不变

- **WHEN** `ThinkingEnabled=false` 且上游仅发送 `content`
- **THEN** 流式层 MUST 仍将 content 路由至 `OnContentDelta`（MUST NOT 误发 answer_delta）
