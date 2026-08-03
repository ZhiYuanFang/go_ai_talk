## MODIFIED Requirements

### Requirement: 推荐 Feed 索引冷启动 SHALL 按需从 MySQL warm Redis

当推荐索引需要 warm（见下方冷/短缺判据）且 `ucg.feed.indexAutoWarmEnabled` 为 true 时，`ListRecommendFeed` MUST 在组装候选集之前执行有界 warm：分页读取 published 帖并调用与 `cmd/ucg-feed-backfill` 等价的 `syncPublishedPostRedis`（ZADD score、有坐标则 GEOADD、写 post/profile snapshot）。warm 完成后 MUST 继续现有 GEO+ZSET+cursor 读路径，**MUST NOT** 改用 MySQL 排序降级。

**需要 warm 的判据** MUST 为：MySQL 存在 `status=published` 帖子，且（`ZCARD ucg:recommend:score` 为 0，**或** `publishedCount - zcard >=` 配置阈值 `indexHealGapThreshold`（默认 **50**））。仅 `ZCARD > 0` 但相对 published **无明显短缺**时 MUST NOT 触发全量/有界 warm。

warm MUST 使用分布式锁 `ucg:feed:index:warm:lock`（cachekit 登记键）防止并发惊群；单请求 warm 帖数 MUST 不超过配置的 `indexWarmMaxPosts`（默认 2000）；单批分页大小 MUST 可配置（默认 200）。单帖 warm 失败 MUST 记日志并继续后续帖（best-effort）。

#### Scenario: ZSET 空且 DB 有 published 帖

- **WHEN** `ZCARD ucg:recommend:score` 为 0 且 MySQL published 计数 > 0，且 auto warm 开启
- **THEN** 系统 MUST warm 至少一批帖至 Redis 并重试 Feed 候选收集，响应 SHOULD 含帖子（在 cap 覆盖范围内）

#### Scenario: ZSET 非空但相对 published 明显短缺

- **WHEN** `ZCARD` 为 20、published 为 500、缺口不低于阈值，且 auto warm 开启
- **THEN** 系统 MUST 触发有界 warm 补齐索引（受 `indexWarmMaxPosts` 限制），MUST NOT 因 ZCARD>0 而跳过

#### Scenario: 索引充足不 warm

- **WHEN** `ZCARD` 与 published 的差低于阈值（含大致齐平）
- **THEN** 系统 MUST NOT 触发有界 warm

#### Scenario: auto warm 关闭

- **WHEN** `indexAutoWarmEnabled=false`
- **THEN** 系统 MUST 保持空/短缺 ZSET 时不在请求路径 warm，依赖启动 heal 或运维 backfill

#### Scenario: 并发 Feed 请求

- **WHEN** 多个请求同时检测到需要 warm
- **THEN** 仅一个请求 MUST 持有 warm 锁执行灌库；其他请求 MUST 等待或短退避后读 ZCARD，不得无界阻塞

#### Scenario: warm 与 publish 写路径一致

- **WHEN** warm 处理某 published 帖
- **THEN** Redis MUST 写入 `ucg:recommend:score` 与（若有坐标）`ucg:feed:geo` 及 snapshot，语义与 `publishPostCAS` 后同步一致
