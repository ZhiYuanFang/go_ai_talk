## ADDED Requirements

### Requirement: System SHALL accept full-replace active ETA sync from the App

The system SHALL expose an authenticated App HTTP API `PUT /device/api/predict/active-eta/pending` for a bound device to full-replace the baby’s in-progress timing ETA list in Redis (MUST NOT use MySQL as authority; MUST NOT require TTL expiry). The request body MUST use `deviceNo` and `items`; the response MUST include `count` (number of retained items after cleaning). Each retained item MUST have positive `rootEventId` and positive `leafEventId`, positive `startAt` and positive `expectEndAt`, and `expectEndAt >= startAt`. Empty `title` MUST NOT cause discard. Items with `expectEndAt` in the past MUST still be retained when otherwise valid. Within one request, duplicate `leafEventId` MUST keep the last occurrence. An empty `items` list MUST clear the Redis key. Concurrent sync for the same device MUST use a short SetNX lock; failure to acquire MUST return a busy client error (e.g. 「同步繁忙」) without partial write. This API MUST NOT publish delayed MQ messages and MUST NOT send push notifications. Imminent pending sync semantics MUST remain unchanged. This App route MUST NOT be counted in gateway-app API usage statistics (maintenance denylist).

系统 MUST 提供 App `PUT /device/api/predict/active-eta/pending` 全量替换进行中 ETA 并写入 Redis；body 为 `deviceNo`+`items`，响应含 `count`；保留项 MUST 同时满足 root/leaf/start/end 为正且 end≥start；空 title 与已过期 end 不因此丢弃；同 leaf 后写覆盖；空列表清空；SetNX 短锁忙则 4xx；MUST NOT 发延时 MQ/推送；MUST NOT 改变 imminent pending；MUST NOT 计入 App usage 统计。

#### Scenario: Replace list

- **WHEN** 客户端先后两次成功提交同一 deviceNo 的不同 items
- **THEN** Redis 中该宝宝进行中 ETA MUST 仅等于第二次内容

#### Scenario: Empty clears

- **WHEN** 客户端提交空 items 且成功
- **THEN** 该宝宝 active-eta Redis MUST 为空且响应 count 为 0

#### Scenario: Discard invalid keep valid

- **WHEN** 同一次 PUT 含一条非法项（leaf≤0 或 expectEndAt 小于 startAt）与一条合法项
- **THEN** 同步 MUST 成功且 Redis 仅含合法项

#### Scenario: Dedup by leafEventId

- **WHEN** 同一次 PUT 含两条相同 leafEventId 的合法项
- **THEN** Redis 中该 leaf MUST 仅保留后一条

#### Scenario: No push side effects

- **WHEN** 同步成功写入一条 ETA
- **THEN** 系统 MUST NOT 因此发布预测临近延时消息或可见推送

#### Scenario: Sync busy

- **WHEN** 同 deviceNo 同步锁已被占用
- **THEN** 新请求 MUST 失败返回同步繁忙类错误且 MUST NOT 改写 Redis 列表

### Requirement: System SHALL expose internal GET of active ETAs for Python

The system SHALL provide `GET /voice/internal/api/predict/active-eta` that returns `{ "items": [...] }` for a `deviceNo` from Redis. When the key is missing or empty, the response MUST be an empty `items` list (not an error solely for emptiness). Field names MUST align with the App sync item schema (`rootEventId`, `leafEventId`, `historyId`, `title`, `startAt`, `expectEndAt`) so Python can match identifiers.

系统 MUST 提供内部 GET 返回 `{ items }`；空键 MUST 返回空列表；字段与 App 同步对齐。

#### Scenario: Get returns stored items

- **WHEN** Redis 已有一条 ETA 且内部 GET 该 deviceNo
- **THEN** 响应 items MUST 包含对应 startAt/expectEndAt 与 rootEventId/leafEventId 字段

#### Scenario: Missing key empty list

- **WHEN** Redis 无该宝宝 active-eta 键
- **THEN** GET MUST 成功返回空 items 列表
