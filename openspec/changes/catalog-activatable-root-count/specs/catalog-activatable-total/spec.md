## REMOVED Requirements

### Requirement: device 提供非叶子事件计数

**Reason**: 预测可激活天花板改为事件字典一级根数量；「有子节点才计入」会排除无子一级根（如洗澡），与产品不符。且「非叶子」与 voice 追问术语冲突。

**Migration**: 使用本变更「device 提供一级根事件计数」；内部 HTTP 路径可暂保留 `/device/internal/api/event/non-leaf-count`，响应 `count` 语义改为一级根个数。

## ADDED Requirements

### Requirement: device 提供一级根事件计数

device 服务 MUST 提供可被 cash 调用的契约，返回事件字典中**一级根**事件数量（`parent_id = 0` 的事件行数，MUST 包含无子节点的一级根）。计数权威 MUST 为 device 事件字典，MUST NOT 要求 cash 直查 device 库表。该计数 MUST NOT 再等同于「至少有一个子事件的节点数」。

#### Scenario: 统计一级根含无子根

- **WHEN** 事件字典存在 8 个 `parent_id = 0` 的事件，其中部分无子节点
- **THEN** 返回的 count MUST 为 8

#### Scenario: 子事件不计入

- **WHEN** 某一级根下存在若干子事件
- **THEN** 这些子事件 MUST NOT 增加 count；仅该一级根本身计 1

## MODIFIED Requirements

### Requirement: catalog 聚合 totalActivatableCount

cash 合成功能目录时，对 `prediction_unlock` 项 MUST 调用 device **一级根**事件计数并写入响应字段 `totalActivatableCount`。永久合成 `allowedCount` MUST 仍为 defaultFree+permanentDelta（MUST NOT 因 VIP 或邀请改为 -1）。当计数契约失败时，系统 MUST NOT 返回误导性的正数天花板冒充成功（宜省略或 0），并记录日志。客户端预测锁槽位与「已全部激活」判断 MUST 与一级根集合对齐（本仓外 Flutter 同步）。

#### Scenario: catalog 含一级根天花板

- **WHEN** App 拉取 feature catalog 且 device 一级根计数成功返回 N
- **THEN** `prediction_unlock.totalActivatableCount` MUST 等于 N

#### Scenario: 无子一级根计入天花板

- **WHEN** 字典含无子一级根「洗澡」且 device 计数成功
- **THEN** 「洗澡」MUST 反映在 `totalActivatableCount` 中（即计入 N）
