## Why

生产 Redis 换机冷启后，广场推荐 Feed 依赖的 `ucg:recommend:score` 索引未灌全；旧帖仍在 MySQL，但 lazy warm 仅在 `ZCARD==0` 时触发——冷启后若先有新帖写入导致 ZCARD>0，旧帖永久进不了广场。需要 `ucg-service` 启动时自检缺口并**自动补齐**，同时修正请求路径冷判据，避免依赖人工跑 `ucg-feed-backfill`。

## What Changes

- **新增** `ucg-service` 启动期 Feed 索引自检：比较 MySQL published 计数与 `ZCARD(ucg:recommend:score)`；缺口超阈值时 ERROR 日志并**异步自动 heal**（复用既有 published 帖 Redis 同步语义，分布式锁防多副本并发）。
- **修正** 请求路径 lazy warm 冷判据：不再仅以 `ZCARD==0` 为冷；当索引相对 MySQL published **明显短缺**时亦允许有界 warm（堵住「先写新帖跳过 warm」陷阱）。
- 环境/配置开关：自检与自动 heal 可独立关闭；heal 有界（批大小/上限），失败 best-effort 记日志，**不阻塞** HTTP 监听启动。
- **非范围**：不改 Feed 排序算法、不迁 Redis AOF、不把 MySQL 列表当作推荐读主路径、不新增 ticker 周期扫表 reconciler、不改 App API 契约。

## Capabilities

### New Capabilities

- `ucg-feed-index-startup-heal`：`ucg-service` 启动自检 + 异步自动补齐推荐索引（ZSET/GEO/snapshot）。

### Modified Capabilities

- `ucg-feed-index-lazy-warm`：冷启动判据从「仅 ZCARD=0」扩展为「空或相对 published 明显短缺」时允许有界 warm。

## Impact

- **宿主进程**：`ucg-service`（`cmd/ucg-service/main.go` 启动钩子）。
- **代码**：`internal/services/ucg/feed_index_warm.go`（冷判据）；新增 startup heal 模块；复用 `syncPublishedPostRedis` / backfill 等价路径；`cachekit` 可新增 startup heal 锁键（若与 warm 锁分离）。
- **配置**：`manifest/config/config.ucg-service.yaml` 与/或 env（如 `UCG_FEED_INDEX_STARTUP_CHECK_*` / heal 开关）。
- **背景任务**：启动一次性异步 heal（非 ticker）；须在 design 声明开关与失败语义（符合「新增后台任务须 OpenSpec 批准」）。
- **运维**：保留 `cmd/ucg-feed-backfill` 作人工全量工具；本变更不删除 CLI。
- **当前生产缺口**：上线前仍建议先手动 `--posts-only` backfill 救急；本变更为防再发。
