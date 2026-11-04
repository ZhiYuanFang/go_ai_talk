## ADDED Requirements

### Requirement: Imminent Live Activity start ContentState SHALL carry optional structured expect fields

When Go fans out an Imminent Live Activity **start**, ContentState MUST continue to include `title`, `phase`, and `nextAt` (unix seconds). When the fire/pending payload has a usable structured expect, ContentState MUST also include: `expectKind` set to `duration` or `amount`; for duration, `expectDurationSec` only when the value is &gt; 0; for amount, `expectAmount` as a floating JSON number only when &gt; 0, and `expectUnit` only when non-empty after trim. The server MUST NOT write a preformatted Chinese expect label into ContentState for this path, MUST NOT require `expectLeafTitle` in ContentState, and MUST NOT add `logo` / `color` / `logoFile`. When expect is absent or unusable, ContentState MUST omit those keys (MUST NOT write zero placeholders that imply a real expect).

Imminent Live Activity **start** 时，ContentState MUST 仍含 `title`、`phase`、`nextAt`（unix 秒）。当开火/pending 载荷有可用结构化 expect 时，ContentState MUST 另含：`expectKind` 为 `duration` 或 `amount`；duration 仅当 `expectDurationSec`&gt;0 时写入该键；amount 仅当 `expectAmount`&gt;0 时以浮点 JSON number 写入，且 `expectUnit` 仅 trim 非空时写入。本路径 MUST NOT 写入成品中文预计文案，MUST NOT 要求 ContentState 含 `expectLeafTitle`，MUST NOT 加入 `logo`/`color`/`logoFile`。无可用 expect 时 MUST 省略上述键，MUST NOT 写 0 占位冒充有效预计。

#### Scenario: Duration expect on Imminent start

- **WHEN** Go 扇出 Imminent start 且开火载荷 `expectKind=duration`、`expectDurationSec=5400`
- **THEN** ContentState MUST 含 `expectKind=duration` 与 `expectDurationSec=5400`
- **AND** MUST NOT 含成品「预计1时30分钟」类 label 键（除非另有与本需求无关的既有 alert 字段）

#### Scenario: Amount expect on Imminent start

- **WHEN** 开火载荷 `expectKind=amount`、`expectAmount=120.5`、`expectUnit=ml`
- **THEN** ContentState MUST 含上述三键且 amount 为浮点语义
- **AND** MUST NOT 要求写入 `expectLeafTitle`

#### Scenario: No expect omits keys

- **WHEN** 开火载荷无 expect 或 duration/amount 不可用
- **THEN** ContentState MUST 省略 `expectKind` / `expectDurationSec` / `expectAmount` / `expectUnit`

### Requirement: Imminent Live Activity title SHALL be the base sentence without expect suffix

When Go fans out an Imminent Live Activity **start**, the `title` written into Live Activity Attributes and ContentState (and any Live Activity alert title/body that reuses that string) MUST be the base sentence of the form `{nick}要{eventName}了` (nick from device profile with existing fallback, eventName from pending/fire `title`). That Live Activity title MUST NOT include the Chinese expect suffix produced for ordinary notifications (e.g. segments starting with `，预计这次约`). Ordinary visible `PushByBizType` / non-LA alert bodies MAY continue to append that expect suffix. Structured expect fields in ContentState remain the sole LA carrier for duration/amount display on the lock screen.

Imminent Live Activity **start** 时，写入 LA Attributes / ContentState 的 `title`（以及复用该串的 LA AlertTitle/AlertBody）MUST 为基句 `{昵称}要{事件名}了`（昵称与现网画像回落一致，事件名取 pending/fire 的 `title`）。该 LA 标题 MUST NOT 含普通通知用的中文预计后缀（如以「，预计这次约」开头的片段）。普通可见 `PushByBizType` / 非 LA 通知正文 MAY 继续追加该后缀。锁屏 duration/amount 展示仅以 ContentState 结构化 expect 字段为载体。

#### Scenario: LA title omits expect Chinese suffix

- **WHEN** 开火载荷含可用 duration expect，且普通通知 alert 会拼出「，预计这次约…」后缀
- **THEN** LA ContentState.`title`（及 Attributes.`title`）MUST 等于 `{昵称}要{事件名}了`
- **AND** MUST NOT 包含「，预计这次约」子串
- **AND** ContentState 仍 MUST 按结构化规则携带 `expectKind` / `expectDurationSec`（或 amount 维字段）

#### Scenario: Ordinary push may keep long alert

- **WHEN** 同上开火且 LA 已启用
- **THEN** 发往 `PushByBizType` 的可见正文 MAY 仍为基句 + 预计中文后缀
- **AND** MUST NOT 因此要求把该长文案写入 LA `title`
