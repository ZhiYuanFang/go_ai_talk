# async-cache-projection-sync Specification

## Purpose
TBD - created by archiving change async-redis-read-model-for-history-action-event-user. Update Purpose after archive.
## Requirements
### Requirement: 写入后异步更新缓存投影
系统对 `history/action/event/user` 的写操作在权威存储提交成功后 MUST 发布缓存投影事件，并由低延迟异步消费者更新 Redis 读模型；系统 MUST NOT 仅通过删除缓存完成一致性维护。

#### Scenario: 写入成功后发布投影事件
- **WHEN** 新增、修改或删除操作在数据库事务中提交成功
- **THEN** 系统 MUST 发布对应缓存投影事件并包含实体主键、操作类型与版本信息

#### Scenario: 消费者应用缓存补丁
- **WHEN** 缓存投影事件被消费者成功消费
- **THEN** 系统 MUST 按操作语义对 Redis 读模型执行增量补丁更新

### Requirement: 乱序与重复事件保护
异步缓存更新链路 MUST 具备事件幂等和版本顺序保护，确保旧事件或重复事件不会覆盖新状态。

#### Scenario: 处理重复事件
- **WHEN** 消费者收到相同事件 ID 的重复投递
- **THEN** 系统 MUST 识别为重复并跳过二次更新

#### Scenario: 处理乱序事件
- **WHEN** 消费者收到版本号低于当前缓存版本的事件
- **THEN** 系统 MUST 拒绝应用该事件并记录版本冲突指标

### Requirement: 失败补偿与可重建
系统 MUST 提供缓存投影失败重试与修复机制，保障最终一致；当异步更新持续失败时 MUST 支持按实体重建 Redis 读模型。

#### Scenario: 暂时失败自动重试
- **WHEN** 缓存补丁更新因临时错误失败
- **THEN** 系统 MUST 按重试策略重新消费并在成功后清除失败状态

#### Scenario: 持续失败进入修复流程
- **WHEN** 某实体缓存补丁达到最大重试次数仍失败
- **THEN** 系统 MUST 将该实体标记为待修复并通过重建流程恢复缓存一致性

