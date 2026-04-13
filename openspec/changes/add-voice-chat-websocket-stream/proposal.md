# Change: 语音智能对话改为流式传输并按设备管理 WebSocket

## Why
- 当前语音对话为单次 HTTP 请求，不满足实时连续传输诉求。
- 设备端会在连接头中携带设备信息，需要服务端按设备隔离接收与管理音频流。
- 智能对话链路需要支持流式响应，以降低用户等待感知。

## What Changes
- 新增 WebSocket 语音对话入口 `/voice/chat/ws`，在 9701 端口执行标准 HTTP Upgrade（101 Switching Protocols）。
- WebSocket 会话协议改为：首条文本 `start`（含 `deviceNo` 与音频参数）→ 后续二进制音频分片 → 文本 `end` 触发收尾处理。
- 不要求额外鉴权头；允许 `ws://`；允许并忽略自定义请求头 `X-Device-No`。
- 同设备新连接建立时替换旧连接，避免多连接并发导致音频路由混乱。
- 客户端可发送文本控制帧 `{"type":"end"}` 显式结束当前音频片段，服务端立即执行识别/回复收尾并返回结果。
- 服务端返回可读文本结果消息（`type=result`）或可读错误消息，不直接无提示断连。
- DeepSeek 对话请求支持流式模式（SSE），服务端聚合增量内容后进入 TTS。

## Impact
- Affected specs: audio-chat
- Affected code:
  - `internal/cmd/cmd.go`
  - `internal/service/voice_ws_manager.go`
  - `internal/service/voice_chat.go`
  - `manifest/config/config.yaml`
  - `README.MD`
