## Context

推荐广场读路径以 Redis `ucg:recommend:score` / `ucg:feed:geo` / snapshot 为候选与展示加速；MySQL `ucg_post`（published）为正文权威。换机设计明确生产 Redis 空集群冷启。现有 `ensureFeedIndexWarm` 仅在 `ZCARD==0` 时有界 warm；冷启后若先有新帖写入，ZCARD>0 导致旧帖永不入索引。运维已有 `cmd/ucg-feed-backfill`，但易漏跑。产品选择：**启动自检 + 异步自动补齐**，并修正请求路径冷判据。

约束：跨域禁止直查他域表；Redis 经 `cachekit` + 键 builder；新增后台行为须 OpenSpec 声明宿主/开关/失败语义；不新增 ticker 扫表。

## Goals / Non-Goals

**Goals:**

- `ucg-service` 启动检测索引缺口并异步自动 heal（复用 `syncPublishedPostRedis` 语义）。
- 请求路径 lazy warm 能识别「非空但短缺」，堵住先写新帖陷阱。
- 多副本安全（锁）；开关可关；不阻塞 HTTP 启动。

**Non-Goals:**

- 不改复合分排序、cursor 协议、App API。
- 不在启动时同步阻塞全量 backfill。
- 不引入周期 reconciler（非本变更；热区 reconciler 保持原状）。
- 不自动清理 ZSET 中已软删孤儿 member（可后续另案）。

## Decisions

### 1. 缺口判据

- **选择**：`gap = publishedCount - zcard`；当 `publishedCount > 0` 且（`zcard == 0` **或** `gap >= threshold`）视为需要 heal/warm。
- **默认 threshold**：`max(50, publishedCount/10)` 或配置固定绝对值（如 `indexHealGapThreshold`，默认 **50**）——优先**可配置绝对值**，避免早期小站误判。
- **理由**：简单、可运维调参；接受软删导致 zcard≥published 时不触发（漏补风险低，因写路径删应 ZREM）。
- **备选**：逐帖抽样缺员 —— 更准但更贵，本变更不做。

### 2. 启动 heal：异步 + 独立锁

- **任务名**：`FeedIndexStartupHeal`
- **宿主**：`ucg-service`（`main` 在依赖检查与 HTTP `Run` 之前 `go`/`Start*` 触发一次）
- **触发**：进程启动一次；非 ticker
- **开关**：
  - `UCG_FEED_INDEX_STARTUP_CHECK_ENABLED`（默认 true）— 是否自检
  - `UCG_FEED_INDEX_STARTUP_HEAL_ENABLED`（默认 true）— 检出缺口后是否自动补齐；仅 check 时可只打 ERROR
- **流程**：读 published COUNT + ZCARD → 若缺口 → ERROR 日志（含两数）→ 若 heal 开 → 争用锁 `ucg:feed:index:startup-heal:lock`（新建 cachekit builder）→ 分页 published 调 `syncPublishedPostRedis`（与 backfill/warm 同语义）→ 有界 `indexStartupHealMaxPosts`（默认与 warm max 对齐或更高，如 **10000**，可配置）
- **锁 TTL**：足够覆盖有界灌库（如 10–30min 可配置）；持锁失败则跳过并打 Warning（他副本在 heal）
- **与请求 warm 锁**：分离键，避免启动 heal 与 Feed 请求 warm 长时间互斥；语义都是幂等 ZADD
- **失败**：单帖失败继续；整体失败不 `Fatal` 进程
- **理由**：自动补齐产品要求；异步不拖发布；多副本锁防打爆 Redis/MySQL

### 3. 请求路径冷判据修正

- **选择**：`isFeedIndexCold`（或重命名 `needsFeedIndexWarm`）改为：autoWarm 开启且（zcard==0 或 gap≥threshold）且 MySQL published>0。
- **有界**：仍受 `indexWarmMaxPosts` 约束；短缺场景下 warm 是**补充**而非仅「空才灌」。
- **Scenario 变更**：原「ZCARD>0 不 warm」改为「无短缺不 warm」。
- **理由**：不重启也能自愈部分缺口；与启动 heal 互补。

### 4. 不阻塞 HTTP

- **选择**：heal 在独立 goroutine；`s.Run()` 照常。
- **备选**：阻塞启动至 heal 完成 —— 发布窗口过长，否决。

### 5. Redis 键

- 沿用既有 score/geo/snapshot 键；**新增**仅 startup-heal 锁键（platform `keys_ucg.go` 登记 + 中文注释）。
- 非新读缓存键族；沿用 Feed 索引模式，design 已说明。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 多副本同时 heal | 独立分布式锁；未获锁跳过 |
| 启动 heal 打满 Redis | maxPosts / batchSize 可配；prod 先观察 |
| gap 阈值误报/漏报 | 可配 threshold；日志可观测 |
| 与 publish 并发写 | ZADD 幂等；锁分离 |
| 启动瞬间广场仍缺帖 | 异步语义接受；请求路径短缺 warm 兜底 |

## Migration Plan

1. 上线前对当前生产缺口：人工 `ucg-feed-backfill --posts-only`（可选）。
2. 部署含本变更的 `ucg-service`；确认开关默认开启。
3. 观察启动日志 `feed_index_startup_*` 与 ZCARD 收敛。
4. 回滚：关 `UCG_FEED_INDEX_STARTUP_HEAL_ENABLED` / check；或回滚镜像；索引数据无害残留。

## Open Questions

- （无阻塞）prod 默认 `indexStartupHealMaxPosts` 取 10000 还是与 `indexWarmMaxPosts` 共用——实现默认 **10000**，yaml 可覆盖。
