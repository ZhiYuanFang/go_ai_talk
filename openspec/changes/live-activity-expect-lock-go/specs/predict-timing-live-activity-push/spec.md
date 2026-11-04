## ADDED Requirements

### Requirement: Timing Live Activity ContentState SHALL support optional expectEndAt

When Go fans out a Timing Live Activity **start** or **update**, ContentState MAY include `expectEndAt` as a unix timestamp in **seconds** representing the predicted end instant. When `expectEndAt` is known and &gt; 0, the server MUST include it; when unknown, the server MUST omit the key (MUST NOT write `0` as a sentinel that clients treat as epoch). ContentState MUST NOT include a preformatted Chinese subtitle string for the expect-end line. Attributes and ContentState MUST still MUST NOT include `logo` / `color` / `logoFile`. Timing **end** payloads are not required to carry `expectEndAt`.

Timing Live Activity **start** 或 **update** 时，ContentState MAY 含 `expectEndAt`（unix **秒**，预计结束时刻）。已知且 &gt;0 时 MUST 写入；未知时 MUST 省略该键，MUST NOT 写 `0` 占位。MUST NOT 写入成品中文「预计 HH:mm 结束」字符串。Attributes/ContentState 仍 MUST NOT 含 logo/color。**end** 不要求带 `expectEndAt`。

#### Scenario: Update carries expectEndAt

- **WHEN** Go 对已有 Timing 活动扇出 update 且已知 `expectEndAt=1730000000`
- **THEN** ContentState MUST 含 `expectEndAt` 且值为该 unix 秒

#### Scenario: Unknown ETA omits expectEndAt

- **WHEN** Timing start 时尚无预计结束
- **THEN** ContentState MUST 省略 `expectEndAt`
- **AND** MUST 仍含既有 `eventName` / `startTime` / `phase` 契约字段
