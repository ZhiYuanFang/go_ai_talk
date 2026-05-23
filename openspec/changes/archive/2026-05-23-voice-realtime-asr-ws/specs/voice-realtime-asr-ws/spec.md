## ADDED Requirements

### Requirement: Voice-service SHALL 提供实时听写 WebSocket 端点

`voice-service` MUST 在路径 `/voice/asr/ws` 提供 WebSocket 服务，将客户端上行的 PCM 音频流送入已配置的流式 STT（当前为百度流式 ASR），并将识别出的中文文本实时返回给客户端。

#### Scenario: 握手成功并开始会话

- **WHEN** 客户端对 `/voice/asr/ws` 发起 WebSocket Upgrade 且握手成功
- **AND** 客户端发送合法 `start` 文本帧（含非空 `deviceNo` 与有效 `sampleRate`/`bits`/`channels`）
- **THEN** 服务端 MUST 回复 `{"type":"started","code":0,"mode":"stream"}`（或等价字段集）
- **AND** 服务端 MUST NOT 调用对话 LLM、TTS 或设备最近对话落库接口

#### Scenario: 流式 STT 未启用或不可用

- **WHEN** `start` 成功但流式 STT 配置不可用（如 `stt.streamEnabled=false` 或 provider 不支持）
- **THEN** 服务端 MUST 发送 `{"type":"error","code":1,"stage":"stt",...}` 且 MUST NOT 假装听写成功

### Requirement: 上行协议 SHALL 限定为听写所需子集

听写 WebSocket MUST 接受以下上行消息类型，且 MUST NOT 要求客户端发送 `commit` 以完成听写：

- Text JSON：`type` 为 `start`（开始会话）
- Binary：16-bit 小端 PCM 音频分片，参数与 `start` 声明一致
- Text JSON：`type` 为 `end`（结束当前听写会话）
- Text：心跳 `ping`（服务端回复 `pong`）

#### Scenario: 未 start 即发送二进制音频

- **WHEN** 客户端在发送 `start` 之前发送二进制帧
- **THEN** 服务端 MUST 返回 `error`，`stage` 表明状态错误（如 `state`），且 MUST NOT 将音频送入 STT

#### Scenario: 客户端发送 end

- **WHEN** 客户端在已 `start` 的会话中发送 `{"type":"end"}`
- **THEN** 服务端 MUST 关闭当前流式 ASR 会话并回复 `{"type":"ended","code":0}`
- **AND** 服务端 MAY 在后续再次 `start` 开启新会话

#### Scenario: 客户端发送 commit

- **WHEN** 客户端发送 `{"type":"commit"}`
- **THEN** 服务端 MAY 将其视为可选「手动截句」触发 ASR finalize，或返回 `unsupported`；实现 MUST 在文档中固定一种行为且不得触发对话/TTS 链路

### Requirement: 下行协议 SHALL 仅包含听写相关事件

服务端下行 Text JSON MUST 以听写为主，至少包含：

- `asr_partial`：非空中间识别文本
- `asr_final`：一句或一轮识别定稿文本
- `error`：含 `stage` 与 `message`
- `started`、`ended`

服务端 MUST NOT 在该端点下发 `audio_chunk`、`chat_delta`、`exit` 或 TTS 相关字段。

#### Scenario: 收到 ASR 中间结果

- **WHEN** 流式 STT 产生新的中间文本且与上次 partial 不同
- **THEN** 服务端 MUST 发送 `{"type":"asr_partial","code":0,"text":"<识别文本>"}`

#### Scenario: 收到 ASR 最终结果

- **WHEN** 流式 STT 产生最终文本（含客户端 `end`、手动截句或服务端静音策略触发的 finalize）
- **THEN** 服务端 MUST 发送 `{"type":"asr_final","code":0,"text":"<识别文本>"}`

### Requirement: 听写连接 SHALL 与对话连接隔离

`voice-service` 在处理 `/voice/asr/ws` 时 MUST NOT 将连接注册到用于 `/voice/chat/ws` 的「单设备单连接」替换管理器（`VoiceWSManager`），以避免听写页与对话页互相踢连接。

#### Scenario: 同一 deviceNo 同时存在 chat 与 asr 连接

- **WHEN** 设备 `device-001` 已建立 `/voice/chat/ws` 连接且另建 `/voice/asr/ws` 连接
- **THEN** 两条连接 MUST 均可保持，直至各自关闭

### Requirement: Voice 域边界 SHALL 保持不变

听写实现 MUST 仅使用 voice 域已有 STT 能力与配置（`voice-chat.shared.yaml` / `Voice().CreateStreamASRSession`），且 MUST NOT 在 voice-service 内直接访问 device/history/user 等他域数据库表。

#### Scenario: 听写会话不查 device 库

- **WHEN** 客户端在 `start` 中提供 `deviceNo`
- **THEN** 服务端 MAY 将其用于日志与限流键，且 MUST NOT 为听写路径新增对 device 表 DAO 的依赖
