## ADDED Requirements

### Requirement: History mutation MUST invalidate imminent pending Redis

When history create, update, or delete for a device succeeds, the twin Go service MUST delete (or empty) that device’s Redis imminent pending authority key as part of the same post-commit side effects that clear Imminent Live Activities. Subsequent delayed fire consumers MUST continue to treat a missing or empty pending list as no-match (MUST NOT push solely because an old delayed message arrived). This invalidation MUST NOT require a `PUT /device/api/predict/imminent/pending` call. Active-eta Redis keys are out of scope for this requirement unless a later change says otherwise.

宝宝 history 增/删/改成功后，Go MUST 删除（或清空）该 deviceNo 的 Redis imminent pending 权威键，与 Imminent LA 清屏同属提交后副作用。延时开火消费 MUST 继续在 pending 缺失/为空时视为不匹配而不推。作废 MUST NOT 依赖再次 `PUT …/pending`。本要求默认不动 active-eta Redis。

#### Scenario: Pending key gone after history write

- **WHEN** 某 deviceNo 在 Redis 仍有非空 imminent pending，随后 history 对该宝宝写入成功
- **THEN** 该 deviceNo 的 imminent pending Redis 键 MUST 被删除或等价清空
- **AND** 之后到达的、仅对账旧 pending 的延时消息 MUST NOT 因旧列表再推 Imminent

#### Scenario: Pending PUT still authoritative for new schedules

- **WHEN** history 已清空 pending 后客户端再次 `PUT …/predict/imminent/pending` 写入新列表
- **THEN** Go MUST 按既有整表替换与延时投递规则接受新日程
- **AND** MUST NOT 因该 PUT 执行 Imminent 全员清屏（清屏挂 history）
