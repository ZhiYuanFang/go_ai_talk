## Why

当前 `/voice/chat/ws` 对话链路使用百度流式 ASR（`dev_pid=1537` 近场模型），在**设备远场**场景下（老人、普通成人普通话）转写准确率不足，导致用户感知「说的和识别出来的不一致」。听写线 `/voice/asr/ws` 为**手机近场**场景且百度仍有免费额度，短期保持不变；**优先将对话场景的 STT 切换至阿里云百炼 `qwen-audio-3.0-asr-flash-streaming`**，以提升远场识别效果，同时复用生产环境已有 DashScope 密钥体系。

## What Changes

- 为 `/voice/chat/ws` 引入 **DashScope 流式 STT**（`qwen-audio-3.0-asr-flash-streaming`），替换对话链路的百度流式 ASR 主路径。
- 配置层拆分 **对话 STT**（`sttChat`）与 **听写 STT**（`sttDictation`）：对话走百炼，听写继续百度。
- 新增 `stt_stream_dashscope.go`，实现 `StreamASRSession`（WebSocket run-task / 二进制音频 / finish-task）。
- `CreateStreamASRSession` 增加 **profile** 参数（`chat` | `dictation`），由 `voice_ws.go` / `voice_asr_ws.go` 分别传入。
- 整段识别 fallback（`transcribe`）在对话 profile 下同样走百炼非实时或流式 commit 结果，听写 profile 仍走百度。
- 环境变量：`VOICE_DASHSCOPE_API_KEY`（可 fallback 至 `UCG_DASHSCOPE_API_KEY`）、`DASHSCOPE_WORKSPACE_ID`（WebSocket 端点）。
- **不改动**：`/voice/asr/ws` 听写协议与百度 STT 配置；TTS 仍用百度；客户端 WS 协议字段（`asr_partial` / `asr_final` 等）保持不变。

## Capabilities

### New Capabilities

- `voice-chat-dashscope-stt`：`/voice/chat/ws` 对话场景使用百炼流式 ASR 的 STT 选型、配置、profile 分发与降级语义。

### Modified Capabilities

- （无）听写线 `voice-realtime-asr-ws` 规格保持「百度流式 ASR」不变；本变更不修改其 Requirement。

## Impact

- **代码**：`internal/services/voice/`（`voice_chat.go`、`stt_stream_dashscope.go` 新增）、`internal/controller/voice_ws.go`、`internal/controller/voice_asr_ws.go`、`internal/services/contracts/runtime_contracts.go`（`CreateStreamASRSession` 签名扩展）。
- **配置**：`manifest/config/voice-chat.shared.yaml` 增加 `sttChat` / `sttDictation`（或等价嵌套）；`manifest/docker/env/.env.example` 增加 DashScope 相关 env。
- **依赖**：voice-service 出站 WebSocket 至阿里云百炼（`*.cn-beijing.maas.aliyuncs.com`）。
- **网关**：无新路由；`/voice/chat/ws` 透传行为不变。
- **usage 统计**：WebSocket 不计入（既有约定）。
- **Redis**：无新增读缓存。
- **风险**：百炼故障时对话 STT 不可用；设计保留百度 STT 代码作可选 fallback（配置开关，非默认）。
