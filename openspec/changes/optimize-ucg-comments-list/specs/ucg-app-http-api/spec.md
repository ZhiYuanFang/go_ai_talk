## MODIFIED Requirements

### Requirement: App HTTP API SHALL expose UCG REST under /ucg/app/api

ucg-service SHALL implement REST endpoints (also reachable via gateway proxy) including: profile get/update, feed recommend/following, posts CRUD, media presign, follow/unfollow, likes, comments, conversations and messages list.

除评论列表外，分页 MUST 使用 `page`（从 1 开始）与 `pageSize`（默认 20，最大 50），响应 MUST 包含 `{ list, total, page, pageSize }`。

评论列表 `GET /ucg/app/api/posts/{id}/comments` SHALL **不适用**上述分页契约：SHALL 单次返回该帖评论全量列表（见下条 Requirement）。`page`/`pageSize` 查询参数 SHALL 被忽略或废弃，MUST NOT 再驱动服务端 `OFFSET` 分页。

#### Scenario: 推荐 Feed 分页

- **WHEN** `GET /ucg/app/api/feed/recommend?page=1&pageSize=20`
- **THEN** 响应 SHALL 仅含 `status=2` 的帖子，且 SHALL 含分页字段 `list`、`total`、`page`、`pageSize`

#### Scenario: 我的动态含全状态

- **WHEN** 作者 `GET /ucg/app/api/posts/mine`
- **THEN** 响应 SHALL 含 draft/pending/rejected/published 本人帖子

#### Scenario: 关注 Feed 需身份

- **WHEN** 未带有效 `X-Internal-Wx-Id` 请求 following feed
- **THEN** 服务 SHALL 返回未授权错误

## ADDED Requirements

### Requirement: 评论列表 SHALL 单次返回升序全量并批量填充作者 profile

`GET /ucg/app/api/posts/{id}/comments` MUST 仅对已发布（`status=2`）帖子返回评论。评论 MUST 按 `created_at` **升序**排序（最早在前、最新在列表底部）。响应 MUST 为 JSON：

- `list`：评论数组，每项 SHALL 含 `id`、`postId`、`authorWxId`、`content`、`createdAt`，且 SHOULD 含 `author`（公开 profile，与帖子作者 profile 同为实时 `ucg_profile` 语义，MUST NOT 依赖评论表快照列）
- `total`：SHALL 优先为帖子 `comment_count`；若不可用 SHALL 回退为 `len(list)`
- `truncated`：布尔；当评论数超过配置上限且仅返回部分行时为 `true`，否则为 `false`

服务 MUST NOT 对评论列表执行独立 `COUNT(*)` 查询。服务 MUST 通过单次 `ucg_profile` 批量查询（如 `wx_id IN (...)`）填充列表内所有 `author`，MUST NOT 在列表循环中逐条查询 profile（N+1）。评论列表读路径 MUST NOT 使用 Redis 缓存。

当配置 `ucg.comments.listMax`（默认 500，0 表示不限制）大于 0 且 `comment_count` 超过该值时，SHALL 仅返回按 `created_at ASC` 的前 `listMax` 条，并设 `truncated=true`；`total` SHALL 仍为完整 `comment_count`。

#### Scenario: 常规模帖单次拉取全量评论

- **WHEN** 客户端 `GET /ucg/app/api/posts/{id}/comments` 且该帖已发布、评论数不超过 `listMax`
- **THEN** 响应 SHALL 含全部评论且 `list` 按 `created_at` 升序
- **AND** `truncated` SHALL 为 `false`
- **AND** `total` SHALL 等于帖子 `comment_count`
- **AND** 每条评论的 `author` SHALL 经批量 profile 查询填充

#### Scenario: 超长帖评论截断

- **WHEN** 帖子 `comment_count` 为 600 且 `listMax=500`
- **THEN** 响应 `list` SHALL 含 500 条（最早 500 条）
- **AND** `truncated` SHALL 为 `true`
- **AND** `total` SHALL 为 600

#### Scenario: 发表评论响应可供乐观追加

- **WHEN** 客户端 `POST /ucg/app/api/posts/{id}/comments` 成功
- **THEN** 响应 SHALL 含完整评论字段及 `author`
- **AND** 客户端 SHALL 可将该条追加至本地 `list` 末尾而无需再次 GET 全列表

#### Scenario: 评论列表不使用 Redis

- **WHEN** ucg-service 处理评论列表读请求
- **THEN** 系统 MUST NOT 读写 Redis 中的评论列表或 profile 缓存键

### Requirement: 评论列表 GET SHALL 不计入 App API 使用统计

负责人已确认：`GET /ucg/app/api/posts/{id}/comments` MUST NOT 计入 gateway-app App API 使用统计。gateway-app SHALL 在 `maintenance_skip.go` 排除该 GET 路径（原始 `req.URL.Path` 为 `/ucg/app/api/posts/<postId>/comments`；归一化 apiKey 为 `GET /ucg/app/api/posts/{id}/comments`）。`POST /ucg/app/api/posts/{id}/comments`（发表评论）SHALL 仍计入统计。

#### Scenario: 成功 GET 评论列表不写入 usage

- **WHEN** 客户端成功调用 `GET /ucg/app/api/posts/{id}/comments` 经 gateway-app
- **THEN** gateway-app MUST NOT 将该请求计入 App API 使用统计

#### Scenario: 发表评论仍计入 usage

- **WHEN** 客户端成功调用 `POST /ucg/app/api/posts/{id}/comments` 经 gateway-app
- **THEN** gateway-app SHALL 将该请求计入 App API 使用统计
