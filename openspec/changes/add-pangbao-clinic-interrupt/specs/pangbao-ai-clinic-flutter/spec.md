## MODIFIED Requirements

### Requirement: Flutter SHALL 使用 ClinicWsClient 经 gateway-app 连接 WebSocket

App MUST 实现 `clinic_ws_client.dart`。连接 URL MUST 使用 `wsClinicUrl` / `wsClinicUrlEffective`，默认由 `apiBaseUrl`（gateway-app-server 主机）推导为 `wss://{host}/voice/clinic/ws`（对齐 `wsVoiceAsrUrlEffective` 模式）。客户端 **MUST NOT** 配置或连接 voice-service 内网地址。连接成功后 MUST 先发送首帧 `type=auth`（`accessToken` + `deviceNo`，与 history WS / UCG chat 一致），收到 `auth_ok` 后方可发送 `question`。每条 `question` MUST 含客户端生成的 UUID **`turnId`**。客户端 MUST 支持发送 **`type=cancel`**（含 `turnId`）。WS 生命周期：App 进入后台或离开诊疗页时 MUST **先** 对 active turn 发送 `cancel`（best-effort）**再** disconnect；回前台 MAY 重连且 MUST 重新 `auth`。客户端 MUST 解析 `auth_ok`、`session_sync`、`thinking_delta`、`answer_delta`、`answer_done`、**`turn_cancelled`**、`error` 帧；流式帧 **MUST** 按 **`turnId`** 与当前 active turn 对齐，不匹配 **MUST** 丢弃。

#### Scenario: 连接 gateway-app 而非 voice-service

- **WHEN** App 建立胖宝诊疗 WebSocket
- **THEN** 连接目标主机 MUST 与 `apiBaseUrl` 一致（gateway-app-server）
- **AND** 路径 MUST 为 `/voice/clinic/ws`（或含 `apiBaseUrl` path 前缀的等价路径）

#### Scenario: 首帧 auth

- **WHEN** clinic WS 握手成功
- **THEN** 客户端 MUST 发送 `auth` 帧且 MUST NOT 在收到 `auth_ok` 前发送 `question`

#### Scenario: question 携带 turnId

- **WHEN** 用户发送新问题
- **THEN** 客户端 MUST 生成新 UUID 作为 `turnId` 并随 `question` 上行

#### Scenario: 未登录不发 question

- **WHEN** 用户未登录（无有效 accessToken）
- **THEN** 客户端 MUST NOT 建立可提问的 clinic WS 连接；若服务端返回 40301 MUST 引导登录

#### Scenario: 离开页面显式 cancel

- **WHEN** 用户离开胖宝诊疗页或 App 进入后台且存在 active 流式 turn
- **THEN** 客户端 MUST 发送 `cancel`（含 active `turnId`）后再断开 WebSocket

#### Scenario: 流式展示回答

- **WHEN** 收到与 active `turnId` 一致的 `answer_delta` 序列后以 `answer_done` 结束
- **THEN** UI SHALL 逐字/逐段更新回答区域

#### Scenario: session_sync 恢复历史

- **WHEN** 收到 `session_sync` 且 `turns` 非空
- **THEN** App SHALL 将已完成轮次填充至聊天 `_items`（user 问 + assistant 答 + 免责声明）
- **AND** MUST NOT 为历史轮次渲染 thinking（服务端未提供）

#### Scenario: 丢弃 stale turnId 帧

- **WHEN** 收到 `thinking_delta` 或 `answer_delta` 且 `turnId` 不等于当前 active turn
- **THEN** 客户端 MUST 丢弃该帧且 MUST NOT 更新 UI

#### Scenario: turn_cancelled 清理进行中 UI

- **WHEN** 收到 `turn_cancelled` 且 `turnId` 为当前或刚结束的 active turn
- **THEN** UI MUST 清除该 turn 的进行中 thinking/answer 流式状态

## ADDED Requirements

### Requirement: Flutter SHALL 允许流式过程中停止或改问

胖宝诊疗页在 LLM 流式（thinking 或 answer）进行中 **MUST NOT** 全局锁定文本输入。用户 **MUST** 能够：（a）发送新问题，以新 `turnId` supersede 当前 turn；和/或（b）通过停止控件发送 `cancel` 中断当前 turn。改问 **MUST** supersede 服务端上一 turn（由服务端下发 `turn_cancelled` reason=superseded）。

#### Scenario: 流式期间发送新问题

- **WHEN** thinking 或 answer 流式进行中且用户输入并发送新问题
- **THEN** 客户端 MUST 分配新 `turnId` 并发送 `question`
- **AND** UI MUST 展示新问题的 user 气泡并等待新 turn 的流式回复

#### Scenario: 流式期间点击停止

- **WHEN** 流式进行中且用户点击停止
- **THEN** 客户端 MUST 发送 `cancel` 且 `turnId` 为当前 active turn
- **AND** 收到 `turn_cancelled` 后 MUST 结束进行中 assistant 流式 UI

#### Scenario: 可选编辑用户气泡改问

- **WHEN** 实现 tap 用户气泡编辑（optional）
- **THEN** 预填问题文本后发送 MUST 使用新 `turnId` 并 supersede 当前 turn
