## MODIFIED Requirements

### Requirement: Clinic WebSocket SHALL 使用规定的帧协议

握手成功后，客户端首帧 MUST 为 `type=auth`（含 `accessToken` 与 `deviceNo`）。`auth_ok` 之后，服务端 MUST 下发 `type=session_sync`（见 `amend-pangbao-clinic-ux`）。`session_sync` 之后，客户端上行 MUST 支持 `type=question` 帧（含非空 `text` 与非空 **`turnId`** UUID）与 **`type=cancel`** 帧（含非空 **`turnId`**）。服务端下行 MUST 支持 `auth_ok`、`session_sync`、`thinking_delta`、`answer_delta`、`answer_done`、**`turn_cancelled`**、`error` 七种 `type`。流式下行帧（`thinking_delta`、`answer_delta`、`answer_done`）**MUST** 含与当前 turn 一致的 **`turnId`**。`turn_cancelled` MUST 含 **`turnId`** 与 **`reason`**，取值为 **`superseded`**、**`cancelled`** 或 **`disconnected`** 之一。`error` 帧 MUST 含 numeric `code` 与 `message` 字符串。

#### Scenario: 流式 thinking 与 answer 携带 turnId

- **WHEN** 客户端发送 `question` 且 `turnId=uuid-A` 且 LLM 流式返回 reasoning 与 content
- **THEN** 服务端 MUST 推送的 `thinking_delta`、`answer_delta` 与最终 `answer_done` 均 MUST 含 `turnId=uuid-A`

#### Scenario: auth 前拒绝 question

- **WHEN** 客户端未发送 `auth` 或 `auth` 未成功即发送 `question`
- **THEN** 服务端 SHALL 返回 `error` 帧且 MUST NOT 调用 LLM

#### Scenario: 空问题或缺少 turnId 拒绝

- **WHEN** 客户端发送 `question` 且 `text` 为空或仅空白，或缺少/空 `turnId`
- **THEN** 服务端 SHALL 返回 `error` 帧且 MUST NOT 调用 LLM

#### Scenario: cancel 中断当前 turn

- **WHEN** 客户端在 turn `uuid-A` 流式进行中发送 `cancel` 且 `turnId=uuid-A`
- **THEN** 服务端 MUST 取消该 turn 的 LLM 上下文
- **AND** MUST 下发 `turn_cancelled` 且 `turnId=uuid-A` 且 `reason=cancelled`
- **AND** MUST NOT 下发 `answer_done` 且 MUST NOT consume `clinic_ai`

#### Scenario: 新 question supersede 上一 turn

- **WHEN** turn `uuid-A` 仍在流式进行中且客户端发送新 `question` 且 `turnId=uuid-B`
- **THEN** 服务端 MUST 取消 turn A 的 LLM 上下文
- **AND** MUST 下发 `turn_cancelled` 且 `turnId=uuid-A` 且 `reason=superseded`
- **AND** MUST 开始处理 turn B

#### Scenario: WS 断开取消 active turn

- **WHEN** 连接存在 active LLM turn 且 WebSocket 读循环因关闭或错误退出
- **THEN** 服务端 MUST 取消该 turn 的 LLM 上下文
- **AND** MUST NOT consume `clinic_ai` 且 MUST NOT append session

### Requirement: Clinic SHALL 强制执行 clinic_ai 月度额度（per wxId）

voice-service 在调用 Clinic LLM 前 MUST 使用 auth 已绑定的 `wxId>0`；`wxId≤0` MUST 返回 `error` code **40301**。LLM 调用前 MUST 经 device internal 对 feature `clinic_ai` 以该 wxId 执行 check；`allowed=false` MUST 返回 code **40302** message **「本月额度已用完」** 且 MUST NOT 调用 LLM。**仅** 当 turn 以 **`answer_done` 成功结束** 时 MUST 以同一 wxId consume。**cancelled**、**superseded**、**disconnected** 或 LLM/摘要失败而中断的 turn **MUST NOT** consume。

#### Scenario: 未登录

- **WHEN** wxId 解析为 0 且用户发送 `question`
- **THEN** WS SHALL 返回 40301 且 MUST NOT 调用 LLM

