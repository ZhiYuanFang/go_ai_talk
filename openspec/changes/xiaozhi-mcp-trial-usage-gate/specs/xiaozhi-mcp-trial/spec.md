## ADDED Requirements

### Requirement: 每账号一次试用 24 小时

系统 SHALL 为小智 MCP 能力提供按 `wx_id` 计的免费试用：每账号仅可成功领取一次，领取后权益有效期 MUST 为 24 小时（与既有商业功能试用时长常量对齐）。试用 MUST NOT 写入永久权益（`expires_at` MUST 为未来截止时间，MUST NOT 为 0）。

#### Scenario: 首次领取得到限时权益

- **WHEN** 某 wx 试用状态为未使用且尚无有效小智 MCP 权益，并成功完成试用领取
- **THEN** 系统标记该账号试用已使用，并写入有效权益且 `expires_at` 约为领取时刻起 24 小时
- **AND** MUST NOT 将 `expires_at` 记为 0（永久）

#### Scenario: 不可重复领取

- **WHEN** 某 wx 已成功领取过试用（无论当前权益是否已过期）
- **THEN** 再次试用领取 MUST 失败或被跳过且不再授予新的试用时段

### Requirement: 首次成功 Add 时领取试用

系统 SHALL 在 App 添加小智绑定的成功路径上，对「无有效权益且试用未用」的账号执行试用领取。领取与绑定写入 MUST 保证：领取失败则 MUST NOT 留下本次新增绑定；已有有效权益（试用未过期或永久）时 MUST NOT 再次消耗试用。

#### Scenario: 未开通未试用时 Add 触发 claim

- **WHEN** 当前 wx 无有效权益、试用未用，且绑定其它校验通过
- **THEN** 系统先（或原子地）完成试用领取，再写入绑定
- **AND** 写成功后该 wx 在试用窗内视为已开通

#### Scenario: 试用已用且已过期时拒绝 Add

- **WHEN** 当前 wx 试用已用尽且当前无有效权益
- **THEN** 添加绑定 MUST 失败
- **AND** MUST NOT 写入新绑定行

#### Scenario: 永久或试用有效时 Add 不消耗试用

- **WHEN** 当前 wx 已有有效永久或未过期试用权益并添加绑定
- **THEN** 添加按既有规则成功
- **AND** MUST NOT 再次将试用从 unused 转为 used（若已 used 则保持）

### Requirement: 开通态暴露试用信息

App 小智 MCP 开通态接口 SHALL 返回是否已开通、试用是否仍可领取，以及当前权益到期时间（永久时以约定哨兵如 `expiresAt=0` 表示）。

#### Scenario: 可试用

- **WHEN** 用户未开通且试用未用
- **THEN** 响应含 `unlocked=false` 且 `trialAvailable=true`（或等价字段）

#### Scenario: 试用中

- **WHEN** 用户处于未过期试用
- **THEN** 响应 `unlocked=true` 且带有未来的 `expiresAt`
