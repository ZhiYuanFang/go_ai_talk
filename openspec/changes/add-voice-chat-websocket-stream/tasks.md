## 1. Implementation
- [x] 1.1 增加 WebSocket 语音接口 `/voice/chat/ws`，支持二进制音频帧接收与回传。
- [x] 1.2 从连接头读取 `X-Device-No` 与 `X-Audio-*`，缺失或非法时立即拒绝。
- [x] 1.3 增加按设备号的 WebSocket 连接管理（注册、替换、注销）。
- [x] 1.4 改造 DeepSeek 调用支持 `stream=true` 并解析 SSE 增量回复。
- [x] 1.5 更新配置与文档，说明流式模式与设备头要求。
- [x] 1.6 支持客户端发送文本控制帧 `{"type":"end"}` 显式结束当前音频片段，并立即触发识别/回复收尾。
- [x] 1.7 调整 WebSocket 协议为 `start`（文本）+ BIN 分片 + `end`（文本），并以 `result` 文本消息返回音频结果。
- [x] 1.8 保持无额外鉴权头要求，允许并忽略 `X-Device-No` 请求头，错误时返回可读 `error` 消息。
- [x] 1.9 避免单条超大 `result.audio`，改为连续发送 `audio_chunk`（Base64片段）并以 `audio_end` 结束。
- [x] 1.10 在流式会话中，若从首次有效音频开始连续2秒未再收到有效音频，则主动触发 commit，并向前端发送 `interrupt_commit` 以结束录音阶段。

## 2. Validation
- [x] 2.1 运行 `go test ./...` 并确认与本次改动相关模块通过。
