## 1. Voice-service 听写 WebSocket

- [x] 1.1 新增 `internal/controller/voice_asr_ws.go`：实现 `/voice/asr/ws` 读循环（`start`/`end`/`ping`、二进制 PCM、`asr_partial`/`asr_final`/`error`/`started`/`ended`）
- [x] 1.2 复用 `CreateStreamASRSession` 与必要静音 finalize 逻辑；确认不调用 `HandleTranscriptForStreaming`、TTS、`UpdateLastTalk`、`VoiceWSManager`
- [x] 1.3 在 `register_voice_service.go`（或同级注册函数）中注册 `registerVoiceAsrWS`
- [x] 1.4 固定 `commit` 行为（实现 `unsupported` 或可选手动截句）并在代码注释中说明

## 2. Gateway 透传与 App 白名单

- [x] 2.1 扩展 `installVoiceWSProxyMiddleware`（或等价）使 `/voice/asr/ws` 与 `/voice/chat/ws` 共用 `VOICE_WS_ROUTE_MODE` / `VOICE_WS_PROXY_URL`
- [x] 2.2 在 `gateway_app_auth_exempt.go` 将 `/voice/asr/ws` 加入 `gatewayAppAuthExemptExactGET`
- [x] 2.3 本地验证：gateway `:9701` 与 gateway-app `:9702` 均可 Upgrade 并透传到 voice-service `:9802`

## 3. 文档与可观测性

- [x] 3.1 更新 `README.MD`：新增「实时听写 WebSocket」小节（URL、start/end、下行类型、与 chat WS 对比）
- [x] 3.2 关键路径补充中文注释（为何不用 VoiceWSManager、为何不落库）
- [x] 3.3 运行 `go build ./...` 确认编译通过

## 4. 归档前自检（实现完成后）

- [x] 4.1 对照 `specs/voice-realtime-asr-ws/spec.md` 与 `specs/gateway-ws-edge-proxy/spec.md` 场景逐项勾选
- [x] 4.2 确认未新增跨域 DAO import、未修改 `/voice/chat/ws` 对外行为
