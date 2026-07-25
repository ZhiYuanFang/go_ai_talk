## 1. P0 小贴士流式接口（Go 侧）

- [x] 1.1 在 `api/v1/` 新增 `device_tip_http.go`，定义 `DeviceTipGenerateReq`（path:`/device/tip/generate`, method:post）与 SSE 响应结构
- [x] 1.2 在 `internal/controller/` 新增 `device_tip.go`，实现 `TipGenerate()` 控制器：设置 SSE 响应头、调用 `PythonAIClient.TipStream()`、逐帧转发 thinking/answer/done 事件
- [x] 1.3 在 `internal/services/voice/voice_chat.go` 新增 `TipStream()` 方法，封装 `PythonAIClient.TipStream()` 供控制器调用
- [x] 1.4 验证 Go 编译通过，小贴士 SSE 接口可通过 curl 测试返回流式事件

## 2. P1 语音球流式文案 - Go 侧接口与委派

- [x] 2.1 在 `api/v1/device_history_http.go` 新增 `DeviceHistoryChatStreamReq`（path:`/device/history/api/chat/stream`, method:post）
- [x] 2.2 在 `internal/controller/device_history.go` 新增 `ChatStream()` 方法：SSE 响应头 + 调用 `DelegateTextChatStream()` + 逐帧转发
- [x] 2.3 在 `internal/services/history/delegate_http.go` 新增 `DelegateTextChatStream()`：以 HTTP SSE 客户端调用 voice internal 接口
- [x] 2.4 在 `internal/controller/voice_internal_text_chat.go` 新增 internal 流式接口 `/voice/internal/api/text/chat/stream`：调用 `HandleTranscriptForIntentStream()` + SSE 写回
- [x] 2.5 验证 Go 编译通过，chat/stream 接口链路可端到端返回流式 thinking/answer

## 3. P1 语音球流式文案 - Flutter 侧改造

- [ ] 3.1 在 `remote_feed_repository.dart` 改造 `sendCommand()`：由 `Future<String?>` 改为 `Stream<String>`，建立 SSE 连接到 `/chat/stream`，按事件顺序产出 thinking/answer
- [ ] 3.2 在 `home_screen.dart` 新增 `_chatThinking` 状态，`_voiceStripText` getter 中优先展示 thinking（answer 到达后切换）
- [ ] 3.3 验证 Flutter 编译通过，语音球发送文本后能实时看到 thinking 过程并最终切换到 answer
