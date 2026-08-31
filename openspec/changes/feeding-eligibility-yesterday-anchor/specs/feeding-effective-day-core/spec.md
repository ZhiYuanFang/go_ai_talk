## MODIFIED Requirements

### Requirement: 有效喂养日判定 MUST 独立于场景门槛且仅由 cash-service 执行

系统 MUST 在 **cash-service** 提供与场景解耦的有效日判定：使用 `Asia/Shanghai` 日历日；某日有效当且仅当该 `device_no` 在该日 history 记录数 ≥ 该次计算使用的 `minRecordsPerDay`；连续有效日 MUST 以 **上海昨日** 为锚向前连续累计（`days[0]` MUST 为昨天），遇无效日中断；**请求当日（今日）MUST NOT** 计入有效日或连续 streak。资格计算 MUST 经 history-service HTTP 按日计数，MUST NOT 由 cash 直查 history 库。UCG 入场与值得留意喂养资格的天数判断 MUST NOT 实现在 device-service、voice-service、history-service 或客户端权威路径。

#### Scenario: 日条数达门槛计有效

- **WHEN** 某上海已闭合日 history 条数等于该次 `minRecordsPerDay`
- **THEN** 该日 MUST 计为有效喂养日

#### Scenario: 不足门槛中断连续

- **WHEN** 从昨天往回扫描时某日条数小于 `minRecordsPerDay`
- **THEN** 该日 MUST NOT 计入，且连续计数 MUST 在该日中断

#### Scenario: 今日不计入

- **WHEN** 请求时刻为上海某日且当日 history 条数已 ≥ `minRecordsPerDay`
- **THEN** 该「今日」MUST NOT 增加 `effectiveDays`，连续扫描 MUST 自昨天起算

#### Scenario: 跨日立即合格

- **WHEN** 上海日切换后，以昨天为锚的连续有效日 ≥ 该场景 `requiredDays`
- **THEN** 当日资格请求 MUST 可返回 `qualified=true`（按请求日缓存键 miss 后重算即可），MUST NOT 依赖额外后台 ticker

#### Scenario: 资格合成宿主为 cash

- **WHEN** 计算 UCG 或值得留意喂养资格
- **THEN** 连续天数与 `qualified` 合成 MUST 在 cash-service 完成，history MUST 仅返回按日条数

### Requirement: 取数窗口 MUST 等于所需连续天数

调用 history 按日统计时，请求天数 MUST 等于本次合成所需的 `requiredDays`（单场景）；窗口 MUST 覆盖自 **上海昨日** 起往前共 `requiredDays` 个已闭合日历日，MUST NOT 包含今日；MUST NOT 无必要请求大于该阈值的固定窗口（例如仅为缓冲而固定 14）。若一次请求需服务多个场景，窗口 MUST 为各场景 `requiredDays` 的最大值且 MUST NOT 更大。history 实现 MUST 在数据库侧按上海日聚合 `COUNT`（或等价），MUST NOT 为资格统计拉取窗口内全部 history 行的 `start_time` 列表再在应用层逐条累加。

#### Scenario: UCG 仅拉 requiredDays 已闭合日

- **WHEN** 计算 `ucg_entry` 且配置 `requiredDays=7`
- **THEN** 对 history 的按日统计请求天数 MUST 为 `7`，且返回序列 `days[0]` MUST 为上海昨天

#### Scenario: 值得留意仅拉 requiredDays 已闭合日

- **WHEN** 计算 `care_alert_entry` 且配置 `requiredDays=2`
- **THEN** 对 history 的按日统计请求天数 MUST 为 `2`，且窗口实质为上海昨天与前天（不含今日）

#### Scenario: 按日 COUNT 返回

- **WHEN** history 处理 `feeding-day-stats` 且窗口内某日有多条记录
- **THEN** 响应该日 MUST 为聚合后的单条 `{date, count}`，MUST NOT 要求调用方再按原始行重算

## ADDED Requirements

### Requirement: 各场景门槛语义不变且 care_alert 为促活门槛

场景 `ucg_entry` 与 `care_alert_entry` MUST 分别持久化并使用各自的 `requiredDays` 与 `minRecordsPerDay`；合成某一场景资格时 MUST 只使用该场景配置，MUST NOT 强制两场景共用同一日条数门槛。VIP 与功能权益 MUST NOT 改变任一场景的 `qualified`。`care_alert_entry` 的连续有效日门槛 MUST 视为功能解锁 / 促活门槛，MUST NOT 被定义为「为值得留意数据准确性而要求昨日有发生」的继承语义；排除今日后 `requiredDays=2` 几何上对应「昨天+前天」连续有效，属阈值结果而非旧客户端闸门的等价替换说明义务。

#### Scenario: 两场景门槛可不同

- **WHEN** `ucg_entry` 的 `minRecordsPerDay=10` 且 `care_alert_entry` 的 `minRecordsPerDay=5`
- **THEN** 同一设备同一已闭合日的 history 条数=6 时，该日对值得留意 MUST 计有效，对 UCG MUST NOT 计有效

#### Scenario: UCG 合格判据读配置

- **WHEN** `ucg_entry.requiredDays=7` 且以昨天为锚的连续有效日 ≥ 7
- **THEN** UCG eligibility MUST `qualified=true` 且 `requiredDays=7`

#### Scenario: care_alert 满两闭合日合格

- **WHEN** `care_alert_entry.requiredDays=2` 且上海昨天与前天均为有效喂养日
- **THEN** care-alert eligibility MUST `qualified=true`
