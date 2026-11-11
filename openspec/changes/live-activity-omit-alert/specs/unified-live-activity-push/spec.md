## ADDED Requirements

### Requirement: Live Activity APNs start MUST include aps.alert; update and end MUST omit it

The push-service MUST build ActivityKit Live Activity APNs bodies (`apns-push-type=liveactivity`) so that `event=start` includes `aps.alert` and `event=update`/`event=end` omit `aps.alert`. For `event=start` (including first push-to-start and same-dispatch invalid-token restart start), `aps` MUST include an `alert` object with at least a non-empty title or body; if the internal op supplies empty alert title and body, the assembler MUST apply a non-empty default title/body so presentation is not dropped. For `event=update` and `event=end`, `aps` MUST NOT contain an `alert` key, even if the internal op carries alert title/body fields. Lock-screen content MUST continue to rely on `content-state` (and start `attributes` / end `dismissal-date` as today).

`event=start` MUST 含 `aps.alert`（含 invalid→restart；空文案时组装层 MUST 默认兜底）；`update`/`end` MUST NOT 含 `alert`。锁屏内容继续靠 content-state。

#### Scenario: Start includes alert

- **WHEN** push sends Live Activity `event=start` via push-to-start
- **THEN** the APNs request body `aps` MUST include `alert` with non-empty title or body
- **AND** MUST include `attributes-type` / `attributes` / `content-state` as today

#### Scenario: Update omits alert

- **WHEN** Refresh fans out a non-empty items list and push sends Live Activity `event=update`
- **THEN** the APNs request body `aps` MUST include `content-state` and MUST NOT include `alert`

#### Scenario: End omits alert

- **WHEN** Refresh fans out empty items and push sends Live Activity `event=end`
- **THEN** the APNs request body MUST omit `alert` and MUST keep existing end/`dismissal-date` semantics

#### Scenario: Start with empty caller alert still has default alert

- **WHEN** push sends `event=start` and the op alert title and body are both empty
- **THEN** the APNs `aps.alert` MUST still contain a non-empty default title or body

### Requirement: Live Activity Refresh SHALL supply alert copy for upsert fanout

When history-service Refresh schedules Live Activity fanout for a non-empty items list (`action=upsert` or equivalent), it MUST set companion `AlertTitle`/`AlertBody` (e.g. fixed title「宝宝动态」and body from the existing alert-body helper) so that when push resolves the send to `event=start`, presentation alert copy is available. Empty items `end` fanout MUST remain without alert fields. MUST NOT change ContentState item projection rules solely due to alert copy. Whether those fields appear in the APNs JSON is governed by the start-vs-update/end assembler rule above.

非空 Refresh upsert MUST 填充伴随 title/body 供 start 使用；end 无 alert 字段；items 投影规则不变；APNs 是否写出 alert 由组装层按 event 决定。

#### Scenario: Non-empty Refresh supplies alert fields

- **WHEN** Refresh builds a non-empty items list and calls the push Live Activity fanout
- **THEN** the fanout op MUST carry non-empty alert title/body for potential start presentation
- **AND** ContentState `items` MUST still be built as before
