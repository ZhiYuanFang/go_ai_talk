# Spec Delta

## ADDED Requirements

### Requirement: cash-service SHALL 支持按宝宝号迁移对机权益分区键

cash-service MUST 提供仅内部/管理可信调用方可访问的迁移接口：将以 `device_no` 为分区键的对机权益与次数数据从 `oldDeviceNo` 更新为 `newDeviceNo`（至少覆盖 `feature_entitlement` 与 `feature_allowed_count`）。`newDeviceNo` 上若已存在冲突唯一键，MUST 拒绝或按实现文档化的合并策略处理且 MUST 可观测；默认实现 MUST 在冲突时失败并保持可重试语义清晰。支付订单等审计快照中的 `device_no` MUST NOT 被本需求强制跟迁。

#### Scenario: 对机权益随号迁移

- **WHEN** 可信调用方请求将旧宝宝号的对机权益迁移到新宝宝号且无唯一键冲突
- **THEN** 旧号下对机权益与次数 MUST 在新号下生效，旧号 MUST 不再持有这些对机权益行
