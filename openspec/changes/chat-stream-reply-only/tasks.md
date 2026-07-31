## 1. 契约与返回值

- [x] 1.1 从 `contracts.IntentStreamCallback` 删除 `OnAnswer`，更新注释（仅 OnThinking）
- [x] 1.2 瘦身 `IntentStreamResult`：删除 Thinking/AnswerJSON；`Answer` 改为 `Reply`（与 chatResult 对齐）
- [x] 1.3 更新 `VoiceContract` 上 `HandleTranscriptForIntentStream` 相关注释

## 2. Voice 服务编排

- [x] 2.1 `callDeepSeekUnifiedIntentStream`：不再将对外 cb 接到 `streamCb.OnAnswer`
- [x] 2.2 `HandleTranscriptForIntentStream`：所有返回路径改为填充 Ask/Reply/Exit/FinishTalk（preamble 短路、降级、成功落地）

## 3. Controller 与 history 委派

- [x] 3.1 `voice_internal_text_chat.ChatStream`：仅挂 OnThinking；结束后用 Reply 写 `event: answer`；错误写 error
- [x] 3.2 `DelegateTextChatStream`：`event: answer` 只累积返回值，不再调用 cb.OnAnswer
- [x] 3.3 `device_history.ChatStream`：仅转发 thinking；结束后用 delegate 返回的 reply 写 `event: answer`

## 4. 校验

- [x] 4.1 全仓检索确认无 `IntentStreamCallback{...OnAnswer` 残留，且无对已删字段的引用
- [x] 4.2 `go build` 相关包（voice-service / history 相关）通过
