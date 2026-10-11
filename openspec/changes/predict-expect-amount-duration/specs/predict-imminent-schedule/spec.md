## ADDED Requirements

### Requirement: Pending items MAY carry optional expect fields through Redis and MQ

The `PUT /device/api/predict/imminent/pending` event item schema SHALL accept optional fields `expectKind`, `expectDurationSec`, `expectAmount`, `expectUnit`, and `expectLeafTitle` in addition to `eventId`, `nextAt`, and `title`. On successful full-replace sync, the voice service MUST persist these fields on the Redis pending item when present, and MUST copy them onto the delayed fire message payload alongside `deviceNo`, `eventId`, `nextAt`, and `title`. Redis match for stale fire messages MUST continue to use `eventId`+`nextAt` only (expect MUST NOT participate in match). Missing expect fields MUST NOT cause sync rejection; invalid expect combinations MUST NOT reject the whole sync (push path MAY ignore invalid expect). The existing `PUT /device/api/predict/imminent/pending` path MAY accept these optional omitempty fields without introducing a new API version.

pending 项 MUST 接受可选 expect 字段；成功整表替换后 MUST 写入 Redis 并复制到延时载荷；对账仍仅 `eventId`+`nextAt`；缺省/非法 expect MUST NOT 导致整表同步失败；允许同路径可选加字段、不强制 v+1。

#### Scenario: Sync stores duration expect

- **WHEN** 客户端提交含 `expectKind=duration` 与正 `expectDurationSec` 的事件且同步成功
- **THEN** Redis 该宝宝待办中对应项 MUST 保留上述 expect 字段，且随后发布的延时消息 MUST 携带相同字段

#### Scenario: Sync stores amount expect

- **WHEN** 客户端提交含 `expectKind=amount`、正 `expectAmount`、单位与可选叶子名的事件且同步成功
- **THEN** Redis 与延时消息 MUST 保留这些 expect 字段

#### Scenario: Omitting expect still syncs

- **WHEN** 客户端仅提交 `eventId`/`nextAt`/`title`（无 expect）
- **THEN** 同步 MUST 成功，行为与扩展前一致（待办可调度）
