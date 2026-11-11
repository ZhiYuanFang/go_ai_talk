## ADDED Requirements

### Requirement: Live Activity APNs payloads MUST omit aps.alert

When the push-service builds an ActivityKit Live Activity APNs request body (`apns-push-type=liveactivity`) for `event` of `start`, `update`, or `end`, the JSON under `aps` MUST NOT contain an `alert` key (including title, body, or sound). Lock-screen / Dynamic Island presentation MUST continue to rely solely on `content-state` (and `attributes` / `attributes-type` for start, plus `dismissal-date` for end as today). Callers MAY still pass empty or non-empty alert title/body fields on the internal Live Activity op; the APNs assembler MUST ignore them and MUST NOT emit `aps.alert`. This requirement MUST apply to all sends through the shared Live Activity APNs path, including upsert→update, push-to-start start, end, and same-dispatch invalid-token restart start.

经共享 Live Activity APNs 路径发出的 start/update/end 请求体 MUST NOT 含 `aps.alert`；锁屏展示继续只依赖 content-state（及 start 的 attributes、end 的 dismissal-date）；内部 op 的 alert 字段 MUST 被组装层忽略。含 invalid→restart 的 start。

#### Scenario: Upsert update has content-state without alert

- **WHEN** Refresh fans out a non-empty items list and push sends Live Activity `event=update`
- **THEN** the APNs request body `aps` MUST include `content-state` and MUST NOT include `alert`

#### Scenario: Push-to-start has no alert

- **WHEN** push sends Live Activity `event=start` via push-to-start (including after invalid update restart)
- **THEN** the APNs request body `aps` MUST include `attributes-type` / `attributes` / `content-state` as today and MUST NOT include `alert`

#### Scenario: End remains without alert

- **WHEN** Refresh fans out empty items and push sends Live Activity `event=end`
- **THEN** the APNs request body MUST still omit `alert` and MUST keep existing end/`dismissal-date` semantics

### Requirement: Live Activity Refresh MUST NOT supply alert copy for fanout

When history-service Refresh schedules Live Activity fanout for a non-empty items list (`action=upsert` or equivalent), it MUST NOT set non-empty companion notification title/body intended for `aps.alert` (e.g. MUST NOT assign the previous fixed title「宝宝动态」or `liveActivityAlertBody` result into the push client op). Empty items `end` fanout MUST remain without alert fields as today. MUST NOT change ContentState item projection rules solely due to omitting alert copy.

非空列表 Refresh 扇出 MUST NOT 再填伴随通知 title/body；空列表 end 保持无 alert；MUST NOT 因省略 alert 文案而改 items 投影规则。

#### Scenario: Non-empty Refresh omits alert fields

- **WHEN** Refresh builds a non-empty items list and calls the push Live Activity fanout
- **THEN** the fanout op MUST NOT carry non-empty alert title/body for APNs alert presentation
- **AND** ContentState `items` MUST still be built as before
