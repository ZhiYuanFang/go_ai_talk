## MODIFIED Requirements

### Requirement: 听写 WebSocket SHALL 使用百炼流式 STT（dictation profile）

`/voice/asr/ws` 流式 STT MUST 使用 **dictation profile** 配置（`sttDictation`；未配置时回退 legacy `stt` 块的既有规则仍适用），默认 provider MUST 为 `dashscope`，默认模型 MUST 为 `qwen-audio-3.0-asr-flash-streaming`（与 chat 路一致）。本变更 MUST NOT 改变听写端点的 WS 协议与事件类型（`asr_partial` / `asr_final` 由 client `commit`/`end` 触发定稿等）。

#### Scenario: 听写 WS 默认走 DashScope

- **WHEN** 客户端对 `/voice/asr/ws` 发送合法 `start`
- **AND** `sttDictation.provider=dashscope` 且 `streamEnabled=true`
- **AND** 有效 DashScope API Key 与 Workspace ID 已配置
- **THEN** 服务端 MUST 通过百炼 WebSocket 建立流式 ASR 会话（`CreateStreamASRSession` dictation profile）
- **AND** 下行 MUST 仍支持 `asr_partial`、`asr_final`（由 client commit/end 触发定稿）
- **AND** 服务端 MUST NOT 为该默认配置调用百度流式 ASR 建连

#### Scenario: 听写可配置回滚百度

- **WHEN** 运维将 `sttDictation.provider` 设为 `baidu` 且百度流式凭证可用
- **THEN** `CreateStreamASRSession(dictation)` MUST 创建百度流式 ASR 会话
- **AND** 听写 WS 协议 MUST 保持不变

### Requirement: CreateStreamASRSession SHALL 支持 profile 分流

`Voice().CreateStreamASRSession`（及 `VoiceContract` 等价接口）MUST 接受 `profile` 参数，取值 `chat` 或 `dictation`，并 MUST 根据 profile 选择 `sttChat` 或 `sttDictation` 配置块；实际引擎 MUST 由所选配置块的 `provider` 决定。

#### Scenario: chat 与 dictation 均按各自配置 provider 建连

- **WHEN** `voice_ws.go` 调用 `CreateStreamASRSession(..., profile=chat, ...)` 且 `sttChat.provider=dashscope`
- **THEN** 实现 MUST 读取 `sttChat` 并创建 DashScope 会话
- **WHEN** `voice_asr_ws.go` 调用 `CreateStreamASRSession(..., profile=dictation, ...)` 且 `sttDictation.provider=dashscope`
- **THEN** 实现 MUST 读取 `sttDictation` 并创建 DashScope 会话

#### Scenario: DashScope 降级百度使用 legacy stt

- **WHEN** 所选 profile 的 `provider=dashscope` 且建连失败
- **AND** 该配置块 `fallbackProvider=baidu`
- **THEN** 实现 MUST 使用 legacy `voiceChat.stt`（百度）配置创建百度流式会话
- **AND** MUST NOT 使用已为 dashscope 的 `sttDictation` 作为百度会话配置源

## ADDED Requirements

### Requirement: 听写 STT 配置 SHALL 位于 voice-chat.shared.yaml 的 sttDictation

听写 STT 配置 MUST 位于 `manifest/config/voice-chat.shared.yaml` 的 `voiceChat.sttDictation`（或经 `GF_VOICE_CHAT_FILE` 加载的等价文件）。默认仓库配置 MUST 将 `sttDictation.provider` 设为 `dashscope`。MUST NOT 将 DashScope STT 专属字段回流到 `manifest/config/config.yaml` 主网关配置。

#### Scenario: voice-service 加载 sttDictation

- **WHEN** `voice-service` 启动并加载 `voice-chat.shared.yaml`
- **THEN** `sttDictation.provider`、`sttDictation.model`、`sttDictation.streamEnabled` MUST 可供 `CreateStreamASRSession(dictation)` 读取
- **AND** 默认 `sttDictation.provider` MUST 为 `dashscope`
