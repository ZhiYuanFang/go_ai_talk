## ADDED Requirements

### Requirement: Imminent pending SHALL accept optional parentEventName

When accepting `PUT` imminent pending items, the service MUST accept an optional `parentEventName` (string). Missing or blank after trim MUST mean no parent. Persistence and subsequent GET / voice internal projections MUST round-trip the trimmed non-empty value when present. The service MUST NOT resolve the name to an event id and MUST NOT use it for scheduling, dedup, or in-progress gates (those remain `eventId`-based). imminent pending MUST 接受可选 parentEventName 并原样投影。

#### Scenario: Legacy body without parent name

- **WHEN** 旧客户端 PUT pending 项无 `parentEventName`
- **THEN** 服务 MUST 接受并调度，该项父名视为空

#### Scenario: Child item with root name

- **WHEN** 新客户端 PUT 子 `eventId` 且 `parentEventName` 为非空根事件名（如「吃奶」）
- **THEN** 存储与后续 GET / internal 读 MUST 能读回该 trim 后的 `parentEventName`

### Requirement: expectLeafTitle MUST be dropped end-to-end

The service MUST NOT require `expectLeafTitle`. Legacy clients MAY still send the JSON key; the service MUST NOT fail the sync because of it. The service MUST NOT persist `expectLeafTitle` to Redis pending, MUST NOT copy it onto delayed fire payloads, and MUST NOT append it in push alert amount suffixes. Amount expect copy MUST use formatted amount and optional unit only. expectLeafTitle MUST 全链路丢弃；alert 不拼 leaf。

#### Scenario: Legacy body with expectLeafTitle still succeeds

- **WHEN** 旧客户端 PUT pending 项含 `expectLeafTitle` 及其它合法字段
- **THEN** 服务 MUST 接受并调度，MUST NOT 因该字段失败

#### Scenario: Amount alert without leaf

- **WHEN** 延时开火拼 amount 预计后缀且 `expectAmount>0`
- **THEN** 后缀 MUST 含格式化数量与（若有）单位，MUST NOT 拼接叶子标题

### Requirement: Active-eta SHALL allow empty or zero expectEndAt

When accepting active-eta items, the service MUST treat omitted, null, or non-positive `expectEndAt` (including `0`) as no estimated end, and MUST still retain the item when `rootEventId` / `leafEventId` / `startAt` are otherwise valid. Positive `expectEndAt` MUST remain valid and MUST satisfy `expectEndAt >= startAt`. Downstream Live Activity / projections MUST tolerate missing end (omit ContentState key; MUST NOT write `0` as sentinel). active-eta MUST 允许无预计结束（含 0）；正值兼容旧客户端。

#### Scenario: Merge-root active without end

- **WHEN** PUT active-eta 项含合法 root/leaf/start 且 `expectEndAt` 为 0 或省略
- **THEN** 服务 MUST 接受该进行中项且 MUST NOT 因无正结束而丢弃该项或拒绝整次 PUT

#### Scenario: Legacy positive end still accepted

- **WHEN** 旧客户端 PUT active-eta 项含正 `expectEndAt` 且 `>= startAt`
- **THEN** 服务 MUST 按既有语义保留该项
