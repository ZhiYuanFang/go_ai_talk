## ADDED Requirements

### Requirement: voice-service SHALL 提供胖宝 AI 诊室 WebSocket handler

`voice-service` MUST 在路径 `/voice/clinic/ws` 注册 WebSocket handler（`BindHandler`），处理经 gateway-app 透传而来的客户端文本提问并流式返回 DeepSeek 回答。该路径为**集群内业务端点**；App 对外入口 MUST 为 gateway-app-server 同路径透传（见 `gateway-ws-edge-proxy`）。实现 MUST NOT 将连接注册到 `VoiceWSManager`。实现 MUST NOT 提供 TTS 或音频上行能力（MVP 纯文本）。

#### Scenario: 经 gateway-app 透传后握手成功

- **WHEN** 客户端经 gateway-app 透传对 voice-service `/voice/clinic/ws` 完成 WebSocket Upgrade
- **THEN** voice-service SHALL 接受连接并等待首帧 `auth` JSON

### Requirement: Clinic WebSocket SHALL 以 wxId 为主键绑定身份

`/voice/clinic/ws` 的连接、会话、限流与额度维度 MUST 以 **`wx.id`（正整数）** 为主键，与 `/voice/chat/ws` 以 `deviceNo` 注册 `VoiceWSManager` 的行为 MUST 显式不同。实现 MUST NOT 使用 `deviceNo` 反查 wxId 作为 clinic 鉴权路径（即 MUST NOT 调用 `ResolveVoiceWxID` 的 deviceNo fallback）。`deviceNo` MAY 仅用于 history 摘要拉取，且 MUST 与 JWT `device_no` claim 一致。

#### Scenario: 首帧 auth 绑定 wxId

- **WHEN** 客户端握手后发送 `type=auth` 且 JWT 含 `sub=1001` 与有效 `device_no`
- **THEN** 服务端 SHALL 解析 `wxId=1001` 并返回 `auth_ok`
- **AND** 后续 session/限流/额度 MUST 使用 wxId=1001

#### Scenario: 未登录拒绝

- **WHEN** 客户端 `auth` 帧 JWT 无效或 `sub≤0`
- **THEN** 服务端 SHALL 返回 `error` code **40301** 且 MUST NOT 进入 `question` 处理

#### Scenario: deviceNo 与 JWT 不一致

- **WHEN** `auth` 帧 `deviceNo` 与 JWT `device_no` claim 不一致
- **THEN** 服务端 SHALL 返回参数错误 `error` 帧且 MUST NOT 返回 `auth_ok`

#### Scenario: 与 voice ball 连接互不踢线

- **WHEN** 同一 `deviceNo` 已建立 `/voice/chat/ws` 且同一 `wxId` 另建 `/voice/clinic/ws`
- **THEN** 两条连接 SHALL 同时保持，且 clinic 连接 MUST NOT 触发 `VoiceWSManager` 替换逻辑

### Requirement: Clinic WebSocket SHALL 使用规定的帧协议

握手成功后，客户端首帧 MUST 为 `type=auth`（含 `accessToken` 与 `deviceNo`）。`auth_ok` 之后，客户端上行 MUST 发送 `type=question` 帧，含非空 `text`。服务端下行 MUST 支持 `auth_ok`、`thinking_delta`、`answer_delta`、`answer_done`、`error` 五种 `type`。`error` 帧 MUST 含 numeric `code` 与 `message` 字符串。

#### Scenario: 流式 thinking 与 answer

- **WHEN** 客户端已完成 `auth` 并发送合法 `question` 且 LLM 流式返回 reasoning 与 content
- **THEN** 服务端 MUST 先/交错推送 `thinking_delta`，再推送 `answer_delta`，最终以 `answer_done` 结束

#### Scenario: auth 前拒绝 question

- **WHEN** 客户端未发送 `auth` 或 `auth` 未成功即发送 `question`
- **THEN** 服务端 SHALL 返回 `error` 帧且 MUST NOT 调用 LLM

#### Scenario: 空问题拒绝

- **WHEN** 客户端发送 `question` 且 `text` 为空或仅空白
- **THEN** 服务端 SHALL 返回 `error` 帧且 MUST NOT 调用 LLM

### Requirement: Clinic SHALL 注入近 7 天喂养事件聚合摘要

每次处理 `question` 前，系统 MUST 取得该设备近 7 天喂养 history，并按 event 聚合为摘要（含 count、amount 合计、duration 合计等），注入 LLM system/context。摘要 MUST NOT 为原始 history 行全量 dump。history 数据 MUST 经 HTTP 契约（如 `DeviceHistory`）获取；voice-service MUST NOT 直连 history 库表。

#### Scenario: 摘要在 prompt 中可见

