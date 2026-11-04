## ADDED Requirements

### Requirement: Timing Live Activity end SHALL dismiss immediately and MUST NOT wipe startTime

When Go fans out a Timing Live Activity **end** (history closed by setting endTime, or the history row is deleted), the APNs liveactivity payload MUST include `event=end` and `dismissal-date` set to a unix timestamp in the **past** relative to send time, so the system removes the Lock Screen activity immediately. Omitting `dismissal-date` is forbidden. ContentState MUST include `phase` indicating ended and MUST NOT set `startTime` to 0 when a real start timestamp is known; if `eventName`/`startTime` are available they MUST be included so a delayed dismiss does not decode as epoch-zero elapsed time.

Timing Live Activity **end**（history 写入 endTime 闭合，或删除该行）扇出时，APNs liveactivity 载荷 MUST 含 `event=end` 且 `dismissal-date` 为发送时刻之前的 unix 秒，使锁屏立刻摘条。MUST NOT 省略 `dismissal-date`。ContentState MUST 含表示已结束的 `phase`；已知真实开始时间时 MUST NOT 把 `startTime` 写成 0；若有 `eventName`/`startTime` MUST 带上，以免 dismiss 延迟时被解成从 epoch 计时。

#### Scenario: Close timing dismisses immediately

- **WHEN** Go 将一条 history 更新为已设置 endTime 并扇出 Timing end
- **THEN** APNs 体 MUST 含 `dismissal-date` 且其值 MUST 早于发送时刻
- **AND** ContentState MUST 含 `phase` 表示 ended
- **AND** 若该行 `startTime` > 0，ContentState 的 `startTime` MUST 等于该 unix 秒而非 0

#### Scenario: Delete history ends Timing

- **WHEN** Go 成功删除一条 history（含未闭合计时行）
- **THEN** Go MUST 向该 deviceNo 绑定用户扇出同一 activity id（`timing:{deviceNo}:{historyId}`）的 Timing Live Activity end
- **AND** 该 end MUST 同样携带过去的 `dismissal-date`

#### Scenario: End without dismissal-date is forbidden

- **WHEN** push-service 组装 Timing `event=end` 的 APNs JSON
- **THEN** `aps` MUST 包含键 `dismissal-date`
- **AND** 业务未传 `DismissalDate` 时组装层 MUST 填入过去 unix 秒，MUST NOT 省略该键

### Requirement: Timing Live Activity start Attributes SHALL carry eventName and startTime

When Go fans out a Timing Live Activity **start**, the fixed attributes object MUST include `eventName` (catalog display name when the history row name is empty but `eventId` is valid) and `startTime` as unix seconds when known. ContentState MUST continue to include the same fields. Clients MAY read attributes when ContentState is still the Push-to-Start default empty state.

Timing **start** 固定 Attributes MUST 含 `eventName`（行内名为空且 eventId 有效时用事件主档名）及已知的 unix 秒 `startTime`。ContentState MUST 继续含同名字段。Push-to-Start 未套用 ContentState 时客户端 MAY 读 attributes。

#### Scenario: Timing start attributes include eventName

- **WHEN** Go 因未闭合计时 history 落库成功而扇出 Timing start，且能解析出非空展示名
- **THEN** 固定 Attributes MUST 含键 `eventName` 且值为该展示名
- **AND** ContentState MUST 同样含该 `eventName`

#### Scenario: Empty row name uses catalog

- **WHEN** 该 history 的 `eventName` 为空且 `eventId` > 0，事件主档有标准名
- **THEN** Attributes 与 ContentState 的 `eventName` MUST 使用该主档名，MUST NOT 以空串作为唯一展示名
