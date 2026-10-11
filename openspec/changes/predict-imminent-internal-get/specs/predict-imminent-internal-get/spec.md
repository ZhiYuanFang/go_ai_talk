## ADDED Requirements

### Requirement: System SHALL expose internal GET of imminent pending for Python

The system SHALL provide `GET /voice/internal/api/predict/imminent/pending` that returns `{ "events": [...] }` for a `deviceNo` from the existing imminent Redis pending key. The route MUST be registered on voice-service behind the same internal secret middleware as `GET /voice/internal/api/predict/active-eta` (`X-Device-Gateway-Internal-Secret`). When the key is missing, empty, or contains illegal JSON, the response MUST be an empty `events` list (not an error solely for emptiness or corrupt payload). Field names MUST align with the App sync event item schema (`eventId`, `nextAt`, `title`, and optional `expectKind`, `expectDurationSec`, `expectAmount`, `expectUnit`, `expectLeafTitle`). This API MUST be read-only: MUST NOT publish delayed MQ, MUST NOT send push, and MUST NOT alter the App PUT sync semantics.

系统 MUST 提供内部 GET 返回 `{ events }`；须挂内部密钥中间件；空键/非法 JSON MUST 返回空列表；字段与 App PUT 对齐；MUST NOT 改推送或 PUT。

#### Scenario: Get returns stored events

- **WHEN** Redis 已有一条 imminent 待办且内部 GET 该 deviceNo（密钥合法）
- **THEN** 响应 events MUST 包含对应 eventId/nextAt/title 及已存 expect* 字段

#### Scenario: Missing key empty list

- **WHEN** Redis 无该宝宝 imminent pending 键
- **THEN** GET MUST 成功返回空 events 列表

#### Scenario: No push side effects

- **WHEN** 内部 GET 成功
- **THEN** 系统 MUST NOT 因此发布延时消息或可见推送