#### Scenario: clinic_ai 额度用尽

- **WHEN** check 得到 clinic_ai used=limit
- **THEN** WS SHALL 返回 40302 且 MUST NOT 调用 LLM

#### Scenario: 用户 cancel 不扣额度

- **WHEN** turn 在流式过程中被 `cancel` 或 `superseded` 结束且未收到 `answer_done`
- **THEN** 系统 MUST NOT 对该 turn 调用 consume `clinic_ai`

#### Scenario: 成功完成才扣额度

- **WHEN** turn 完整流式结束并下发 `answer_done`
- **THEN** 系统 MUST consume `clinic_ai` 一次

### Requirement: Clinic SHALL 实施 Redis 限流（per wxId）

系统 MUST 对 clinic 提问路径实施 Redis 限流，键 **`voice:clinic:rate:{wxId}`**。限流计数 **MUST** 在 **`answer_done` 成功** 后递增；**cancelled**、**superseded**、**disconnected** 或失败 turn **MUST NOT** 递增限流计数。处理新 `question` 前 MUST 检查当前窗口计数；超限时 MUST 返回 WS `error` code **42901** 且 MUST NOT 调用 LLM。

#### Scenario: 短时间频繁提问

- **WHEN** 同一 wxId 在限流窗口内已成功完成（`answer_done`）次数超过阈值
- **THEN** 下一次 `question` SHALL 返回 42901 且 MUST NOT 调用 LLM

#### Scenario: supersede 未完成 turn 不计入限流

- **WHEN** 用户在窗口内多次改问但均未产生 `answer_done`
- **THEN** 限流计数 MUST NOT 因 supersede 而额外递增

### Requirement: Clinic 会话 SHALL 使用固定 12 小时 TTL（wxId 键）

Redis 键 **`voice:clinic:session:{wxId}`** MUST 在**首条**成功完成的 `question`（即首次 `answer_done` 后 append）时创建，TTL 为 12 小时自 `firstQuestionAt` 起算。后续提问 MUST NOT 滑动续期 TTL。进入胖宝页与 `auth_ok` MUST NOT 预创建 session。Session MUST 记录 auth 时锁定的 `deviceNo` 供摘要使用。**未完成**（cancelled / superseded / disconnected）的 turn **MUST NOT** append 至 session。

#### Scenario: 首问创建 session

- **WHEN** wxId=1001 用户首条 turn 以 `answer_done` 成功结束
- **THEN** Redis MUST 写入键 `voice:clinic:session:1001` 且 EX=12h

#### Scenario: 会话内多轮上下文

- **WHEN** 12h 内同一 wxId session 已有完成轮次且新 turn 以 `answer_done` 成功结束
- **THEN** LLM 上下文 SHALL 包含同 session 内先前已完成 Q&A

#### Scenario: TTL 过期后会话重置

- **WHEN** 首问后超过 12h 同一 wxId 再提问
- **THEN** 系统 MUST 创建新 session 且 prior Q&A 上下文 SHALL 为空

#### Scenario: 取消的 turn 不写入 session

- **WHEN** turn 被 cancel 或 supersede 且未产生 `answer_done`
- **THEN** Redis session MUST NOT 追加该 partial Q&A

## ADDED Requirements

### Requirement: Clinic WS handler SHALL 非阻塞处理 question 并支持显式取消

`/voice/clinic/ws` 读循环 MUST NOT 同步阻塞在单条 `HandleQuestion` 直至 LLM 结束。每条合法 `question` MUST 在独立 goroutine 中处理，且连接 MUST 维护 at most one **active turn** 及其 `context.CancelFunc`。收到匹配 active turn 的 `cancel`、收到新 `question`（supersede）或连接关闭时 MUST 调用 cancel 中断 LLM 流。

#### Scenario: 读循环可并发接收 cancel

- **WHEN** LLM 流式进行中且读循环收到 `cancel` 帧
- **THEN** handler MUST 在不等待 LLM 自然结束的情况下处理 cancel

#### Scenario: HandleQuestion 尊重 turn context

- **WHEN** turn context 已被 cancel
- **THEN** `HandleQuestion` MUST 停止 LLM 读流且 MUST NOT 写 `answer_done`
