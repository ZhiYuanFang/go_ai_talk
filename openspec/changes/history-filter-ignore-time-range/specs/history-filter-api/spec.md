## ADDED Requirements

### Requirement: filter 支持 ignoreTimeRange 强制忽略时间窗

系统 SHALL 在既有 `GET /device/history/api/filter` 上接受可选查询参数 `ignoreTimeRange`。未传或为假时，时间筛选语义 MUST 与现网一致（`startTime`/`endTime` > 0 才施加对应条件，=0 跳过）。为真时，系统 MUST **完全忽略**请求中的 `startTime` 与 `endTime`（即使均为正数），MUST NOT 对其施加任何时间 WHERE 条件；`deviceNo`、`eventIds`、`remark`、`limit` 与按 id 倒序等既有规则 MUST 保持不变。本参数为 additive 扩展，MUST NOT 要求客户端升级后才能调用原筛选行为。

#### Scenario: 默认不忽略时间（兼容旧客户端）

- **WHEN** 客户端调用 filter 且未传 `ignoreTimeRange`（或显式为假），并传入 `startTime`/`endTime` 正数
- **THEN** 系统 SHALL 按现网规则对 `start_time` 施加对应上下界过滤

#### Scenario: ignoreTimeRange 为真时忽略已填时间

- **WHEN** 客户端以 `ignoreTimeRange` 为真，且同时传入非零 `startTime` 与/或 `endTime` 调用 filter
- **THEN** 系统 SHALL 返回结果时 MUST NOT 使用上述时间值作为过滤条件
- **AND** 其他条件（如 `eventIds`、`remark`、`limit`）SHALL 仍按既有规则生效

#### Scenario: ignoreTimeRange 为真且无事件/备注时仍按 limit 收口

- **WHEN** 客户端以 `ignoreTimeRange` 为真、`eventIds` 为空、`remark` 为空调用 filter
- **THEN** 系统 SHALL 在无时间条件下按设备维度返回最多 `limit` 条（默认/上限规则与现网一致）的历史记录（id 倒序）

#### Scenario: local/remote/canary 透传一致

- **WHEN** history 以 remote 或 canary 模式调用 filter
- **THEN** `ignoreTimeRange` 为真时 MUST 经 HTTP 契约透传至 history-service，本地与远程语义 MUST 一致
