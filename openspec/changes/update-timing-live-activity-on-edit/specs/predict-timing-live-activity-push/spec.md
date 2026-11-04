## ADDED Requirements

### Requirement: Open timing history update SHALL fan out Timing Live Activity ContentState update

When Go successfully updates a history row that remains open (`endTime` unset), it MUST fan out a Timing Live Activity **update** to bound users using the same activity id as start (`timing:{deviceNo}:{historyId}`). The update payload MUST include ContentState `phase` of timing, `eventName` when a display name is available, and `startTime` as unix seconds when `startTime` > 0. MUST NOT set `startTime` to 0. MUST NOT end and restart the activity solely because `eventId` changed; Lock Screen event artwork MAY remain from original start attributes.

Go 成功更新仍未闭合（`endTime` 未设置）的 history 时，MUST 按 start 同一 activity id 向绑定用户扇出 Timing Live Activity **update**。ContentState MUST 含 `phase=timing`；有展示名则含 `eventName`；`startTime>0` 则含 unix 秒 `startTime`，MUST NOT 写成 0。MUST NOT 仅因 `eventId` 变化而 end 再 start；锁屏事件图 MAY 沿用 start 时 Attributes。

#### Scenario: Edit start time updates lock screen timer

- **WHEN** App 将一条未闭合 history 的 `startTime` 改为另一 unix 秒且更新成功
- **THEN** Go MUST 对该 `timing:{deviceNo}:{historyId}` 扇出 `action=update`
- **AND** ContentState `startTime` MUST 为新的 unix 秒

#### Scenario: Edit event name updates lock screen title

- **WHEN** App 更新未闭合 history 的展示名（或 eventId 对应主档名）且行仍未闭合
- **THEN** ContentState MUST 含非空 `eventName`（主档优先）
- **AND** MUST NOT 因此结束该 Timing 活动

#### Scenario: Close still only ends

- **WHEN** 同一次 Update 将 `endTime` 设为非 0
- **THEN** Go MUST 扇出 Timing end
- **AND** MUST NOT 再对该行发 update
