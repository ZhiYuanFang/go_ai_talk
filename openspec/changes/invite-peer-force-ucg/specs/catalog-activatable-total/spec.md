## ADDED Requirements

### Requirement: device 提供非叶子事件计数

device 服务 MUST 提供可被 cash 调用的契约，返回事件字典中非叶子事件数量（至少有一个子事件以该事件为 `parent_id` 的节点数）。计数权威 MUST 为 device 事件字典，MUST NOT 要求 cash 直查 device 库表。

#### Scenario: 统计非叶子

- **WHEN** cash 请求非叶子计数且字典中存在有子节点的事件
- **THEN** 返回的 count 等于这些非叶子节点个数

### Requirement: catalog 聚合 totalActivatableCount

cash 合成功能目录时，对 `prediction_unlock` 项 MUST 调用 device 非叶子计数并写入响应字段 `totalActivatableCount`。永久合成 `allowedCount` MUST 仍为 defaultFree+permanentDelta（本变更下 MUST NOT 因 VIP 或邀请改为 -1）。当计数契约失败时，系统 MUST NOT 返回误导性的正数天花板冒充成功（宜省略或 0），并记录日志。

#### Scenario: catalog 含天花板

- **WHEN** App 拉取 feature catalog 且 device 计数成功返回 N
- **THEN** `prediction_unlock.totalActivatableCount` 等于 N
