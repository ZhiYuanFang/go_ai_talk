## ADDED Requirements

### Requirement: Timing Live Activity start payload SHALL include eventId and unix-second startTime without visuals

When Go fans out a Timing Live Activity **start**, the fixed ActivityKit attributes object MUST include `eventId` set from the persisted history row’s event id (same identity the Flutter event catalog uses). ContentState `startTime` MUST be a unix timestamp in **seconds** (not milliseconds). The attributes and ContentState MUST NOT include `logo`, `color`, `logoFile`, or remote image URL fields; the client resolves visuals locally from `eventId`.

Timing Live Activity **start** 扇出时，固定 Attributes MUST 含来自该 history 行的 `eventId`（与 Flutter 事件目录 id 对齐）。ContentState 的 `startTime` MUST 为 unix **秒**（非毫秒）。Attributes 与 ContentState MUST NOT 含 `logo` / `color` / `logoFile` 或远程图 URL；视觉由客户端按 `eventId` 本地解析。

#### Scenario: Timing start carries eventId

- **WHEN** Go 因未闭合计时 history 落库成功而扇出 Timing start，且该行 `eventId` > 0
- **THEN** 推送固定 Attributes MUST 含键 `eventId` 且值等于该 history 的事件 id

#### Scenario: startTime is unix seconds

- **WHEN** Timing start ContentState 含 `startTime`
- **THEN** 其值 MUST 表示自 Unix epoch 起的秒数，MUST NOT 使用毫秒

#### Scenario: No logo or color in Timing push

- **WHEN** Go 构造 Timing Live Activity start 或 update 载荷
- **THEN** Attributes 与 ContentState MUST NOT 包含 `logo`、`color`、`logoFile` 或等价远程图字段
