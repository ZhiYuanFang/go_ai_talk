## ADDED Requirements

### Requirement: Imminent Live Activity end SHALL dismiss immediately and MUST NOT wipe nextAt

When Go fans out an Imminent Live Activity **end** (pending sync clear of all Imminent slots, or replacing the per-wx slot before a new start), the APNs liveactivity payload MUST include `event=end` and `dismissal-date` set to a unix timestamp in the **past** relative to send time. Omitting `dismissal-date` is forbidden. ContentState MUST include `phase` indicating cleared or ended; when the previous `nextAt`/`title` are known they MUST be included so a delayed dismiss does not decode `nextAt` as 0.

Imminent Live Activity **end**（pending 同步清全部 Imminent 槽，或换新 start 前结束该 wx 旧槽）时，APNs 载荷 MUST 含 `event=end` 且 `dismissal-date` 为发送时刻之前的 unix 秒。MUST NOT 省略 `dismissal-date`。ContentState MUST 含表示清场/结束的 `phase`；若已知旧 `nextAt`/`title` MUST 带上，以免 dismiss 延迟时 `nextAt` 被解成 0。

#### Scenario: Pending PUT end includes past dismissal-date

- **WHEN** Go 因 `PUT …/predict/imminent/pending` 向绑定用户扇出 Imminent end
- **THEN** 每条 APNs end 体 MUST 含 `dismissal-date` 且值 MUST 早于发送时刻
- **AND** MUST NOT 因此结束 Timing Live Activities

#### Scenario: Slot replace end includes past dismissal-date

- **WHEN** Go 在 Imminent start 前对该 wx 先 end 旧槽
- **THEN** 该 end MUST 同样携带过去的 `dismissal-date`

#### Scenario: Imminent end without dismissal-date is forbidden

- **WHEN** push-service 组装 Imminent `event=end` 的 APNs JSON
- **THEN** `aps` MUST 包含键 `dismissal-date`
- **AND** 业务未传 `DismissalDate` 时组装层 MUST 填入过去 unix 秒

### Requirement: Imminent Live Activity start Attributes SHALL carry title

When Go fans out an Imminent Live Activity **start**, the fixed attributes object MUST include `title` set to the same display sentence as ContentState `title` (the imminent alert body). ContentState MUST continue to include `title`. Clients MAY read attributes.title when ContentState title is empty after Push-to-Start.

Imminent **start** 固定 Attributes MUST 含 `title`，与 ContentState `title` 为同一句展示文案。ContentState MUST 继续含 `title`。Push-to-Start 后 ContentState 标题为空时客户端 MAY 读 attributes.title。

#### Scenario: Imminent start attributes include title

- **WHEN** Go 向已注册 token 的用户扇出 Imminent start 且 alert 非空
- **THEN** 固定 Attributes MUST 含键 `title` 且值等于该 alert
- **AND** ContentState MUST 含相同 `title`
