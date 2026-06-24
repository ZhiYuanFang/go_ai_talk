## MODIFIED Requirements

### Requirement: App HTTP API SHALL expose UCG REST under /ucg/app/api

ucg-service SHALL implement REST endpoints (also reachable via gateway proxy) including: profile get/update, feed recommend/following, posts CRUD, media presign, follow/unfollow, likes, comments, conversations and messages list。

**推荐 Feed**（`GET /ucg/app/api/feed/recommend`）MUST 使用 **cursor 分页**（见 ADDED Requirement），**不适用** `{ total, page, pageSize }` 契约。

除推荐 Feed 与评论列表外，其他列表分页 MUST 仍使用 `page`（从 1 开始）与 `pageSize`（默认 20，最大 50），响应 MUST 包含 `{ list, total, page, pageSize }`。

评论列表 `GET /ucg/app/api/posts/{id}/comments` SHALL **不适用**上述分页契约：SHALL 单次返回该帖评论全量列表。`page`/`pageSize` 查询参数 SHALL 被忽略或废弃，MUST NOT 再驱动服务端 `OFFSET` 分页。

#### Scenario: 推荐 Feed cursor 分页

- **WHEN** `GET /ucg/app/api/feed/recommend?pageSize=20&lat=31.2&lng=121.5`
- **THEN** 响应 SHALL 仅含 `status=2` 的帖子，且 SHALL 含 `list`、`hasMore`，MAY 含 `nextCursor`；MUST NOT 含 `total`

#### Scenario: 我的动态含全状态

- **WHEN** 作者 `GET /ucg/app/api/posts/mine`
- **THEN** 响应 SHALL 含 draft/pending/rejected/published 本人帖子

#### Scenario: 关注 Feed 需身份

- **WHEN** 未带有效 `X-Internal-Wx-Id` 请求 following feed
- **THEN** 服务 SHALL 返回未授权错误

## ADDED Requirements

### Requirement: v2 创建帖 API SHALL 支持可选坐标

系统 MUST 提供 `POST /ucg/app/api/v2/posts`，请求体 MUST 兼容 v1 创建帖字段并 MAY 含可选 `lat`、`lng`（WGS84 十进制度）。成功响应 MUST 与 v1 创建帖相同结构（`UcgPostItem`）。该 endpoint MUST 计入 gateway-app App API 使用统计。

#### Scenario: v2 创建带坐标

- **WHEN** 客户端 `POST /ucg/app/api/v2/posts` 含合法 body 与 `lat`/`lng`
- **THEN** 服务 MUST 持久化坐标（MySQL 或等价）且在 publish 后 GEOADD 与 snapshot 含坐标

#### Scenario: v2 创建无坐标

- **WHEN** 客户端 `POST /ucg/app/api/v2/posts` 不含 lat/lng
- **THEN** 服务 MUST 成功创建且帖 MUST NOT 进入 GEO 索引

#### Scenario: v2 创建计入 usage

- **WHEN** 经 gateway-app 的 `POST /ucg/app/api/v2/posts` 返回 HTTP 2xx
- **THEN** gateway-app MUST 将该请求计入 App API 使用统计

### Requirement: 帖子读写 API SHALL 可选接受 viewer 坐标并返回 distanceMeters

下列接口 MUST 支持可选 viewer 坐标；当 viewer 与目标帖均有坐标时，响应 MUST 含 `distanceMeters` 字符串（米，纯数字字符串，如 `"1234"`），MUST NOT 含距离文案标签；否则 MUST 省略该字段：

- `GET /ucg/app/api/feed/recommend` — query `lat`、`lng`（可选）
- `GET /ucg/app/api/posts/{id}` — query `lat`、`lng`（可选）
- `PUT /ucg/app/api/posts/{id}` — body 可选 `lat`、`lng`（更新帖坐标）

#### Scenario: Feed item 含距离

- **WHEN** `GET /ucg/app/api/feed/recommend?lat=31.2&lng=121.5` 且帖有坐标
- **THEN** 该帖 JSON MUST 含 `distanceMeters` 且值为米数字符串

#### Scenario: 无坐标省略距离

- **WHEN** `GET /ucg/app/api/posts/{id}` 无 lat/lng query 或帖无坐标
- **THEN** 响应 MUST NOT 含 `distanceMeters` 字段

#### Scenario: 更新帖坐标

- **WHEN** `PUT /ucg/app/api/posts/{id}` body 含新 lat/lng
- **THEN** 服务 MUST 更新存储坐标并同步 Redis GEO 与 post snapshot

### Requirement: 推荐 Feed cursor 参数 SHALL 冻结 session 上下文

`GET /ucg/app/api/feed/recommend` MUST 接受 query：

- `cursor`（可选，opaque）
- `pageSize`（可选，默认 20，最大 50）
- `lat`、`lng`（可选，仅首屏无 cursor 时生效）

有 `cursor` 时 MUST 忽略新的 lat/lng。响应 `nextCursor` MUST 在 `hasMore=true` 时存在。

#### Scenario: 翻页携带 cursor

- **WHEN** 客户端用上一页 `nextCursor` 请求第二页
- **THEN** 系统 MUST 使用 cursor 内冻结坐标与 session 且 MUST 返回不重复 postId

#### Scenario: 下拉刷新无 cursor

- **WHEN** 客户端不传 `cursor` 请求 Feed
- **THEN** 系统 MUST 创建新 feed session 且 MAY 使用本次 lat/lng
