## ADDED Requirements

### Requirement: Admin 已开通人员快照列表

系统 SHALL 提供经 Admin 鉴权的接口，分页返回 `xiaozhi_mcp_entitlement` 当前快照（含有效与已失效行），并为本页 `wxId` 批量附带公开昵称（昵称拉取失败时 MUST 仍返回列表，昵称可为空）。列表 MUST NOT 经 App Bearer 匿名可访问。

#### Scenario: 分页列出含昵称

- **WHEN** 运维携带合法 Admin 凭据请求开通人员列表（含 limit/offset）
- **THEN** 系统返回 `total` 与当前页 `list`
- **AND** 每行至少包含 `wxId`、`unlockMethod`、`active`、`expiresAt`、`unlockedAt`、`updatedAt`、`channelRef`
- **AND** 在能拉到公开资料时填充 `nickname`

#### Scenario: active 语义

- **WHEN** 某行 `status` 为有效且（`expiresAt=0` 永久或 `expiresAt` 未过期）
- **THEN** 该行 `active` MUST 为 true
- **WHEN** 某行已撤销或试用已过期
- **THEN** 该行 `active` MUST 为 false 且仍可出现在列表中（非默认过滤掉）

#### Scenario: 无口令拒绝

- **WHEN** 请求缺少合法 Admin 凭据
- **THEN** 列表接口 MUST 拒绝

### Requirement: Hub 区 C 展示与行内撤销

系统 SHALL 在「小智 MCP 开通」静态页提供 **区 C · 已开通**，调用上述列表接口渲染表格（含昵称、方式、状态、到期、时间、痕迹），并支持刷新与分页。对仍有效且 `unlockMethod=admin` 的行 MUST 提供行内撤销入口，调用既有手工授撤销接口；支付或试用开通 MUST NOT 显示可成功撤销的行内操作（或不提供按钮）。区 B 手工授或撤销成功后，页面 SHOULD 刷新区 C。

#### Scenario: 进页可见已开通表

- **WHEN** 运维打开小智 MCP 开通页并通过 Admin 鉴权
- **THEN** 可见区 C 表格并由列表接口填充当前页数据

#### Scenario: 行内撤销手工授

- **WHEN** 某行 `active=true` 且 `unlockMethod=admin`，运维确认行内撤销
- **THEN** 系统按既有手工授撤销语义使该账号能力变为未开通（除非另有仍有效支付开通——首版一人一行权益则即未开通）
- **AND** 区 C 刷新后该行不再显示为有效 admin 开通

#### Scenario: 支付开通无行内撤销

- **WHEN** 某行仍有效且 `unlockMethod=payment`（或 `trial`）
- **THEN** 区 C MUST NOT 提供可对该行成功撤销的行内按钮（或按钮不可用）
