## ADDED Requirements

### Requirement: chat WS 支持输入输出模态组合

`/voice/chat/ws` MUST 在保留现有音频流式能力的基础上，支持自然语言喂养的四种 I/O 组合：输入 `audio|text`，输出 `audio|text`。模态 MUST 由客户端在 `start` 帧中声明（字段名实现可选 `inputModality`/`outputModality` 或等价约定）；缺省 MUST 为 `audio`/`audio`，以保持横屏语音兼容。

#### Scenario: 默认音入音出兼容横屏

- **WHEN** 客户端发送既有横屏 `start`（未声明模态或显式 audio/audio）并上传 PCM
- **THEN** 服务端 MUST 按现网行为进行 ASR、意图流式、TTS，并下发 `thinking_delta`/`answer`/`audio_*`（及既有 asr 帧）

#### Scenario: 文入文出

- **WHEN** 客户端 `start` 声明 text 输入与 text 输出，随后发送 text 上行帧（非空文本）
- **THEN** 服务端 MUST NOT 依赖 PCM/ASR 完成该轮，MUST 下发 `thinking_delta` 与 `answer`，MUST NOT 下发 TTS `audio_chunk`/`audio_end`

#### Scenario: 文入音出

- **WHEN** 客户端声明 text 输入与 audio 输出并发送 text 帧
- **THEN** 服务端 MUST 跳过 ASR，MUST 在意图完成后执行 TTS 并下发音频帧

#### Scenario: 音入文出

- **WHEN** 客户端声明 audio 输入与 text 输出并完成一轮 ASR commit
- **THEN** 服务端 MUST 下发思考与 answer，MUST NOT 下发 TTS 音频帧

### Requirement: 文模式 start 仍校验音频元数据

纯 text 输入时，`start` MUST 仍要求与现网一致的音频元数据字段（含 `sampleRate`、`bits`、`channels`、`length`、`mode=stream` 等既有必填项）校验通过；服务端 MUST NOT 因文模式放宽或省略这些字段的必填校验。客户端 MAY 传入与横屏一致的占位数值。

#### Scenario: 缺 sampleRate 的文模式 start 被拒

- **WHEN** 客户端声明 text 输入但 `start` 缺少合法 `sampleRate`（或既有必填音频字段）
- **THEN** 服务端 MUST 拒绝该 `start`（错误帧或等价失败），MUST NOT 进入 `started` 可对话状态

### Requirement: 喂养对话经硬件特权与 VU 正式模

经 `/voice/chat/ws` 触发的喂养意图 LLM MUST 标记硬件特权，MUST 使用 `voiceUnderstanding` lane 的正式（非 free）模型配置，MUST 经该 lane 并发闸门 `Acquire`。该路径 MUST NOT 因 `voice_ai` 月度额度用尽而拒绝对话或改走 free 模型。

#### Scenario: 不计 voice_ai 次

- **WHEN** 一轮 WS 喂养意图 LLM 成功完成（含文模式）
- **THEN** 系统 MUST NOT 对该轮执行 `voice_ai` 成功计次（consume 为 no-op 或不调用）
