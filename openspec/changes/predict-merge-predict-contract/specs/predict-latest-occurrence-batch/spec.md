## ADDED Requirements

### Requirement: Server SHALL expose batch latest occurrence by event from history only

The system MUST provide an authenticated App HTTP API `GET /device/history/api/event/latest-by-event` that accepts `deviceNo` (same bind semantics as existing history App APIs). The handler MUST query only the history domain table (MUST NOT read `appointment_next` or other appointment stores). For each `event_id` that has at least one history row for that device, the response MUST include at most one summary whose occurrence time is that row's `start_time` (unix seconds), choosing the row with the maximum `start_time` (ties MAY break by id desc). Rows with `end_time=0` (in progress) MUST be eligible. Each item MUST include `eventId` and `occurredAt` (`start_time`); MAY include `remark` and `historyId`. Empty history MUST return an empty list without error. 服务 MUST 提供纯 history 批量最近发生接口；发生时刻为 start_time；含未闭合；不含预约表。

#### Scenario: Multiple events one response

- **WHEN** 宝宝 history 含多个 eventId 的记录
- **THEN** 响应 MUST 为每个有记录的 eventId 给出一条最新发生，且 `occurredAt` 等于该条的 `start_time`

#### Scenario: In-progress counts as latest

- **WHEN** 某 eventId 最近一条 history 的 `end_time=0` 且其 `start_time` 最大
- **THEN** 该 eventId 的摘要 MUST 仍返回，且 `occurredAt` 为该未闭合行的 `start_time`

#### Scenario: Empty history

- **WHEN** 宝宝无 history 行
- **THEN** 响应 MUST 为空列表且 MUST NOT 报错

#### Scenario: Appointment next not mixed in

- **WHEN** 宝宝仅有预约下次约定、无对应 history 行
- **THEN** 该 eventId MUST NOT 仅因预约表出现在本接口结果中

### Requirement: latest-by-event MUST require login and MUST NOT count toward usage

The `GET /device/history/api/event/latest-by-event` route MUST require Bearer authentication (MUST NOT be listed on gateway auth exempt). The path MUST be registered in App usage maintenance skip list so it is not counted toward usage stats (owner confirmed). latest-by-event MUST 登录；MUST NOT 计入 usage；MUST NOT Bearer 豁免。

#### Scenario: Unauthenticated rejected

- **WHEN** 无有效 Bearer 调用该 GET
- **THEN** 网关或服务 MUST 拒绝（与既有需登录 history 读接口一致）
