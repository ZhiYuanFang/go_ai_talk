## REMOVED Requirements

### Requirement: Server SHALL clear all Imminent Live Activities on pending sync

**Reason**: 清屏应以 history 权威写为准；挂在 pending PUT 会在预测重报时误清、在历史已变更时又可能滞后。

**Migration**: 改由 history 增/删/改落库成功路径清 Imminent，并 Del Redis pending；见新增 Requirement「Server SHALL clear all Imminent Live Activities on history mutation」。`PUT …/predict/imminent/pending` 仅保留整表替换与延时投递。

## ADDED Requirements

### Requirement: Server SHALL clear all Imminent Live Activities on history mutation

When the twin Go **history** service successfully persists a history create, update, or delete for a device (including instantaneous events with non-zero end time, and any successful update regardless of which fields changed), it MUST send Live Activity end (or equivalent dismiss) for all tracked Imminent Live Activities to bound users of that baby/device, and MUST NOT end Timing Live Activities as part of this clear. The voice handling of `PUT /device/api/predict/imminent/pending` MUST NOT perform this all-Imminent clear. As part of the same clear, the service MUST delete that device’s Redis imminent pending authority key so previously scheduled delayed fires no longer match pending. After an Imminent Live Activity end (history clear or per-slot replace), the push domain MUST delete persisted `token_kind=activity` rows for that activity id (same push-layer rule as Timing end; see `predict-timing-live-activity-push`); Push-to-Start tokens MUST NOT be deleted by this cleanup.

宝宝 history 增/删/改持久化成功时（含瞬时事件、任意成功 Update），Go MUST 向绑定用户结束全部已跟踪 Imminent Live Activity，MUST NOT 借此结束 Timing。`PUT …/predict/imminent/pending` MUST NOT 再执行该全员清屏。清屏同时 MUST 删除该 deviceNo 的 Redis imminent pending 权威键。Imminent end（含清屏与单槽换新）后，push 域 MUST 删除该 `activityId` 的 `token_kind=activity` 行（与 Timing end 同一 push 层规则）；MUST NOT 删除 Push-to-Start。

#### Scenario: History create clears Imminent and pending

- **WHEN** history 成功新增一条记录（含瞬时事件）
- **THEN** Go MUST 向该 deviceNo 绑定用户发出清空/结束全部 Imminent Live Activity 的推送（或等价）
- **AND** MUST NOT 因此结束 Timing Live Activities
- **AND** MUST 删除该 deviceNo 的 Redis imminent pending 键

#### Scenario: History update clears Imminent and pending

- **WHEN** history 任意成功更新（含仅改备注、闭合、改开始时间或事件名等）
- **THEN** Go MUST 同样清全部 Imminent 并删除 Redis imminent pending
- **AND** MUST NOT 因此结束 Timing

#### Scenario: History delete clears Imminent and pending

- **WHEN** history 删除某行成功
- **THEN** Go MUST 同样清全部 Imminent 并删除 Redis imminent pending

#### Scenario: Pending PUT no longer clears Imminent

- **WHEN** 客户端提交 `PUT …/predict/imminent/pending`（空或非空列表）
- **THEN** Go MUST NOT 仅因该 PUT 向绑定用户扇出 Imminent 全员 end/清屏
- **AND** 既有写 Redis / 延时投递语义 MUST 保持

#### Scenario: Activity token rows removed after Imminent end

- **WHEN** Go 对某 `activityId` 扇出 Imminent `event=end`（history 清屏或开火换槽）
- **THEN** push 域 MUST 删除该 `activityId` 下 `token_kind=activity` 的持久化行
- **AND** MUST NOT 删除该用户的 Push-to-Start 行

### Requirement: Imminent clear on history MUST be best-effort after commit

Imminent Live Activity clear and Redis pending deletion triggered by history mutation MUST run only after the history write has successfully committed, and MUST NOT roll back or fail the history write if clear or Redis delete fails. Failures MUST be logged. Timing Live Activity start/update/end on the same write path MUST remain independently best-effort as today.

由 history 变更触发的 Imminent 清屏与 Redis pending 删除 MUST 仅在 history 写成功提交之后执行；清屏或 Del Redis 失败 MUST NOT 回滚或导致该次 history 写失败，MUST 记日志。同路径 Timing LA 仍按既有 best-effort。

#### Scenario: Clear failure does not fail history write

- **WHEN** history 落库成功但 Imminent end 或 pending Del 失败
- **THEN** 该次 history 写路径 MUST 仍视为落库成功
- **AND** MUST 输出可观测失败日志
