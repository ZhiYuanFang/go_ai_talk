## ADDED Requirements

### Requirement: 预测开通数量 MUST 经功能列表暴露且不得下发事件 ID

对预测类功能，合成 catalog 对应项 MUST 暴露非负整数 `allowedCount`（无记录则为 0）。响应 MUST NOT 包含由服务端挑选的预测事件 ID 列表。客户端负责按本地预测列表顺序解锁前 N 项。系统 MUST NOT 将独立 `GET .../prediction/allowed-count` 作为客户端必调接口（可省略）。

#### Scenario: 列表项含 allowedCount 而非事件 ID

- **WHEN** 客户端请求合成功能目录且存在预测类功能
- **THEN** 该项 MUST 含 `allowedCount`，且 MUST NOT 含事件 ID 数组作为开通清单

#### Scenario: 无开通记录时为 0

- **WHEN** 设备从未获得预测开通数量增量
- **THEN** 该项 `allowedCount` MUST 为 `0`

### Requirement: Admin 对预测功能 MUST 只配置开通数量语义

与预测解锁相关的可配置授予物 MUST 表达为开通数量（增量），MUST NOT 要求运营配置具体事件 ID。

#### Scenario: 授予数量后列表可读到

- **WHEN** 经支付/邀请码/广告等路径为某 `device_no` 增加预测开通数量 N
- **THEN** 随后合成目录中该项 `allowedCount` MUST 反映累加后的权威值（无扣减前提下 ≥ 原值 + N）

### Requirement: allowedCount 权威在 MySQL 且热读经 Redis

`allowedCount` 权威 MUST 存于 MySQL（`feature_allowed_count`）。读路径 MUST 可经 Redis 缓存加速（产品已确认）；履约或 Admin 变更后 MUST 失效或写穿。Redis 访问 MUST 经 `cachekit` 与 platform 键 builder。

#### Scenario: 履约后数量可见

- **WHEN** 功能 SKU 支付成功且 `grant_kind` 为预测数量增量
- **THEN** 该 `device_no` 的 `allowedCount` MUST 增加对应 `grant_quantity`，且后续 catalog 读 MUST 看到更新后的值（缓存失效后）
