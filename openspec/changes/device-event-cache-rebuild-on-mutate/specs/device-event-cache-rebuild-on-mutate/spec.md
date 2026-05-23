## ADDED Requirements

### Requirement: 事件表变更后 Redis 缓存必须从数据库重建

当 `device-service` 成功执行对 `event` 表的插入、更新或删除后，系统 SHALL 使用数据库当前全量事件行（含 `logo` 与 `color`）重建 Redis 中的事件选项缓存，且 SHALL NOT 通过先调用 `ListEvents`（可能仅返回变更前缓存）再写回的方式刷新缓存。

#### Scenario: 更新事件 color 后缓存含新色值

- **WHEN** 管理员通过 API 成功更新某事件的 `color`
- **THEN** 随后对 `ListEvents` 或等价读路径的调用在缓存命中时 SHALL 返回包含新 `color` 的该事件行

#### Scenario: 更新事件 logo 后缓存含新 path

- **WHEN** 管理员成功上传并更新某事件的 `logo` 路径
- **THEN** Redis 事件选项快照中该事件的 `logo` 字段 SHALL 与数据库一致

#### Scenario: 新增事件后缓存含新行

- **WHEN** 管理员成功新增一条事件记录
- **THEN** 随后缓存命中时 SHALL 包含该新事件

#### Scenario: 删除事件后缓存不含已删行

- **WHEN** 管理员成功删除一条事件记录
- **THEN** 随后缓存命中时 SHALL NOT 包含已删除的事件 id

### Requirement: 写后刷新失败可观测

若重建 Redis 缓存失败，系统 SHALL 记录警告级别日志且 SHALL NOT 将写库事务回滚（写库已成功）。

#### Scenario: Redis 不可用时的行为

- **WHEN** 数据库写入成功但 `RebuildEventCache` 因 Redis 错误失败
- **THEN** 系统 SHALL 记录可观测警告日志
- **AND** API 仍可对客户端返回写库成功（与现网语义一致）
