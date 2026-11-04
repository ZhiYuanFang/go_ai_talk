## ADDED Requirements

### Requirement: Imminent Live Activity payload SHALL use unix-second nextAt and MUST NOT carry visuals

When Go fans out an Imminent Live Activity **start**, fixed attributes MUST continue to include `eventId` and `nextAt`, and ContentState MUST include `nextAt`, each as unix timestamps in **seconds** (not milliseconds). Attributes and ContentState MUST NOT include `logo`, `color`, `logoFile`, or remote image URL fields; the client resolves visuals locally from `eventId`.

Imminent Live Activity **start** 时，固定 Attributes MUST 继续含 `eventId` 与 `nextAt`，ContentState MUST 含 `nextAt`，且均为 unix **秒**（非毫秒）。Attributes 与 ContentState MUST NOT 含 `logo` / `color` / `logoFile` 或远程图 URL；视觉由客户端按 `eventId` 本地解析。

#### Scenario: nextAt is unix seconds

- **WHEN** Imminent start 载荷含 Attributes 或 ContentState 的 `nextAt`
- **THEN** 其值 MUST 为 unix 秒

#### Scenario: eventId present without logo color

- **WHEN** Go 向已注册 token 的用户扇出 Imminent start 且事件 id 有效
- **THEN** Attributes MUST 含 `eventId`
- **AND** MUST NOT 同时下发 `logo` / `color` / `logoFile`
