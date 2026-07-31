## Why

文本对话流式路径把 Python 意图分析的 JSON 增量经 `OnAnswer` 推给客户端，与落地后的业务话术共用 `event: answer`，导致 UI 可能展示意图 JSON，且 `HandleTranscriptForIntentStream` 返回值混杂 Thinking/AnswerJSON 与聊天结果。客户端真正需要的是：思考过程推 thinking，结束后推业务 Reply。

## What Changes

- **BREAKING（内部 SSE 语义）**：chat/stream 路径不再转发意图 JSON 为 `event: answer`；`answer` 仅承载落地后的业务话术（`chatResult.Reply`）。
- 从 `contracts.IntentStreamCallback` 移除 `OnAnswer`；流式过程仅保留 `OnThinking`。
- `HandleTranscriptForIntentStream` 返回值对齐聊天结果语义（Ask/Reply/Exit/FinishTalk），不再对外暴露 Thinking/AnswerJSON。
- `voice_internal` ChatStream：thinking 走回调；结束后写一帧 `event: answer` = Reply。
- history 委派/转发随 voice SSE 语义自动对齐；Tip 流与 Clinic 路径不在本变更范围。

## Capabilities

### New Capabilities

- `chat-stream-reply-only`：文本对话流式 SSE 仅推 thinking + 业务 Reply；意图 JSON 仅作内部落库原料。

### Modified Capabilities

- （无）主规格目录无独立 chat-streaming capability；既有行为由 `tip-chat-streaming` change 内 delta 描述，本变更以新 capability 覆盖并纠正其 chat 路径假设。

## Impact

- 代码：`runtime_contracts.go`、`voice_chat.go`、`voice_chat_understanding.go`、`voice_internal_text_chat.go`；可能微调 `device_history.go` / `delegate_http.go` 注释或回调接线。
- API：内部 `/voice/internal/.../chat/stream` 与对外 `/device/history/api/chat/stream` 的 SSE `answer` 语义变更（由意图 JSON → 业务话术）。
- 不动：Python `/v1/analyze/intent/stream`、`AnalyzeIntentStream` 内部 JSON 累积、`TipStreamCallback`、Clinic WS。
- Flutter：未完成的语音球流式接入将按「thinking → 最终回复」消费，不再收到意图 JSON。
