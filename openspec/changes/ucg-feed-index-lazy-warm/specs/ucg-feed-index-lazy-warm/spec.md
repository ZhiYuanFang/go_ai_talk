## ADDED Requirements

### Requirement: 推荐 Feed 索引冷启动 SHALL 按需从 MySQL warm Redis

当 Redis 推荐索引冷启动（`ucg:recommend:score` ZSET 为空且 MySQL 存在 `status=published` 帖子）且 `ucg.feed.indexAutoWarmEnabled` 为 true 时，`ListRecommendFeed` MUST 在组装候选集之前执行有界 warm：分页读取 published 帖并调用与 `cmd/ucg-feed-backfill` 等价的 `syncPublishedPostRedis`（ZADD score、有坐标则 GEOADD、写 post/profile snapshot）。warm 完成后 MUST 继续现有 GEO+ZSET+cursor 读路径，**MUST NOT** 改用 MySQL 排序降级。

warm MUST 使用分布式锁 `ucg:feed:index:warm:lock`（cachekit 登记键）防止并发惊群；单请求 warm 帖数 MUST 不超过配置的 `indexWarmMaxPosts`（默认 2000）；单批分页大小 MUST 可配置（默认 200）。单帖 warm 失败 MUST 记日志并继续后续帖（best-effort）。

#### Scenario: ZSET 空且 DB 有 published 帖

- **WHEN** `ZCARD ucg:recommend:score` 为 0 且 MySQL published 计数 > 0，且 auto warm 开启
- **THEN** 系统 MUST warm 至少一批帖至 Redis 并重试 Feed 候选收集，响应 SHOULD 含帖子（在 cap 覆盖范围内）

#### Scenario: 索引非空不 warm

- **WHEN** `ZCARD ucg:recommend:score` > 0
- **THEN** 系统 MUST NOT 触发全量 warm

#### Scenario: auto warm 关闭

- **WHEN** `indexAutoWarmEnabled=false`
- **THEN** 系统 MUST 保持现有行为（空 ZSET 则空 Feed），依赖运维 backfill

#### Scenario: 并发 Feed 请求

- **WHEN** 多个请求同时检测到冷启动
- **THEN** 仅一个请求 MUST 持有 warm 锁执行灌库；其他请求 MUST 等待或短退避后读 ZCARD，不得无界阻塞

#### Scenario: warm 与 publish 写路径一致

- **WHEN** warm 处理某 published 帖
- **THEN** Redis MUST 写入 `ucg:recommend:score` 与（若有坐标）`ucg:feed:geo` 及 snapshot，语义与 `publishPostCAS` 后同步一致
