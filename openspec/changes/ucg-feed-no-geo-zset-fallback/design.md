## Context

`ucg-feed-geo-composite-score` design D3 规定：GEO 半径阶梯 `50→100→200→500→unlimited`，其中 **`radiusKm=0` 表示不限，改从 ZSET 全量扫**；并 **合并 ZSET 中不在 GEO 的无坐标帖**。现实现 `collectFeedCandidates` 使用条件 `hasViewer && radiusKm >= 0` 走 GEO，导致 **`radiusKm=0` 仍执行 GEOSEARCH（20000km）**，无 GEO member 时候选集为空。

现网验证（测试栈）：lazy warm `posts_ok=15`、snapshot 齐全、`GEOPOS` 为空（帖无 lat/lng），App/curl 带坐标时 Feed 恒空；与规格 Scenario「历史帖无坐标仍参与排序」不符。

约束：Redis 经 `cachekit`；无新 HTTP 路由；无 usage 统计变更；不引入 MySQL 排序降级。

## Goals / Non-Goals

**Goals:**

- 修复 `collectFeedCandidates`，使 viewer 有坐标时 **unlimited 步（radiusKm=0）从 ZSET 补候选**，无 GEO 帖 MUST 出现在 Feed。
- unlimited 步 ZSET 扫描 **不** 因 `inGeo` 跳过而漏掉仅 ZSET 帖；已在较小 GEO 半径命中的帖由 `pool`/`seen` 去重。
- 无 viewer 坐标时行为与现网一致。
- runbook 补充排查：ZCARD>0 但 Feed 空 → 查 GEOPOS / 本 fix。

**Non-Goals:**

- 修改 composite 分公式、cursor/session 语义、lazy warm。
- 强制历史帖 backfill lat/lng（产品可选，非本 change）。
- Flutter 客户端改动（后端修复即可）。
- 修复 partial ZSET 脏数据（lazy warm / backfill 另管）。

## Decisions

### D1：GEO 分支条件（采用）

**现：** `if hasViewer && radiusKm >= 0 { GEO }`

**改：** `if hasViewer && radiusKm > 0 { GEO }`

**理由：** `radiusKm=0` 落入 ZSET 分支，与 design D3「unlimited 改 ZSET 全量扫」一致。

**备选（未采用）：** 仅在有 geo 帖时走 GEO — 复杂且与阶梯语义重复。

### D2：unlimited 步 ZSET 与 `inGeo` 跳过（采用）

在 **`radiusKm == 0`** 的 ZSET 分支：

- **`hasViewer` 时不再 `continue` 跳过 `inGeo` 成员**（全量 ZSET 扫剩余帖；近距帖已在 GEO 50–500km 步进 pool，map 去重）。
- **`radiusKm > 0` 且走 ZSET 的分支**（仅 `!hasViewer` 场景）：保持现逻辑，不 skip inGeo（因无 viewer 时不查 GEO）。

实现上可在 ZSET 内循环按 `radiusKm == 0` 门控 `inGeo` skip，避免改循环结构。

**理由：** 原 `inGeo` skip 设计用于「GEO 步并行合并无 geo 帖」；unlimited 步应为 **ZSET 全量扫**，否则无 GEO 帖永远被 skip 逻辑误伤（当前 bug 根因之一）。

### D3：远距有坐标帖（采用）

GEO 阶梯最大 500km；unlimited ZSET 步 **包含** 在 GEO 索引中但 500km 外未命中的帖（ZSET 成员且 `inGeo`，此前未入 pool）。与 D2「不 skip inGeo」一致。

### D4：可观测（采用）

Debug 日志可选：`feed_collect radius=%v zset_added=%d pool=%d`（tasks 阶段按需，非必须 Info）。

### D5：验收（采用）

测试栈复现条件：ZSET 有 member、GEOPOS 空、snapshot 有、请求带 lat/lng → Feed `list` 非空。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| unlimited ZSET 与 GEO 步重复候选 | `pool` map + session `seen` 去重 |
| ZSET 全量扫略增 Redis 读 | 仅 radius=0 一步；batch 200 |
| 仍无 lat/lng 的帖无 distanceMeters | 符合原规格降级 |

## Migration Plan

1. 部署含 fix 的 `ucg-service`（无需 Redis 迁移）。
2. 已有 ZSET/snapshot 环境 **无需** backfill；直接验证 Feed。
3. 回滚：还原二进制或 revert 条件判断。

## Open Questions

- 无（实现为单行条件 + unlimited 步 skip 门控，范围明确）。
