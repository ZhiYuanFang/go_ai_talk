## Why

推荐 Feed（`ucg-feed-geo-composite-score`）候选集完全依赖 Redis `ucg:recommend:score` ZSET 与 `ucg:feed:geo` GEO。Redis 未 backfill、volume 丢失或 LRU 全冷时，MySQL 虽有 `published` 帖，接口仍返回空列表；snapshot 级 `backfillPostSnapshot` 无法补索引。产品要求「不能长期请求不到推荐列表」，需在读路径增加与 chat/history 一致的 **按需回源 + 回填 Redis** 能力。

## What Changes

- `ListRecommendFeed` 在检测到 **索引冷启动**（ZSET 为空且 MySQL 存在 published 帖）时，**同步**分页 warm：`syncPublishedPostRedis` 写入 ZSET/GEO/snapshot（复用 backfill 语义），完成后继续现有 GEO+ZSET+cursor 路径。
- 新增配置项：`ucg.feed.indexAutoWarmEnabled`、batch/cap/锁 TTL；默认开启；2G 小机可关并依赖运维 `cmd/ucg-feed-backfill`。
- 新增 Redis 锁键 `ucg:feed:index:warm:lock`（cachekit 登记），防并发惊群。
- 可观测：结构化日志（warm 触发、批次数、耗时、成功/跳过）；**不**新增 App usage 统计接口。
- runbook：部署/backfill 与 lazy warm 关系、验收命令。
- **兄弟仓 flutter_ai_talk**：推荐 Feed 首屏可能因 warm 变慢，延长 `/feed/recommend` HTTP 超时或全局 UCG 读超时；空列表文案区分「加载中/暂无」；可选首屏空结果单次自动重试（仅当后端未在同一请求内 warm 完成时的兜底）。

## Capabilities

### New Capabilities

- `ucg-feed-index-lazy-warm`：推荐 Feed Redis 索引冷启动检测、分布式锁、MySQL 分页 warm、与 `ListRecommendFeed` 集成及配置/可观测。

### Modified Capabilities

- `ucg-recommend-feed`：在 geo 复合分读路径上，索引冷启动时 MUST 回源 MySQL 重建 Redis 索引后再返回 Feed（非 MySQL 排序降级）；snapshot miss 语义不变。

## Impact

- **go_ai_talk**：`internal/services/ucg/feed.go`、`feed_snapshot.go`（或新 `feed_index_warm.go`）、`config.ucg-service.yaml`、`cachekit/keys_ucg.go`、runbook；**无**新 HTTP 路由；**无** usage 统计变更。
- **flutter_ai_talk**：`ucg_repository.dart` / `UcgApiClient` 超时、`ucg_square_tab.dart` 加载与空态（可选重试）。
- **Redis**：warm 写入与现有 publish/backfill 相同键；须评估 2G 上 cap 默认值。
- **非目标**：常驻 ticker reconciler；读路径 MySQL 排序降级；部分 LRU 驱逐的自动全量修复（二期告警即可）。
