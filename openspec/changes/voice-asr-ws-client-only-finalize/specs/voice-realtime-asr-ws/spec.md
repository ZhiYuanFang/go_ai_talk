## MODIFIED Requirements

### Requirement: 上行协议 SHALL 限定为听写所需子集

听写 WebSocket MUST 接受以下上行消息类型：

- Text JSON：`type` 为 `start`（开始会话）
- Binary：16-bit 小端 PCM 音频分片，参数与 `start` 声明一致
- Text JSON：`type` 为 `commit`（**一句听写结束**，触发 finalize 并下发 `asr_final`）
- Text JSON：`type` 为 `end`（结束当前听写 WebSocket 会话）
- Text：心跳 `ping`（服务端回复 `pong`）

#### Scenario: 客户端发送 commit

- **WHEN** 客户端在已 `start` 的会话中发送 `{"type":"commit"}`
- **AND** 当前已有流式 ASR 会话或已接收过非空音频缓冲
- **THEN** 服务端 MUST 对当前句执行 STT finalize 并下发 `{"type":"asr_final",...,"source":"client"}`
- **AND** 服务端 MUST NOT 因此进入对话/TTS 链路

#### Scenario: commit 前无音频

- **WHEN** 客户端发送 `{"type":"commit"}` 且当前无 ASR 会话且无已缓冲音频
- **THEN** 服务端 MUST 返回 `error`（如 `stage=validate`）

#### Scenario: 客户端发送 end

- **WHEN** 客户端在已 `start` 的会话中发送 `{"type":"end"}`
- **THEN** 服务端 MUST 关闭当前流式 ASR 会话并回复 `{"type":"ended","code":0}`
- **AND** 若关闭前仍有未 finalize 的音频，服务端 MAY 先执行 finalize 并下发 `asr_final`（`source` 为 `end`）

### Requirement: 下行协议 SHALL 仅包含听写相关事件

服务端下行 Text JSON MUST 以听写为主，至少包含：

- `asr_partial`：非空中间识别文本
- `asr_final`：一句听写定稿文本（**仅**由客户端 `commit` 或 `end` 触发的 finalize 产生）
- `asr_no_result`：finalize 后无有效文本时
- `error`、`started`、`ended`

服务端 MUST NOT 在该端点下发 `audio_chunk`、`chat_delta`、`exit` 或 TTS 相关字段。

#### Scenario: 收到 ASR 中间结果

- **WHEN** 流式 STT 产生新的中间文本且与上次 partial 不同
- **THEN** 服务端 MUST 发送 `{"type":"asr_partial","code":0,"text":"<识别文本>"}`

#### Scenario: 收到 ASR 最终结果

- **WHEN** 客户端发送 `commit` 或 `end` 导致服务端执行 finalize 且得到有效转写文本
- **THEN** 服务端 MUST 发送 `{"type":"asr_final","code":0,"text":"<识别文本>"}` 且 `source` MUST 为 `client` 或 `end`

## ADDED Requirements

### Requirement: 听写 WS MUST NOT 服务端主动截句

`/voice/asr/ws` 实现 MUST NOT 基于服务端静音计时、STT 回调间隔或无回调时长自动调用 STT finalize 或向客户端发送 `asr_final`。

#### Scenario: 长时间静音但未发送 commit

- **WHEN** 客户端已 `start` 并持续发送二进制音频或保持连接
- **AND** 客户端在超过 2 秒内未发送 `commit` 或 `end`
- **AND** 流式 STT 未产生新的 partial 或产生引擎 final 回调
- **THEN** 服务端 MUST NOT 因静音或超时自动下发 `asr_final`
- **AND** 服务端 MUST NOT 因上述原因自动关闭并重建 ASR 会话

#### Scenario: 引擎 onFinal 回调

- **WHEN** 流式 STT 提供商推送引擎级 final 结果（如百度 `FIN_TEXT`）
- **THEN** 服务端 MUST NOT 将该结果作为 `asr_final` 下发给客户端
- **AND** 服务端 MUST NOT 仅因该回调而关闭当前 ASR 会话

#### Scenario: 禁止的 asr_final source

- **WHEN** 服务端在听写 WS 上产生 `asr_final`
- **THEN** `source` MUST NOT 为 `silence`、`auto_commit` 或 `asr_callback`
