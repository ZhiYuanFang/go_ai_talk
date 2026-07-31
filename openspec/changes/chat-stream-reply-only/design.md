## Context

文本对话流式链路：

```
Flutter/App → history ChatStream → DelegateTextChatStream → voice ChatStream
  → HandleTranscriptForIntentStream → AnalyzeIntentStream(Python)
  → applyUnifiedIntentResult → 业务 Reply
```

现状问题：
1. `IntentStreamCallback.OnAnswer` 把 Python 意图 JSON 增量推到 SSE `event: answer`。
2. 落地后 controller 再写一帧 `chatRes.Answer`（业务话术），同名事件语义冲突。
3. 返回值 `IntentStreamResult` 混杂 Thinking/AnswerJSON 与 Ask/Answer/Exit/FinishTalk；业务终点实际是包内 `chatResult{Reply,...}`。

约束：voice 不直连他域 DAO；不改 Python 契约；Tip/Clinic 独立回调类型；不加测试文件。

## Goals / Non-Goals

**Goals:**

- 流式过程仅推 `thinking`（用户可见思考话术）。
- 意图落地后推一次（或等价）`answer` = 业务 `Reply`。
- 对外回调去掉 `OnAnswer`；Python 客户端仍内部累积 JSON 供解析。
- `HandleTranscriptForIntentStream` 返回聊天结果语义（与 `chatResult` 对齐）。

**Non-Goals:**

- 不改 Python `/v1/analyze/intent/stream`。
- 不改 `TipStreamCallback` / tip SSE / Clinic WS。
- 不在本变更实现 Flutter UI。
- 不新增 `Exit`/`FinishTalk` 的 SSE 事件（保留在返回结构，供进程内调用方）。
- 不改非流式 `TextChat` / TTS 路径。

## Decisions

### 1. 契约：`IntentStreamCallback` 仅保留 OnThinking

**决策**：删除 `contracts.IntentStreamCallback.OnAnswer`；注释标明意图 JSON 不外泄。

**备选**：保留字段但调用方传 nil → 易误用，拒绝。

### 2. Python 客户端内部仍可有 AnalyzeIntentStreamCallback.OnAnswer

**决策**：`python_ai_client.AnalyzeIntentStream` 继续本地 `answer += content` 并解析 Result；`callDeepSeekUnifiedIntentStream` **不再**把对外 cb 接到 `streamCb.OnAnswer`。

**理由**：解析依赖完整 JSON；外泄才是 bug。

### 3. 返回值对齐 chatResult

**决策**：将 `IntentStreamResult` 瘦身为聊天结果字段：`Ask`、`Reply`（或保留 `Answer` 但语义=Reply）、`Exit`、`FinishTalk`；删除 `Thinking`、`AnswerJSON`。字段名优先用 `Reply` 与包内 `chatResult` 一致；若为减少调用方改动可暂留 `Answer` 别名，本变更统一改为 `Reply`。

**备选**：返回包内 `chatResult` → 无法跨 contracts 导出未导出类型；拒绝。

### 4. Controller 写入时机

**决策**：`voice_internal` ChatStream：
1. 设置 SSE 头；
2. 调用 `HandleTranscriptForIntentStream`，仅挂 `OnThinking` → `event: thinking`；
3. 成功且 `Reply` 非空 → `event: answer` = Reply；
4. 错误 → `event: error`（降级文案）；有降级 Reply 时仍可写 answer；
5. `data: [DONE]`。

preamble 短路（无 Stream）同样只写最终 answer，无 thinking。

### 5. history 侧

**决策**：`DelegateTextChatStream` 继续按 SSE 解析 `thinking`/`answer`；因 voice 不再推 JSON，`answer` 累积即为业务话术。`device_history.ChatStream` 去掉对已删除 `OnAnswer` 的服务回调依赖——若仍用 IntentStreamCallback 转发线协议 answer，改为在解析层转发；最小改动：保留对 SSE `answer` 事件的转发逻辑（delegate 内），history controller 的 callback 仅需 OnThinking 若 answer 已由 voice 写入且经 delegate 转发。

实际：history → HTTP SSE 读 voice → 回调 OnThinking/OnAnswer 转写给客户端。`IntentStreamCallback` 删 OnAnswer 后，delegate 对 `event: answer` 的处理改为：累积返回值 + 若需实时转发，可用独立参数或保留一个 `onAnswerEvent`。最小方案：

- `IntentStreamCallback` 仅 OnThinking；
- `DelegateTextChatStream` 对 `event: answer` 只累积到返回 `string`，不再要求 cb.OnAnswer；
- `device_history.ChatStream`：OnThinking 转发；流结束后用 delegate 返回的 reply 写一帧 `event: answer`（与 voice 可能重复？）

更干净：voice 已写 answer 进 SSE；delegate 解析时把 answer 事件也转发给客户端。若 callback 无 OnAnswer，可：
- 给 Delegate 增加可选 `onAnswer func(string) error`，或
- history 在结束后用返回 string 写 answer。

**采用**：history ChatStream 在 `DelegateTextChatStream` 返回后，若 reply 非空则 `writeSSEEvent(rw, "answer", reply)`；流式中只转发 thinking。注意：voice 响应体里已有 answer 帧，delegate 会读到并拼进返回值；history 再写一帧给 App——**不要**在 history 解析时又实时转发 answer（避免与「结束后写一次」重复）。Delegate 内部对 answer 只累积、不回调。

若 history 直接把 voice SSE 透传会更简单，但现网是解析后重写。保持「thinking 实时 + 结束写 reply」一致于 voice_internal。

### 6. 错误与降级

**决策**：
- Stream/业务错误：写 `event: error`；若返回结构仍带降级 Reply，可再写 answer。
- Result 为空：服务返回降级 Reply + error；controller 两者都写（与现网接近）。

## Risks / Trade-offs

- **[Risk] Flutter 若已按「answer=JSON」联调** → Mitigation：Flutter Phase 3 未完成；文档/spec 明确 answer=业务话术。
- **[Risk] tip-chat-streaming 旧 spec 仍写 answer delta=意图** → Mitigation：本 capability 覆盖纠正；不强制改已归档/进行中 change 目录，实现以本 change 为准。
- **[Trade-off] answer 非增量流式** → 业务话术在落库后才可得，无法增量推 Reply；可接受。
- **[Risk] IntentStreamResult 字段改名破坏编译** → Mitigation：全仓 grep 更新调用方（仅 controller 等少数点）。

## Migration Plan

- 随 voice-service + history-service 同发；先发 voice 后发 history 时，旧 history 若仍设 OnAnswer 会编译失败（同仓），故单体发布。
- 回滚：恢复 OnAnswer 转发与 IntentStreamResult 旧字段（会恢复 JSON 外泄）。

## Open Questions

- 无阻塞。Exit/FinishTalk 进 SSE 留待后续。
