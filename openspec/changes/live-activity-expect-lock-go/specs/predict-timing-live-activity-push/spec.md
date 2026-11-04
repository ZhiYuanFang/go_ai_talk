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

### Requirement: History Timing Live Activity update SHALL preserve expectEndAt when ETA is known

When history fans out a Timing Live Activity **update** after an open history row is edited (e.g. event name or start time change) without closing the row, the server MUST attempt to resolve `expectEndAt` for that `historyId` from the device’s active-eta Redis list (`cachekit.PredictActiveEtaKey`) and, when a matching item has `expectEndAt` &gt; 0, MUST include that value in the update ContentState together with `phase` / `eventName` / `startTime` as applicable. Because APNs replaces `content-state` as a whole, omitting a previously pushed `expectEndAt` on this path MUST be treated as a defect when the ETA is still present in Redis. If Redis is missing, corrupt, or has no matching positive `historyId` / `expectEndAt`, the update MUST omit `expectEndAt` and MUST still proceed for the other ContentState fields. This path MUST NOT import the voice service package; read-only access via platform cachekit keys is allowed.

history 在未闭合 history 行被编辑（如改事件名 / startTime）后扇出 Timing **update** 时，MUST 尝试从该宝宝 active-eta Redis（`cachekit.PredictActiveEtaKey`）按 `historyId` 解析 `expectEndAt`；匹配且 &gt;0 时 MUST 写入本次 ContentState（并尽量带上 phase/eventName/startTime）。因 APNs 整包替换 `content-state`，在 Redis 仍有该 ETA 时省略 `expectEndAt` 视为缺陷。Redis 缺失/损坏/无匹配时 MUST 省略该键且 MUST 仍推送其它 ContentState 字段。MUST NOT import `services/voice`；允许经 platform cachekit 只读键。

#### Scenario: Edit timing keeps expectEndAt from active-eta

- **WHEN** 某 Timing 已因 active-eta sync 显示副行，且 Redis 仍有该 `historyId` 的正 `expectEndAt`
- **AND** 用户经 history update 修改未闭合行的 `eventName` 或 `startTime`
- **THEN** 随后 Timing update 的 ContentState MUST 仍含同一 `expectEndAt`（或 Redis 当前值）
- **AND** MUST 仍含更新后的 `eventName` / `startTime`（在适用时）

#### Scenario: Edit without ETA still updates title fields

- **WHEN** history 扇出 Timing update 但 active-eta 无匹配条目
- **THEN** ContentState MUST 省略 `expectEndAt`
- **AND** MUST 仍更新 `phase` / 可用的 `eventName` / `startTime`
