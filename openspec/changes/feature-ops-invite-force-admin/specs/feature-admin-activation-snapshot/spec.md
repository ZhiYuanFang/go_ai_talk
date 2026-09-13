## ADDED Requirements

### Requirement: Admin 按功能查看当前开通快照

cash Admin MUST 提供按 `featureId` 查询的开通快照接口（及管理页入口），返回该功能当前持久化开通主体列表（可分页），每项至少包含：主体标识（`deviceNo` 或 `wxId`）、最近开通方式（若适用）、是否未过期有效、到期时间或剩余时效（`expires_at=0` 表示永久；预测条数类可用条数代替时效）。快照 MUST 来自当前权益/条数权威表（方案 A），MUST NOT 要求新建事件账本。页面 MUST 注明不含 VIP 旁路覆盖、非完整历史事件流。

#### Scenario: 成长轨迹列出账号权益

- **WHEN** 运维打开 `growth_trajectory_predict` 详情快照
- **THEN** 列表 MUST 来自 `feature_user_entitlement`（或等价），含 wx、方式、是否有效与到期/永久

#### Scenario: 值得留意列出设备权益

- **WHEN** 运维打开 `care_alert_smart_remind` 详情快照
- **THEN** 列表 MUST 来自 `feature_entitlement`，含 deviceNo、方式、是否有效与到期/永久

#### Scenario: 预测列出条数

- **WHEN** 运维打开 `prediction_unlock` 详情快照
- **THEN** 列表 MUST 展示设备与永久条数（或等价字段），MUST NOT 伪造权益型到期字段为必需

#### Scenario: 已过期仍可见

- **WHEN** 某主体权益 `expires_at` 已早于当前时间且行仍存在
- **THEN** 快照 MUST 仍可列出并标记为未在有效期内
