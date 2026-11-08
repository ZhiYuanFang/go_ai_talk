## ADDED Requirements

### Requirement: History mutation MUST invalidate imminent pending Redis before Live Activity Refresh

When history create, update, or delete for a device succeeds, the twin Go service MUST delete (or empty) that device’s Redis imminent pending authority key as a post-commit side effect, and MUST do so before invoking the unified Live Activity Refresh for that device. Subsequent delayed fire consumers MUST continue to treat a missing or empty pending list as no-match (MUST NOT push ordinary `predict_imminent` solely because an old delayed message arrived). This invalidation MUST NOT require a `PUT /device/api/predict/imminent/pending` call. Active-eta Redis keys are out of scope unless a later change says otherwise.

宝宝 history 增/删/改成功后，Go MUST 删除（或清空）该 deviceNo 的 Redis imminent pending 权威键，且 MUST 在统一 Live Activity Refresh 之前完成。延时开火消费 MUST 继续在 pending 缺失/为空时视为不匹配而不发普通预测推送。作废 MUST NOT 依赖再次 PUT pending。

#### Scenario: Pending key gone after history write

- **WHEN** 某 deviceNo 在 Redis 仍有非空 imminent pending，随后 history 对该宝宝写入成功
- **THEN** 该 deviceNo 的 imminent pending Redis 键 MUST 被删除或等价清空
- **AND** 之后到达的、仅对账旧 pending 的延时消息 MUST NOT 因旧列表再发普通预测推送

#### Scenario: Invalidate happens before Refresh

- **WHEN** history write triggers both pending invalidation and Live Activity Refresh
- **THEN** Refresh MUST observe the cleared pending (predict segment empty) rather than the pre-CUD list

### Requirement: Successful pending sync MUST trigger unified Live Activity Refresh

After a successful full-replace write of the device’s imminent pending Redis (including empty clear), the system MUST invoke the unified Live Activity Refresh for that `deviceNo`. This MUST NOT replace or skip delayed MQ publish rules already required by「同步成功后 MUST 按提前量投递延时叫醒消息」. This MUST NOT perform Imminent-only Live Activity end clear as the sync side effect.

imminent pending 整表替换成功后 MUST 触发统一 LA Refresh；MUST NOT 因此取消延时投递；MUST NOT 以 Imminent end 清屏作为 PUT 副作用。

#### Scenario: Non-empty PUT refreshes LA

- **WHEN** client successfully PUTs a non-empty pending list
- **THEN** Redis MUST hold that list, delayed fires MUST still be scheduled per existing rules, and unified Live Activity Refresh MUST run

#### Scenario: Empty PUT refreshes LA

- **WHEN** client successfully PUTs an empty pending list
- **THEN** Redis pending MUST be empty and unified Live Activity Refresh MUST run

### Requirement: Imminent MQ fire MUST keep ordinary push and also Refresh Live Activity

When delayed imminent fire matches Redis pending and passes existing business gates, the system MUST retain ordinary visible `predict_imminent` push behavior for bound users, and MUST also invoke unified Live Activity Refresh for the device. The fire path MUST NOT end-then-start a separate Imminent Live Activity. Delayed MQ topology, Ack/no-requeue, and Redis reconciliation rules remain as specified by existing schedule requirements.

延时开火在闸门通过后 MUST 保留普通可见推送，并 MUST 调用统一 LA Refresh；MUST NOT Imminent 先 end 再 start。

#### Scenario: Fire dual channel

- **WHEN** fire matches pending and gates pass for a device with bound users
- **THEN** system MUST attempt ordinary predict_imminent visible push AND MUST Refresh unified Live Activity

#### Scenario: Stale fire still skips both when pending mismatch

- **WHEN** fire payload does not match current Redis pending
- **THEN** system MUST NOT send ordinary predict push and MUST NOT be required to Refresh solely for that stale message (Ack as today)
