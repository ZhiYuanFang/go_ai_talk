## ADDED Requirements

### Requirement: Feed Redis 键 SHALL 经 cachekit builder 登记且禁止业务层字面量

UCG Feed 读写的 Redis 键 MUST 在 `internal/platform/cachekit/keys_ucg.go` 登记下列 builder，业务与 controller MUST 经 `cachekit.Cache`（含 `WithObserver`）访问，MUST NOT 使用 `g.Redis()` 或键字面量拼接：

- `UCGFeedGeoKey()` → `ucg:feed:geo`（GEO）
- `UCGRecommendScoreKey()` → `ucg:recommend:score`（ZSET）
- `UCGPostSnapshotKey(postId)` → `ucg:post:snapshot:{postId}`（STRING JSON）
- `UCGProfileSnapshotKey(wxId)` → `ucg:profile:snapshot:{wxId}`（STRING JSON）
- `UCGUserLikedPostsKey(wxId)` → `ucg:user:{wxId}:liked-posts`（SET）
- `UCGFeedSessionKey(sessionId)` → `ucg:feed:session:{sessionId}`（SET，TTL 30min）

#### Scenario: 新增键仅经 platform 登记

- **WHEN** 实现需读写 Feed GEO 索引
- **THEN** 代码 MUST 调用 `UCGFeedGeoKey()` 且 MUST NOT 在 `internal/services/**` 出现 `ucg:feed:geo` 字面量

### Requirement: 帖子 snapshot SHALL 缓存 Feed 展示所需字段及 server-only 坐标

`ucg:post:snapshot:{postId}` JSON MUST 含客户端 Feed 所需字段：`id`、`content`、媒体 CDN URL、`authorWxId`、`likeCount`、`ipLocation`、`publishedAt`、`mediaType`。MAY 含 server-only `lat`、`lng` 供距离计算；该坐标 MUST NOT 出现在 App HTTP 响应体。

帖子 published 或更新时 MUST 写入/刷新 snapshot；unpublished/delete MUST DEL 键。

#### Scenario: snapshot 命中时 Feed 不查帖表

- **WHEN** `ListRecommendFeed` 组装 postId 列表且 snapshot 存在
- **THEN** 系统 MUST 从 Redis JSON 填充 `UcgPostItem` 字段且 MUST NOT 对该帖执行 `ucg_post` 单条 SELECT

#### Scenario: 坐标仅服务端使用

- **WHEN** snapshot 含 `lat`/`lng` 且 API 返回帖子
- **THEN** 响应 JSON MUST NOT 含 `lat`/`lng` 字段

### Requirement: 作者 profile snapshot SHALL 缓存公开 profile 字段

`ucg:profile:snapshot:{wxId}` JSON MUST 含 `wxId`、`nickname`、`bio`、`avatarUrl`、`avatarThumbnailUrl`。帖子 publish/update 时 MUST 刷新作者 snapshot；Feed 读路径 MUST 批量 GET snapshot 填充 `author`。

#### Scenario: Feed 作者信息来自 snapshot

- **WHEN** Feed 返回帖含 `authorWxId`
- **THEN** `author` 字段 MUST 优先来自 `ucg:profile:snapshot:{wxId}` 且 MUST NOT 在循环内逐条查 `ucg_profile`

### Requirement: 用户点赞 SET SHALL 支撑 likedByMe 批量判定

`ucg:user:{wxId}:liked-posts` SET member MUST 为 `postId` 字符串。like MUST SADD；unlike MUST SREM。Feed 读路径对当前页 postIds MUST 批量 SISMEMBER（pipeline）填充 `likedByMe`，MUST NOT 查 `ucg_post_like` 表。

#### Scenario: Feed likedByMe 走 Redis

- **WHEN** 已登录用户请求推荐 Feed 且 page 含 20 帖
- **THEN** 系统 MUST 经一次 pipeline SISMEMBER 判定 liked 状态且 MUST NOT 对每帖查询 MySQL like 表

### Requirement: Feed session SET SHALL 防止 cursor 分页重复下发

`ucg:feed:session:{sessionId}` MUST 为 SET，member 为已返回 `postId`；TTL MUST 为 30min（可配置）。每次 Feed 页返回前 MUST SADD 本页 postIds；候选集 MUST 排除 session 中已见 postId。

#### Scenario: 同 session 翻页无重复

- **WHEN** 客户端用合法 `nextCursor` 连续请求两页且 session 未过期
- **THEN** 两页 `list` 中 postId MUST 互斥（无交集）

#### Scenario: 刷新生成新 session

- **WHEN** 客户端请求 Feed 且不传 `cursor`
- **THEN** 系统 MUST 生成新 `sessionId` 且 MUST NOT 复用旧 session 的 seen SET
