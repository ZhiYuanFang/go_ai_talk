## Why

前端存在「仅需实时听写（语音→中文文本）」的场景，不应承担 `/voice/chat/ws` 的完整对话链路（DeepSeek 意图、TTS、`UpdateLastTalk`、单设备对话连接踢除等）。现有 chat WS 虽已下发 `asr_partial`/`asr_final`，但协议与状态机面向全链路对话，集成成本高且易误收 `audio_chunk` 等无关下行。

## What Changes

- 在 **voice-service** 新增独立 WebSocket 端点 **`/voice/asr/ws`**，仅承载流式 ASR 听写（复用百度流式 STT 与现有 PCM 约定）。
- 新增薄控制器（如 `voice_asr_ws.go`），**不**走 `VoiceWSManager`、**不**调用对话/TTS/设备最近对话落库。
- **gateway-service** 与 **gateway-app-server** 为 `/voice/asr/ws` 增加与 chat WS 同目标的边缘透传；App 网关将该路径加入 Bearer 白名单（与 `/voice/chat/ws` 一致）。
- 更新 README 中 WebSocket API 说明，区分「语音对话」与「实时听写」两条入口。
- **不**实现语种翻译（中→英等）、**不**修改 `/voice/chat/ws` 既有行为。

## Capabilities

### New Capabilities

- `voice-realtime-asr-ws`：定义 `/voice/asr/ws` 的握手、上行控制帧、二进制 PCM、下行 `asr_partial`/`asr_final`/`error` 等契约及 voice-service 边界约束。

### Modified Capabilities

- `gateway-ws-edge-proxy`：在现有 `/voice/chat/ws` 透传基础上，**ADDED** 对 `/voice/asr/ws` 的同配置透传要求（复用 `VOICE_WS_ROUTE_MODE` / `VOICE_WS_PROXY_URL`）。

## Impact

- **代码**：`internal/controller`（新 WS handler、注册、WS 代理中间件）、`internal/controller/register_voice_service.go`、`gateway_app_auth_exempt.go`、部署 manifest（compose/kustomize 一般无需新 env，复用现有 WS 代理变量）。
- **服务**：voice-service（领域执行）、gateway / gateway-app（边缘透传与鉴权白名单）。
- **前端**：新连 `ws(s)://<gateway>/voice/asr/ws`；协议精简，仅处理听写相关 JSON。
- **依赖**：沿用 `voice-chat.shared.yaml` 中 `stt.streamEnabled` 与百度流式 ASR；无新外部供应商。
