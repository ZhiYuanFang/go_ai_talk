## MODIFIED Requirements

### Requirement: voice-service SHALL 提供胖宝诊疗 WebSocket handler

`voice-service` MUST 在路径 `/voice/clinic/ws` 注册 WebSocket handler（`BindHandler`），处理经 gateway-app 透传而来的客户端文本提问并流式返回 DeepSeek 回答。用户可见功能名称为 **胖宝诊疗**；实现与配置仍使用 `clinic` / `clinic_ai` / `voice:clinic:*` 命名。该路径为**集群内业务端点**；App 对外入口 MUST 为 gateway-app-server 同路径透传（见 `gateway-ws-edge-proxy`）。实现 MUST NOT 将连接注册到 `VoiceWSManager`。实现 MUST NOT 提供 TTS 或音频上行能力（MVP 纯文本）。

#### Scenario: 经 gateway-app 透传后握手成功

- **WHEN** 客户端经 gateway-app 透传对 voice-service `/voice/clinic/ws` 完成 WebSocket Upgrade
- **THEN** voice-service SHALL 接受连接并等待首帧 `auth` JSON

### Requirement: Clinic WebSocket SHALL 使用规定的帧协议

握手成功后，客户端首帧 MUST 为 `type=auth`（含 `accessToken` 与 `deviceNo`）。`auth_ok` 之后，服务端 MUST 下发 `type=session_sync`（见 ADDED requirement）。`session_sync` 之后，客户端上行 MUST 发送 `type=question` 帧，含非空 `text`。服务端下行 MUST 支持 `auth_ok`、`session_sync`、`thinking_delta`、`answer_delta`、`answer_done`、`error` 六种 `type`。`error` 帧 MUST 含 numeric `code` 与 `message` 字符串。

#### Scenario: 流式 thinking 与 answer

- **WHEN** 客户端已完成 `auth` 并发送合法 `question` 且 LLM 流式返回 reasoning 与 content
- **THEN** 服务端 MUST 先/交错推送 `thinking_delta`，再推送 `answer_delta`，最终以 `answer_done` 结束

#### Scenario: auth 前拒绝 question

- **WHEN** 客户端未发送 `auth` 或 `auth` 未成功即发送 `question`
- **THEN** 服务端 SHALL 返回 `error` 帧且 MUST NOT 调用 LLM

#### Scenario: 空问题拒绝

- **WHEN** 客户端发送 `question` 且 `text` 为空或仅空白
- **THEN** 服务端 SHALL 返回 `error` 帧且 MUST NOT 调用 LLM

## ADDED Requirements

### Requirement: Clinic SHALL 在 auth_ok 后下发 session_sync

`auth_ok` 成功后，voice-service MUST 读取 Redis `voice:clinic:session:{wxId}` 并立即向客户端下发 `session_sync` 帧。每次 WebSocket 重连并完成 `auth` 后 MUST 重复下发。payload MUST 含 `turns` 数组与 `expiresAt`（Unix 秒）。`turns` 每项 MUST 仅含 `question` 与 `answer` 字符串；MUST NOT 含 `thinking`。仅 MUST 包含 question 与 answer 均非空的已完成轮次（与 session 写入 `appendClinicTurn` 语义一致）。无 session 时 MUST 下发 `turns: []` 且 `expiresAt` 为 0 或省略。

#### Scenario: 有历史会话时恢复轮次

- **WHEN** wxId=1001 在 12h 内已有 2 轮已完成 Q&A 且客户端 `auth` 成功
- **THEN** 服务端 SHALL 在 `auth_ok` 后下发 `session_sync` 且 `turns` 长度为 2
- **AND** 每轮 MUST 含对应 `question` 与 `answer`
- **AND** MUST NOT 含 thinking 字段

#### Scenario: 无 session 时空同步

- **WHEN** wxId=1001 尚未发送过 `question` 且客户端 `auth` 成功
- **THEN** 服务端 SHALL 下发 `session_sync` 且 `turns` 为空数组

#### Scenario: 重连重复同步

- **WHEN** 同一 wxId 断开 WS 后重连并再次 `auth` 成功
- **THEN** 服务端 SHALL 再次下发 `session_sync` 且内容与当前 Redis session 一致

#### Scenario: expiresAt 反映会话过期

- **WHEN** session 存在且 `firstQuestionAt` 已知
- **THEN** `session_sync.expiresAt` SHALL 为会话绝对过期 Unix 秒（首问时刻 + 12h 固定 TTL）
