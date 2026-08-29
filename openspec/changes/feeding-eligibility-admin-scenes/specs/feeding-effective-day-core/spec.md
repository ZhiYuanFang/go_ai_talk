## ADDED Requirements

### Requirement: 有效喂养日判定 MUST 独立于场景门槛且仅由 cash-service 执行

系统 MUST 在 **cash-service** 提供与场景解耦的有效日判定：使用 `Asia/Shanghai` 日历日；某日有效当且仅当该 `device_no` 在该日 history 记录数 ≥ 该次计算使用的 `minRecordsPerDay`；连续有效日 MUST 以请求当日为锚向前连续累计，遇无效日中断。资格计算 MUST 经 history-service HTTP 按日计数，MUST NOT 由 cash 直查 history 库。UCG 入场与值得留意喂养资格的天数判断 MUST NOT 实现在 device-service、voice-service、history-service 或客户端权威路径。

#### Scenario: 日条数达门槛计有效

- **WHEN** 某上海日 history 条数等于该次 `minRecordsPerDay`
- **THEN** 该日 MUST 计为有效喂养日

#### Scenario: 不足门槛中断连续

- **WHEN** 从今日往回扫描时某日条数小于 `minRecordsPerDay`
- **THEN** 该日 MUST NOT 计入，且连续计数 MUST 在该日中断

#### Scenario: 资格合成宿主为 cash

- **WHEN** 计算 UCG 或值得留意喂养资格
- **THEN** 连续天数与 `qualified` 合成 MUST 在 cash-service 完成，history MUST 仅返回按日条数

### Requirement: 取数窗口 MUST 等于所需连续天数

调用 history 按日统计时，请求天数 MUST 等于本次合成所需的 `requiredDays`（单场景）；MUST NOT 无必要请求大于该阈值的固定窗口（例如仅为缓冲而固定 14）。若一次请求需服务多个场景，窗口 MUST 为各场景 `requiredDays` 的最大值且 MUST NOT 更大。

#### Scenario: UCG 仅拉 requiredDays

- **WHEN** 计算 `ucg_entry` 且配置 `requiredDays=7`
- **THEN** 对 history 的按日统计请求天数 MUST 为 `7`

#### Scenario: 值得留意仅拉 requiredDays

- **WHEN** 计算 `care_alert_entry` 且配置 `requiredDays=2`
- **THEN** 对 history 的按日统计请求天数 MUST 为 `2`

### Requirement: 各场景 MUST 使用各自配置的连续天数与日条数门槛

场景 `ucg_entry` 与 `care_alert_entry` MUST 分别持久化 `requiredDays` 与 `minRecordsPerDay`；合成某一场景资格时 MUST 只使用该场景配置，MUST NOT 强制两场景共用同一日条数门槛。VIP 与功能权益 MUST NOT 改变任一场景的 `qualified`。

#### Scenario: 两场景门槛可不同

- **WHEN** `ucg_entry` 的 `minRecordsPerDay=10` 且 `care_alert_entry` 的 `minRecordsPerDay=5`
- **THEN** 同一设备同一日的 history 条数=6 时，该日对值得留意 MUST 计有效，对 UCG MUST NOT 计有效

#### Scenario: UCG 合格判据读配置

- **WHEN** `ucg_entry.requiredDays=7` 且连续有效日 ≥ 7
- **THEN** UCG eligibility MUST `qualified=true` 且 `requiredDays=7`
