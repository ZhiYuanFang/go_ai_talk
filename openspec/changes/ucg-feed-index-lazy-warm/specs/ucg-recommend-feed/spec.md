## ADDED Requirements

### Requirement: 推荐 Feed 读路径 SHALL 在索引冷启动时回填 Redis 再返回

在 `ucg-feed-geo-composite-score` 已采用的 Redis 复合分 Feed 读路径之上，当推荐索引冷启动（ZSET 空且 MySQL 有 published 帖）时，Feed 读路径 MUST NOT 直接返回空列表；MUST 先执行有界索引 warm（见 `ucg-feed-index-lazy-warm`），再继续 GEO/ZSET/cursor 分页。snapshot miss 的 `backfillPostSnapshot` 语义不变。

#### Scenario: 未 backfill 环境首次打开推荐

- **WHEN** 用户请求 `GET /ucg/app/api/feed/recommend` 且无 cursor，Redis 尚无 recommend score，MySQL 有 published 帖
- **THEN** 响应 `list` MUST NOT 因索引缺失而恒为空（在 warm cap 内应有帖）

#### Scenario: 已有索引行为不变

- **WHEN** Redis ZSET 已有成员
- **THEN** Feed 排序、cursor、distance 行为 MUST 与 `ucg-feed-geo-composite-score` 一致，且无额外 warm 开销
