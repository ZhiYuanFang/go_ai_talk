## Context

`ucg-feed-geo-composite-score` 已将推荐 Feed 候选集迁移至 Redis ZSET+GEO；snapshot miss 已有 `backfillPostSnapshot`，但 **索引冷启动**（ZSET 空、MySQL 有 published 帖）时 `ListRecommendFeed` 直接返回空列表。运维 `cmd/ucg-feed-backfill` 可灌库，但依赖人工；Redis volume 丢失或新环境未 backfill 会导致用户看到「暂无动态」。私信已有 `tryWarmChatFromMySQL` 模式可类比。

约束：Redis 经 `cachekit`；新键登记 `keys_ucg.go`；**禁止**新增常驻 ticker（AGENTS.md）；Feed 读路径 Redis 已由负责人确认；本变更为 **请求内同步 warm**，非 background loop。

## Goals / Non-Goals

**Goals:**

- 索引冷启动时，单次 Feed 请求内完成 **有界** warm 并重试现有 Redis 读路径。
- 与 `syncPublishedPostRedis` / backfill **同语义**（ZSET+GEO+snapshot+profile）。
- 分布式锁防惊群；配置开关与 batch/cap；结构化日志。
- runbook 与 geo change backfill 衔接。
- Flutter：首屏 warm 可能 5–30s，调整 HTTP 超时与空态体验。

**Non-Goals:**

- 读路径 MySQL 排序降级（双套算法）。
- 部分 LRU 驱逐的自动全量修复（仅日志/运维；二期可加 ZCARD vs DB count 告警）。
- liked SET 全量 warm（首版可省略；`likedByMe` 冷启动为 false，与现网 backfill 前一致）。
- 新 HTTP 接口；usage 统计变更。

## Decisions

### D1：冷启动判定（采用）

**触发 warm 当且仅当：**

1. `ucg.feed.indexAutoWarmEnabled == true`（默认 true）；
2. `ZCARD ucg:recommend:score == 0`（Cluster 上对 key 所在 slot 执行，与现网单 key 一致）；
3. MySQL `COUNT(*) FROM ucg_post WHERE status=published > 0`。

**不触发：** 仅「本页 GEO 无候选」（用户偏远、真无帖）；ZCARD>0 但 partial eviction。

### D2：Warm 实现（采用）

新增 `ensureFeedIndexWarm(ctx)`，在 `ListRecommendFeed` 调用 `collectFeedCandidates` **之前**：

```
try SET NX ucg:feed:index:warm:lock TTL=60s
  ├─ 未拿到锁 → 短暂 sleep(200ms) 后重读 ZCARD；仍 0 则继续空 Feed（避免无限等）
  └─ 拿到锁 →
        loop: SELECT id FROM ucg_post WHERE status=published AND id>? ORDER BY id LIMIT batch
              每 id: syncPublishedPostRedis(ctx, id)  // 与 backfill 相同
              until 无行 OR 已处理 cap（默认 2000）
        DEL lock（defer）
```

- `batch` 默认 200（与 backfill page-size 一致）；`cap` 默认 2000（config `ucg.feed.indexWarmMaxPosts`）。
- 单帖失败记 warning，**不**中断整批（与 backfill 一致）。
- warm 完成后 **不** return，继续原 `ListRecommendFeed` 逻辑。

**备选（未采用）：** 异步 warm + 首请求仍空 — 用户体验差；MySQL 分页排序 Feed — 与 cursor/GEO 语义冲突。

### D3：Redis 锁键（采用）

`ucg:feed:index:warm:lock` STRING，`SET NX EX 60`。登记于 `keys_ucg.go`。非 feed session 键，TTL 短。

### D4：配置（采用）

`config.ucg-service.yaml` → `ucg.feed` 下：

| 字段 | 默认 | 说明 |
|------|------|------|
| `indexAutoWarmEnabled` | true | false 时仅依赖 backfill |
| `indexWarmBatchSize` | 200 | MySQL 分页 |
| `indexWarmMaxPosts` | 2000 | 单次请求 warm 上限 |
| `indexWarmLockSeconds` | 60 | 锁 TTL |

环境变量覆盖可选（与现有 feed 配置风格一致，tasks 阶段实现）。

### D5：可观测（采用）

Info：`feed_index_warm_start` / `feed_index_warm_done`（posts_ok, posts_fail, duration_ms, zcard_after）。  
Debug：沿用 cachekit 已有日志。

### D6：Flutter（采用）

| 项 | 决策 |
|----|------|
| HTTP 超时 | `/feed/recommend` 读超时 ≥ **45s**（或 UCG 公共 GET 超时），避免 warm 中途客户端断开 |
| 空态 | 首屏 `_loading` 期间不展示「暂无动态」；保持现有 `CircularProgressIndicator` |
| 可选重试 | 若首屏 `items.isEmpty && hasMore==false` 且耗时 <3s（疑似竞态未 warm），**单次**延迟 2s 自动 `_load(refresh:true)`；warm 同步完成时通常不需要 |

兄弟仓路径：`flutter_ai_talk/app/lib/ucg/`。

### D7：App API usage 统计

**无新路由**；不修改 `maintenance_skip.go`。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 首请求 latency 高（2000 帖 warm） | cap 默认 2000；2G 可降 cap 或关开关；日志可观测 |
| MySQL 压力 | 分页 + cap；锁防并发多请求全量扫 |
| 锁等待方仍空 | 200ms 后读 ZCARD；可下拉刷新 |
| warm 与 publish 并发 | syncPublishedPostRedis 幂等 ZADD/GEO |
| Flutter 超时 | 延长至 45s |

## Migration Plan

1. 部署 ucg-service（含 lazy warm，默认 enabled）。
2. 可选：仍建议首次上线跑 `ucg-feed-backfill` 避免首用户等待。
3. 发布 Flutter 超时调整。
4. 验收：`FLUSHDB` 后（测试环境）Feed 首请求有帖；`ZCARD`>0；日志含 warm_done。
5. 回滚：配置 `indexAutoWarmEnabled=false` 或回退二进制；依赖 backfill。

## Open Questions

- 生产 cap 2000 是否在 2G+本机 MySQL 上需降为 500？（tasks 验收后调参，可写 env 覆盖。）