- **WHEN** 设备近 7 天有 3 次「母乳」记录
- **THEN** 注入 LLM 的上下文中 SHALL 包含按 event 聚合后的统计，且 token 量 SHALL 小于同等条数原始 JSON 行列表

#### Scenario: history 契约失败

- **WHEN** history HTTP 契约不可用或返回错误
- **THEN** 系统 SHALL 返回可诊断 `error` 帧且 MUST NOT 在无摘要时静默调用 LLM（除非 design 明确降级为空摘要并记录日志——本 spec 要求显式失败或空摘要+警告日志二选一，实现 MUST 在 design 中择一并在日志中可观测）

### Requirement: Clinic 摘要 SHALL 懒刷新

系统 MUST 在每次 `question` 前对比 history watermark 与已缓存摘要的 watermark；若缓存缺失或 watermark 过期，MUST 重新计算摘要并更新 Redis 缓存 `voice:clinic:summary:{wxId}:{deviceNo}`（或等价键）。

#### Scenario: history 未变化复用缓存

- **WHEN** 同一 session 内第二次提问且 history watermark 未变
- **THEN** 系统 SHALL 复用 Redis 中摘要且 MUST NOT 重复全量聚合计算

#### Scenario: 新喂养记录触发重算

- **WHEN** 用户新增喂养记录导致 history watermark 前进
- **THEN** 下一次 `question` MUST 重算摘要并更新缓存

### Requirement: Clinic 会话 SHALL 使用固定 12 小时 TTL（wxId 键）

Redis 键 **`voice:clinic:session:{wxId}`** MUST 在**首条** `question` 时创建，TTL 为 12 小时自 `firstQuestionAt` 起算。后续提问 MUST NOT 滑动续期 TTL。进入胖宝页与 `auth_ok` MUST NOT 预创建 session。Session MUST 记录 auth 时锁定的 `deviceNo` 供摘要使用。

#### Scenario: 首问创建 session

- **WHEN** wxId=1001 用户发送首条 `question`
- **THEN** Redis MUST 写入键 `voice:clinic:session:1001` 且 EX=12h

#### Scenario: 会话内多轮上下文

- **WHEN** 12h 内同一 wxId session 发送第二条 `question`
- **THEN** LLM 上下文 SHALL 包含同 session 内先前 Q&A

#### Scenario: TTL 过期后会话重置

- **WHEN** 首问后超过 12h 同一 wxId 再提问
- **THEN** 系统 MUST 创建新 session 且 prior Q&A 上下文 SHALL 为空

### Requirement: Clinic LLM SHALL 使用 deepseek-v4-pro 并启用 thinking

Clinic LLM 调用 MUST 使用模型 `deepseek-v4-pro`，MUST 启用 thinking（`extra_body.thinking` 或等价配置，`reasoning_effort: high`）。LLM 请求超时 MUST 为 **120 秒**，配置源 MUST 为 `config.voice-service.yaml` 的 `aiClinic` 块，MUST NOT 依赖 `voice-chat.shared.yaml` 的 voice ball 超时。

#### Scenario: thinking 流映射

- **WHEN** DeepSeek 返回 reasoning 流
- **THEN** 服务端 MUST 将其映射为 `thinking_delta` 帧

### Requirement: Clinic SHALL 强制执行 clinic_ai 月度额度（per wxId）

voice-service 在调用 Clinic LLM 前 MUST 使用 auth 已绑定的 `wxId>0`；`wxId≤0` MUST 返回 `error` code **40301**。LLM 调用前 MUST 经 device internal 对 feature `clinic_ai` 以该 wxId 执行 check；`allowed=false` MUST 返回 code **40302** message **「本月额度已用完」** 且 MUST NOT 调用 LLM。LLM 流式成功完成后 MUST 以同一 wxId consume。

#### Scenario: 未登录

- **WHEN** wxId 解析为 0 且用户发送 `question`
- **THEN** WS SHALL 返回 40301 且 MUST NOT 调用 LLM

#### Scenario: clinic_ai 额度用尽

- **WHEN** check 得到 clinic_ai used=limit
- **THEN** WS SHALL 返回 40302 且 MUST NOT 调用 LLM

### Requirement: Clinic SHALL 实施 Redis 限流（per wxId）

系统 MUST 对 clinic 提问路径实施 Redis 限流，键 **`voice:clinic:rate:{wxId}`**。超限时 MUST 返回 WS `error` code **42901** 且 MUST NOT 调用 LLM。

#### Scenario: 短时间频繁提问

- **WHEN** 同一 wxId 在限流窗口内超过阈值发送 `question`
- **THEN** 系统 SHALL 返回 42901

<!-- gateway-app 注册与 App 对外入口要求见 gateway-ws-edge-proxy 变更增量 -->
