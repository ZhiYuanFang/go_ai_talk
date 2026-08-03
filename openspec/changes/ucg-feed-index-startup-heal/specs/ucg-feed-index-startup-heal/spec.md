## ADDED Requirements

### Requirement: ucg-service SHALL self-check and auto-heal recommend feed index on startup

`ucg-service` 进程启动时（依赖检查通过后、HTTP 服务可接受请求的路径上）MUST 执行一次推荐 Feed 索引自检（任务名 **FeedIndexStartupHeal** 的检测阶段），比较：

- MySQL `ucg_post` 中 `status=published` 的计数 `publishedCount`
- Redis `ZCARD` of `ucg:recommend:score`（经 cachekit / `UCGRecommendScoreKey`）得 `zcard`

当 `publishedCount > 0` 且（`zcard == 0` 或 `publishedCount - zcard >=` 配置阈值 `indexHealGapThreshold`，默认 **50**）时，系统 MUST 打 **ERROR** 级别可观测日志（含 `publishedCount`、`zcard`、gap）。

当自检开关开启且自动补齐开关开启时，系统 MUST **异步**（MUST NOT 阻塞 HTTP `Run` 启动）执行有界 heal：分页读取 published 帖并调用与 `cmd/ucg-feed-backfill` / `syncPublishedPostRedis` 等价的写入（ZADD `ucg:recommend:score`、有坐标则 GEOADD `ucg:feed:geo`、写 post/profile snapshot）。heal MUST 使用独立分布式锁键（cachekit 登记，如 `ucg:feed:index:startup-heal:lock`），与请求路径 warm 锁分离；未获锁 MUST 跳过 heal 并 Warning，MUST NOT Fatal 进程。单帖失败 MUST 记日志并继续。heal 处理帖数 MUST 不超过配置上限 `indexStartupHealMaxPosts`（默认 **10000**）。

环境/配置 MUST 支持关闭自检或仅自检不 heal（例如 `UCG_FEED_INDEX_STARTUP_CHECK_ENABLED`、`UCG_FEED_INDEX_STARTUP_HEAL_ENABLED`；默认均开启）。关闭自检时 MUST NOT 启动 heal。本任务 MUST NOT 实现为 `time.Ticker` 周期扫表。

#### Scenario: 启动发现缺口并异步补齐

- **WHEN** ucg-service 启动且 publishedCount=500、zcard=10、阈值=50、check 与 heal 均开启
- **THEN** 系统 MUST 打 ERROR 日志含两计数，且 MUST 在后台尝试 heal 灌入索引（获锁成功时），HTTP 监听 MUST 仍可启动

#### Scenario: 无缺口不 heal

- **WHEN** publishedCount=100、zcard=100（或 gap 低于阈值）
- **THEN** 系统 MUST NOT 因启动自检触发 heal 灌库

#### Scenario: 仅检查不补齐

- **WHEN** check 开启且 heal 关闭，且存在缺口
- **THEN** 系统 MUST 打 ERROR 日志且 MUST NOT 自动 ZADD 补齐

#### Scenario: 多副本争锁

- **WHEN** 两副本同时启动且均检出缺口
- **THEN** 至多一个副本 MUST 持有 startup-heal 锁执行灌库；另一副本 MUST 跳过或等待锁策略后跳过，MUST NOT 无锁双全量狂写

#### Scenario: 自检关闭

- **WHEN** startup check 开关为 false
- **THEN** 启动 MUST NOT 执行缺口比较与 heal
