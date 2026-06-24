## MODIFIED Requirements

### Requirement: 帖子读写 API SHALL 可选接受 viewer 坐标并返回 distanceMeters

下列接口 MUST 支持可选 viewer 坐标；当 viewer 与目标帖均有坐标时，响应 MUST 含 `distanceMeters` 字符串（米，纯数字字符串，如 `"1234"`），MUST NOT 含距离文案标签；否则 MUST 省略该字段：

- `GET /ucg/app/api/feed/recommend` — query `lat`、`lng`（可选）
- `GET /ucg/app/api/feed/following` — query `lat`、`lng`（可选）
- `GET /ucg/app/api/posts/{id}` — query `lat`、`lng`（可选）
- `PUT /ucg/app/api/posts/{id}` — body 可选 `lat`、`lng`（更新帖坐标）

#### Scenario: 推荐 Feed item 含距离

- **WHEN** `GET /ucg/app/api/feed/recommend?lat=31.2&lng=121.5` 且帖有坐标
- **THEN** 该帖 JSON MUST 含 `distanceMeters` 且值为米数字符串

#### Scenario: 关注 Feed item 含距离

- **WHEN** `GET /ucg/app/api/feed/following?lat=31.2&lng=121.5&page=1` 且帖有坐标
- **THEN** 该帖 JSON MUST 含 `distanceMeters` 且值为米数字符串

#### Scenario: 无坐标省略距离

- **WHEN** `GET /ucg/app/api/posts/{id}` 无 lat/lng query 或帖无坐标
- **THEN** 响应 MUST NOT 含 `distanceMeters` 字段

#### Scenario: 更新帖坐标

- **WHEN** `PUT /ucg/app/api/posts/{id}` body 含新 lat/lng
- **THEN** 服务 MUST 更新存储坐标并同步 Redis GEO 与 post snapshot

## ADDED Requirements

### Requirement: 关注 Feed API SHALL 保持 page/total 分页并可选接受 viewer 坐标

`GET /ucg/app/api/feed/following` MUST 继续要求有效 `X-Internal-Wx-Id`。MUST 接受 query `page`（从 1）、`pageSize`（默认 20，最大 50）及可选 `lat`、`lng`。响应 MUST 为 `{ list, total, page, pageSize }`（`UcgPageRes`），**MUST NOT** 含 `nextCursor` 或 `hasMore`。本 endpoint 为既有路由的 query 扩展，**MUST NOT** 作为新 App API 计入 usage 统计。

#### Scenario: 关注 Feed 分页契约不变

- **WHEN** `GET /ucg/app/api/feed/following?page=2&pageSize=20`
- **THEN** 响应 MUST 含 `list`、`total`、`page`、`pageSize`
- **AND** MUST NOT 含 `nextCursor` 或 `hasMore`

#### Scenario: 关注 Feed 可选坐标 query

- **WHEN** `GET /ucg/app/api/feed/following?page=1&lat=31.2&lng=121.5`
- **THEN** 服务 MUST 接受 lat/lng 并用于列表 item 距离计算
- **AND** 分页字段 MUST 仍为 `{ list, total, page, pageSize }`

#### Scenario: 关注 Feed 需身份

- **WHEN** 未带有效 `X-Internal-Wx-Id` 请求 following feed
- **THEN** 服务 SHALL 返回未授权错误
