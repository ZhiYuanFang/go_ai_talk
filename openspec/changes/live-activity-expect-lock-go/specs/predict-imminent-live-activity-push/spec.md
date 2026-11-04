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
