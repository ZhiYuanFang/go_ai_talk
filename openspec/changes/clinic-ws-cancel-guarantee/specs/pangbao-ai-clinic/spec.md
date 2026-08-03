## MODIFIED Requirements

### Requirement: Clinic WebSocket SHALL 使用规定的帧协议

握手成功后，客户端首帧 MUST 为 `type=auth`（含 `accessToken` 与 `deviceNo`）。`auth_ok` 之后，客户端上行 MUST 支持 `type=question` 帧（含非空 `text` 与非空 **`turnId`** UUID）与 **`type=cancel`** 帧（含非空 **`turnId`**）。服务端下行 MUST 支持 `auth_ok`、`thinking_delta`、`answer_delta`、`answer_done`、**`turn_cancelled`**、`error` 六种 `type`，MUST NOT 下发 `session_sync`。流式下行帧（`thinking_delta`、`answer_delta`、`answer_done`）**MUST** 含与当前 turn 一致的 **`turnId`**。`turn_cancelled` MUST 含 **`turnId`** 与 **`reason`**，取值 MUST 为 **`superseded`** 或 **`cancelled`** 之一（MUST NOT 使用 `disconnected` 作为 `reason`，亦 MUST NOT 附加 question 文本字段）。`error` 帧 MUST 含 numeric `code` 与 `message` 字符串。

#### Scenario: 流式 thinking 与 answer 携带 turnId

- **WHEN** 客户端发送 `question` 且 `turnId=uuid-A` 且 LLM 流式返回 reasoning 与 content
- **THEN** 服务端 MUST 推送的 `thinking_delta`、`answer_delta` 与最终 `answer_done` 均 MUST 含 `turnId=uuid-A`

#### Scenario: auth 前拒绝 question

- **WHEN** 客户端未发送 `auth` 或 `auth` 未成功即发送 `question`
- **THEN** 服务端 SHALL 返回 `error` 帧且 MUST NOT 调用 LLM

#### Scenario: 空问题或缺少 turnId 拒绝

- **WHEN** 客户端发送 `question` 且 `text` 为空或仅空白，或缺少/空 `turnId`
- **THEN** 服务端 SHALL 返回 `error` 帧且 MUST NOT 调用 LLM

#### Scenario: cancel 中断当前 turn（thinking 或 answer）

- **WHEN** 客户端在 turn `uuid-A` 的 thinking 或 answer 流式进行中发送 `cancel` 且 `turnId=uuid-A`
- **THEN** 服务端 MUST 取消该 turn 的 LLM/Python 请求上下文
- **AND** MUST 下发 `turn_cancelled` 且 `turnId=uuid-A` 且 `reason=cancelled`
- **AND** MUST NOT 下发 `answer_done` 且 MUST NOT consume `clinic_ai`
- **AND** MUST NOT 递增 clinic 限流成功计数

#### Scenario: 新 question supersede 上一 turn

- **WHEN** turn `uuid-A` 仍在流式进行中且客户端发送新 `question` 且 `turnId=uuid-B`
- **THEN** 服务端 MUST 取消 turn A 的 LLM 上下文
- **AND** MUST 下发 `turn_cancelled` 且 `turnId=uuid-A` 且 `reason=superseded`
- **AND** MUST 开始处理 turn B
- **AND** MUST NOT 因 supersede 路径对 turn A 下发 `answer_done` 或 consume `clinic_ai`

#### Scenario: 不匹配的 cancel 静默忽略

- **WHEN** 客户端发送 `cancel` 且 `turnId` 与当前 active turn 不一致，或当前无 active turn
- **THEN** 服务端 MUST NOT 取消其他 turn 的上下文
- **AND** MUST NOT 下发 `turn_cancelled`（对该 cancel）

#### Scenario: WS 断开取消 active turn 但不下发 turn_cancelled

- **WHEN** 连接存在 active LLM turn 且 WebSocket 读循环因关闭或错误退出
- **THEN** 服务端 MUST 取消该 turn 的 LLM 上下文
- **AND** MUST NOT consume `clinic_ai`
- **AND** MUST NOT 下发 `turn_cancelled`（含 MUST NOT 使用 `reason=disconnected`）

### Requirement: Clinic WS handler SHALL 非阻塞处理 question 并支持显式取消

`/voice/clinic/ws` 读循环 MUST NOT 同步阻塞在单条 `HandleQuestion` 直至 LLM 结束。每条合法 `question` MUST 在独立 goroutine 中处理，且连接 MUST 维护 at most one **active turn** 及其 `context.CancelFunc`。收到匹配 active turn 的 `cancel`、收到新 `question`（supersede）或连接关闭时 MUST 调用 cancel 中断 LLM/Python 流。用户显式 `cancel` 与 supersede MUST 分别对应下行 `turn_cancelled` 的 `reason=cancelled` 与 `reason=superseded`；连接关闭 MUST 仅取消上下文且 MUST NOT 依赖下行 `turn_cancelled`。

#### Scenario: 读循环可并发接收 cancel

- **WHEN** LLM 流式进行中且读循环收到 `cancel` 帧
- **THEN** handler MUST 在不等待 LLM 自然结束的情况下处理 cancel

#### Scenario: HandleQuestion 尊重 turn context

- **WHEN** turn context 已被 cancel
- **THEN** `HandleQuestion` MUST 停止 LLM 读流且 MUST NOT 写 `answer_done`

#### Scenario: cancel 在 thinking 阶段即时生效

- **WHEN** 服务端已开始下发 `thinking_delta` 且尚未进入 `answer_delta`，客户端发送匹配的 `cancel`
- **THEN** 服务端 MUST 取消 turn context 并下发 `turn_cancelled` 且 `reason=cancelled`

#### Scenario: cancel 在 answer 阶段即时生效

- **WHEN** 服务端已开始下发 `answer_delta` 且尚未 `answer_done`，客户端发送匹配的 `cancel`
- **THEN** 服务端 MUST 取消 turn context 并下发 `turn_cancelled` 且 `reason=cancelled`
- **AND** MUST NOT 随后下发该 turn 的 `answer_done`
