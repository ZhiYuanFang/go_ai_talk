## ADDED Requirements

### Requirement: Chat stream pushes thinking then business reply only

文本对话流式路径（voice internal chat/stream 及经 history 委派的对外 chat/stream）MUST 在思考阶段仅推送 `event: thinking`；MUST 在意图落地（或 preamble 短路）后推送业务话术为 `event: answer`。MUST NOT 将 Python 意图分析 JSON 增量作为 `event: answer` 推送给客户端。

#### Scenario: Successful stream land with thinking

- **WHEN** 调用方发起流式文本对话且 Python `AnalyzeIntentStream` 产出 thinking 增量并成功落地得到业务 Reply
- **THEN** 服务端 SHALL 将 thinking 增量以 SSE `event: thinking` 推送
- **AND** SHALL 在落地完成后以 SSE `event: answer` 推送业务 Reply
- **AND** MUST NOT 推送意图结果 JSON 作为 answer

#### Scenario: Preamble short-circuit without Python stream

- **WHEN** preamble 短路（如 pending child）直接得到业务 Reply 且未调用 AnalyzeIntentStream
- **THEN** 服务端 MAY 不推送 thinking
- **AND** SHALL 以 SSE `event: answer` 推送业务 Reply

### Requirement: Intent stream callback has no OnAnswer

`contracts.IntentStreamCallback` MUST NOT 提供用于外泄意图 JSON 的 `OnAnswer` 回调；流式过程对外仅允许 `OnThinking`。Python 客户端内部累积意图 JSON 供解析 MUST 仍可用，但 MUST NOT 再转发到 `IntentStreamCallback`。

#### Scenario: Intent JSON not forwarded via callback

- **WHEN** `HandleTranscriptForIntentStream` 处理 AnalyzeIntentStream 的 answer（意图 JSON）事件
- **THEN** 系统 SHALL 在内部累积并解析意图结果
- **AND** MUST NOT 通过 `IntentStreamCallback` 将 JSON 增量回调给调用方

### Requirement: HandleTranscriptForIntentStream returns chat result

`HandleTranscriptForIntentStream` 成功或可降级返回时，返回结构 MUST 对齐聊天结果语义，至少包含 Ask、Reply、Exit、FinishTalk；MUST NOT 再要求调用方依赖 Thinking 或 AnswerJSON 字段消费 UI 内容（UI 内容由 thinking 回调与 Reply 承担）。

#### Scenario: Controller writes reply from return value

- **WHEN** voice internal ChatStream 调用 `HandleTranscriptForIntentStream` 结束且 Reply 非空
- **THEN** 控制器 SHALL 使用返回值中的 Reply 写入 SSE `event: answer`

#### Scenario: Stream failure surfaces error event

- **WHEN** 流式意图分析失败或业务错误返回
- **THEN** 控制器 SHALL 写入 SSE `event: error`
- **AND** 若返回结构仍含降级 Reply，MAY 同时写入 `event: answer`
