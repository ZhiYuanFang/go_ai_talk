## ADDED Requirements

### Requirement: 对话 WebSocket SHALL 使用百炼流式 STT（chat profile）

`voice-service` 在处理 `/voice/chat/ws` 且输入模态为 `audio` 时，流式 STT MUST 使用 **chat profile** 配置（`sttChat`），provider MUST 为 `dashscope`，默认模型 MUST 为 `qwen-audio-3.0-asr-flash-streaming`。MUST NOT 在该路径默认使用百度 `dev_pid=1537` 流式 ASR。

#### Scenario: 对话 WS 建连后创建 DashScope 流式 ASR

- **WHEN** 客户端对 `/voice/chat/ws` 发送合法 `start` 且 `inputModality=audio`（或省略默认为 audio）
- **AND** `sttChat.provider=dashscope` 且 `sttChat.streamEnabled=true`
- **AND** 有效 DashScope API Key 与 Workspace ID 已配置
- **THEN** 服务端 MUST 通过百炼 WebSocket 建立流式 ASR 会话
- **AND** 服务端 MUST NOT 为该会话调用百度流式 ASR 建连

#### Scenario: DashScope 凭证缺失

- **WHEN** 客户端对 `/voice/chat/ws` 开始 audio 会话
- **AND** chat profile 的 API Key 或 Workspace ID 未配置
- **THEN** 服务端 MUST 返回 `error` 帧且 `stage=stt`
- **AND** MUST NOT 静默降级为无 ASR 继续对话

### Requirement: 听写 WebSocket SHALL 继续使用百度 STT（dictation profile）

`/voice/asr/ws` 流式 STT MUST 使用 **dictation profile** 配置（`sttDictation` 或与现有 `stt` 块等价），provider MUST 保持 `baidu`。本变更 MUST NOT 改变听写端点的 WS 协议与事件类型。

#### Scenario: 听写 WS 仍走百度

- **WHEN** 客户端对 `/voice/asr/ws` 发送合法 `start`
- **AND** `sttDictation.provider=baidu` 且 `streamEnabled=true`
- **THEN** 服务端 MUST 使用百度流式 ASR（`CreateStreamASRSession` dictation profile）
- **AND** 下行 MUST 仍支持 `asr_partial`、`asr_final`（由 client commit/end 触发定稿）

### Requirement: CreateStreamASRSession SHALL 支持 profile 分流

`Voice().CreateStreamASRSession`（及 `VoiceContract` 等价接口）MUST 接受 `profile` 参数，取值 `chat` 或 `dictation`，并 MUST 根据 profile 选择 `sttChat` 或 `sttDictation` 配置块。

#### Scenario: chat 与 dictation 调用不同 provider

- **WHEN** `voice_ws.go` 调用 `CreateStreamASRSession(..., profile=chat, ...)`
- **THEN** 实现 MUST 读取 `sttChat` 并创建 DashScope 会话
- **WHEN** `voice_asr_ws.go` 调用 `CreateStreamASRSession(..., profile=dictation, ...)`
- **THEN** 实现 MUST 读取 `sttDictation` 并创建百度会话

### Requirement: DashScope 流式 ASR SHALL 兼容现有 PCM 参数

百炼流式 ASR 实现 MUST 接受与现有链路一致的音频元数据：`sampleRate=16000`（或 `start` 声明值）、`bits=16`、`channels=1`、PCM 小端二进制分片。`WriteAudio` / `Commit` / `Close` 语义 MUST 与 `StreamASRSession` 契约一致，以便 `/voice/chat/ws` 无需修改客户端。

#### Scenario: 对话 WS 收到 asr_partial

- **WHEN** DashScope 返回中间识别结果且文本非空
- **THEN** `/voice/chat/ws` MUST 向客户端发送 `{"type":"asr_partial","code":0,"text":"..."}`（与现有行为一致）

#### Scenario: 对话 WS commit 后 asr_final

- **WHEN** 服务端对 DashScope 会话执行 `Commit` 且得到有效转写文本
- **THEN** `/voice/chat/ws` MUST 继续现有对话链路（LLM/TTS），且 MUST 下发 `asr_final` 或等价最终文本事件（与现实现一致）

### Requirement: 对话 STT 配置 SHALL 位于 voice-chat.shared.yaml

对话 STT 配置 MUST 位于 `manifest/config/voice-chat.shared.yaml` 的 `voiceChat.sttChat`（或经 `GF_VOICE_CHAT_FILE` 加载的等价文件）。MUST NOT 将 DashScope STT 专属字段回流到 `manifest/config/config.yaml` 主网关配置。

#### Scenario: voice-service 加载 sttChat

- **WHEN** `voice-service` 启动并加载 `voice-chat.shared.yaml`
- **THEN** `sttChat.provider`、`sttChat.model`、`sttChat.streamEnabled` MUST 可供 `CreateStreamASRSession(chat)` 读取

### Requirement: 环境变量 SHALL 支持 DashScope 凭证注入

部署 MUST 支持经环境变量注入百炼凭证：`VOICE_DASHSCOPE_API_KEY`（优先）或复用 `UCG_DASHSCOPE_API_KEY`；`DASHSCOPE_WORKSPACE_ID` 用于 WebSocket 端点。`manifest/docker/.env.example` MUST 文档化上述变量。

#### Scenario: 仅配置 UCG 密钥

- **WHEN** `VOICE_DASHSCOPE_API_KEY` 未设置且 `UCG_DASHSCOPE_API_KEY` 已设置
- **AND** `sttChat.apiKey` 配置为空
- **THEN** chat profile STT MUST 使用 `UCG_DASHSCOPE_API_KEY` 鉴权
