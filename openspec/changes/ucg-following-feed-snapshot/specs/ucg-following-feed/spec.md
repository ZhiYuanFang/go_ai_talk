## ADDED Requirements

### Requirement: 关注 Feed SHALL 经 MySQL 分页取 postId 并由 Redis snapshot 组装展示

关注 Feed 读路径 MUST 按下列顺序执行：

1. 从 MySQL `ucg_follow` 读取当前用户 followee `wxId` 列表；无关注时 MUST 返回空 `list` 与 `total=0`。
2. 从 MySQL `ucg_post` 查询 `status=2`（published）且 `author_wx_id IN (followees)` 的帖子，**MUST** 按 `published_at DESC` 排序，**MUST** 使用 `page`/`pageSize` 分页并返回准确 `total`。
3. DB 查询 MUST 仅获取组装所需 postId（及排序字段），**MUST NOT** 在关注 Feed 读路径 JOIN profile、like 或媒体表。
4. 对当前页 postId 列表 MUST 调用与推荐 Feed 共用的 snapshot 组装逻辑（post snapshot、profile snapshot、liked SET batch），**MUST NOT** 使用 `postsFromResult` 全量 MySQL 组装。
5. snapshot miss MUST best-effort 调用 `backfillPostSnapshot` 回源 MySQL 并回填 Redis（与推荐 Feed 相同语义）。
6. 本读路径 **MUST NOT** 新增 Redis 键（不得创建 author ZSET、followees SET 或 following session）。
7. 本读路径 **MUST NOT** 使用 cursor、`nextCursor`、`hasMore` 替代 `{ total, page, pageSize }`。
8. 本读路径 **MUST NOT** 使用 GEO 半径筛选或 composite score 改变关注 Feed 排序；排序权威来源 MUST 仍为 MySQL `published_at DESC`。

#### Scenario: 有关注时分页返回 published 帖

- **WHEN** 已登录用户 `GET /ucg/app/api/feed/following?page=1&pageSize=20` 且其关注的人中有 published 帖
- **THEN** 响应 MUST 含 `{ list, total, page, pageSize }`
- **AND** `list` 中每条 MUST 为 `status=2` 且作者 MUST 在当前用户 followee 集合内
- **AND** `list` 顺序 MUST 按 `published_at` 降序

#### Scenario: 组装走 Redis snapshot 而非帖表 N+1

- **WHEN** 关注 Feed 返回非空 `list` 且对应 post snapshot 已存在于 Redis
- **THEN** 系统 MUST 从 post/profile snapshot 与 liked SET 填充 item 字段
- **AND** MUST NOT 对该页帖子逐条 SELECT `ucg_post` 完整行用于展示组装

#### Scenario: snapshot miss 回源

- **WHEN** 关注 Feed 当前页某 postId 在 Redis 无 post snapshot
- **THEN** 系统 MUST 尝试 `backfillPostSnapshot` 回源并回填
- **AND** 回填成功时该帖 SHOULD 出现在 `list` 中

#### Scenario: 无关注时空列表

- **WHEN** 已登录用户未关注任何人
- **THEN** 响应 MUST 为 `{ list: [], total: 0, page, pageSize }`
- **AND** MUST NOT 查询 `ucg_post` 全表

#### Scenario: 不使用 Redis 关注 timeline

- **WHEN** 实现关注 Feed 读路径
- **THEN** 系统 MUST NOT 使用 ZUNIONSTORE 或 per-author 帖 ZSET 作为分页来源
- **AND** MUST NOT 新增 followees SET 或 author timeline 键

### Requirement: 关注 Feed SHALL 可选计算并返回 distanceMeters

当 viewer 请求含有效 `lat`/`lng` 且帖子 snapshot（或回源后）含坐标时，关注 Feed item MUST 含 `distanceMeters` 字符串（米，纯数字字符串）；否则 MUST 省略该字段。距离计算 MUST 复用推荐 Feed 的 haversine 逻辑，**MUST NOT** 将 post 坐标下发客户端。

#### Scenario: 关注 Feed item 含距离

- **WHEN** `GET /ucg/app/api/feed/following?lat=31.2&lng=121.5&page=1` 且某帖 snapshot 含坐标
- **THEN** 该帖 JSON MUST 含 `distanceMeters` 且值为米数字符串

#### Scenario: 无 viewer 坐标省略距离

- **WHEN** `GET /ucg/app/api/feed/following?page=1` 不含 lat/lng
- **THEN** 响应 item MUST NOT 含 `distanceMeters`

#### Scenario: 帖无坐标省略距离

- **WHEN** viewer 含 lat/lng 但帖无坐标
- **THEN** 该 item MUST NOT 含 `distanceMeters`
