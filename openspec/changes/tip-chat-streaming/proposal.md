## Why

上一个需求（fix-python-api-alignment）已完成 Python 流式接口对齐与 Go 侧流式核心逻辑封装（`TipStream`、`AnalyzeIntentStream`、`HandleTranscriptForIntentStream`），但这些能力目前只停留在 voice 进程内部，没有暴露成对外 HTTP 接口供 Flutter 客户端和 MCP 调用方使用。导致：1）小贴士 Flutter 端 UI 已完成但无法从 Go 服务获取流式结果；2）语音球转文本后只能同步一次性返回回复，无法流式展示思考过程。

## What Changes

- 新增 Go 侧 `/device/tip/generate` SSE 接口，内部调用已有的 `PythonAIClient.TipStream()`，向 Flutter 透传 thinking/answer/done 流式事件
- 新增 Go 侧 `/device/history/api/chat/stream` SSE 接口，经由 history → voice internal 链路透传，内部调用已有的 `HandleTranscriptForIntentStream()`
- 新增 voice internal 流式接口 `/voice/internal/api/text/chat/stream`，供 history 与 mcpbridge 进程委派
- 新增 history 侧 `DelegateTextChatStream()` 流式委派方法
- Flutter 端 `sendCommand()` 由同步 HTTP 改为 SSE 流式调用 `/chat/stream`
- Flutter 端首页新增 `_chatThinking` 状态，语音消息条优先展示 thinking 过程，answer 到达后切换展示最终回复

## Capabilities

### New Capabilities

- `tip-streaming`: 小贴士流式生成接口，Flutter 客户端通过 SSE 获取思考过程与最终建议文案
- `chat-streaming`: 纯文本对话流式接口，支持语音球文本与 MCP 调用方以 SSE 方式获取思考过程与最终回复

### Modified Capabilities

无

## Impact

- Go 侧新增 3 个 HTTP 接口定义 + 4 个控制器/委派方法，均为纯新增无破坏性修改
- Flutter 侧 `sendCommand()` 返回值由 `Future<String?>` 改为 `Stream<String>`，调用方需适配
- 复用已有的 `PythonAIClient.TipStream()`、`HandleTranscriptForIntentStream()`、`IntentStreamCallback` 等核心逻辑，无重复开发
- 现有同步接口 `/device/history/api/chat` 保持不变，作为兜底与非流式场景使用
