## Why

对话链路 `/voice/chat/ws` 已切百炼 DashScope 流式 STT，听写 `/voice/asr/ws` 仍默认百度，两套引擎增加运维与质量差异。希望听写与对话统一使用百炼 ASR，同时保持听写 WS 协议与「前端 commit 定稿」语义不变。

## What Changes

- 将 `voiceChat.sttDictation.provider` 默认改为 `dashscope`，模型与凭证解析与 `sttChat` 对齐（配置切换路径 A）。
- 修订基线规格：撤销「听写 MUST 保持百度」约束，改为听写默认百炼；WS 协议与事件类型不变。
- 修正 DashScope 建连失败降级百度时的配置来源：MUST 使用 legacy `stt`（百度）块，禁止再读已改为 dashscope 的 `sttDictation`。
- 同步注释与发版 runbook（听写亦依赖 Workspace / API Key）。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `voice-chat-dashscope-stt`：听写 profile 默认 provider 改为 dashscope；profile 分流场景与降级配置来源更新。
- `voice-realtime-asr-ws`：听写端点描述由「当前为百度」改为「由 `sttDictation` 配置的流式 STT（默认 dashscope）」。

## Impact

- **配置**：`manifest/config/voice-chat.shared.yaml` 的 `sttDictation`。
- **voice-service**：`CreateStreamASRSession` 降级分支；注释/说明；无新 WS 路径、无新 DB/Redis。
- **部署**：听写与对话共用 `VOICE_DASHSCOPE_API_KEY` / `UCG_DASHSCOPE_API_KEY` 与 `DASHSCOPE_WORKSPACE_ID`。
- **客户端**：`/voice/asr/ws` 协议不变（非 BREAKING）；识别效果可能随引擎变化。
- **回滚**：将 `sttDictation.provider` 改回 `baidu` 并恢复百度相关字段后重启 voice-service。
