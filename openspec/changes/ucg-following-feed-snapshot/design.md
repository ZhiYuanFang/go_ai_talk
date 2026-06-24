## Context

v2.0.5 基线中关注 Feed（`ListFollowingFeed`）流程：查 `ucg_follow` 得 followee wxId 列表 → `ucg_post` 按 `published_at DESC` 分页 → `postsFromResult` 逐行查 profile、like、媒体等。`ucg-feed-geo-composite-score` 已为推荐 Feed 实现 `assembleFeedPosts`（Redis post/profile snapshot、liked SET、`distanceMeters`、snapshot miss 回源），并明确 Non-Goal「不改关注 Feed」。

本变更在**不引入 Redis 关注 timeline** 的前提下，将关注 Feed 的**展示组装层**对齐推荐 Feed，使客户端 distance badge 在关注 Tab 可用，同时保留 MySQL 分页与排序权威来源。

约束：Redis MUST 经 `cachekit`；**禁止**新增 Redis 键；**禁止** cursor 分页；**禁止**新 App HTTP 路由（避免 usage 统计问询）；snapshot 基础设施 MUST 已由 `ucg-feed-geo-composite-score` 部署。

## Goals / Non-Goals

**Goals:**

- MySQL 仅负责：followee 集合、`ucg_post` 分页（`status=published`，`published_at DESC`）与 `total` 计数。
- Redis 负责：`assembleFeedPosts` 填充 `UcgPostItem`（snapshot、author、likedByMe、`distanceMeters`）。
- `GET /ucg/app/api/feed/following` 支持可选 viewer `lat`/`lng`；双方有坐标时返回 `distanceMeters`。
- snapshot miss → `backfillPostSnapshot`（与推荐 Feed 相同语义）。
- Flutter 关注 Tab 请求携带坐标；翻页复用首屏坐标。

**Non-Goals:**

- 全 Redis 关注 timeline（如 per-author ZSET + ZUNIONSTORE）。
- 关注 Feed 改为 cursor 分页或去掉 `total`。
- 新增 Redis 键（author ZSET、followees SET、following session 等）。
- 改变关注 Feed 排序（仍为 `published_at DESC`，**不**引入 composite score / GEO 半径筛选）。
- 新 App HTTP 路由或 usage 统计登记。
- 新增常驻 background ticker 或 `*_test.go`。

## Decisions

### D1：Hybrid B-lite 分层（采用）

| 层 | 职责 | 数据源 |
|----|------|--------|
| 身份与关系 | 当前用户 wxId | Header |
| 关注对象 | followee wxId 列表 | MySQL `ucg_follow` |
| 候选与分页 | published 帖 ID 列表 + total | MySQL `ucg_post`（`author_wx_id IN (...)`，`ORDER BY published_at DESC`，`LIMIT/OFFSET`） |
| 展示组装 | PostDTO / distance / liked | Redis snapshot + liked SET → `assembleFeedPosts` |

**理由**：关注集合与排序语义简单，MySQL 分页成熟；snapshot 组装与推荐 Feed 复用，避免重复 N+1。全 Redis timeline 成本高且本阶段无产品需求。

**替代方案（不采用）**：ZUNIONSTORE 各 author 帖 ZSET — 需新键、写路径维护、与 `published_at` 全局排序不一致。

### D2：复用 `assembleFeedPosts`（采用）

`ListFollowingFeed` 在 DB 分页得到 `pageIDs []uint64` 后：

1. 解析 viewer 坐标：`ValidViewerCoords(lat, lng)`。
2. 构造 `candidates []feedCandidate`：关注 Feed **无 GEO 预计算距离**，传空 `distKm`（`assembleFeedPosts` 内对 `distKm <= 0` 用 haversine 重算）。
3. 调用 `assembleFeedPosts(ctx, wxID, viewer, hasViewer, pageIDs, candidates)`。
4. 返回 `PageResult{List, Total, Page, PageSize}`。

**理由**：与推荐 Feed 单点组装逻辑一致；距离计算已在 `assembleFeedPosts` / `haversineKm` 实现。

### D3：DB 查询最小化（采用）

分页查询 SHOULD 仅 SELECT `id`（及组装顺序所需字段，若 ORM 需 `published_at` 保序则 ORDER BY 即可，结果只取 id 列）。**不再** JOIN profile / like / media 表。

Total 计数保留现有 `Count()` on filtered model。

### D4：API 契约（采用）

| 方法 | 路径 | 变更 |
|------|------|------|
| GET | `/ucg/app/api/feed/following` | 新增可选 query `lat`、`lng`；响应 item MAY 含 `distanceMeters`；**仍** `{ list, total, page, pageSize }` |

**非新路由** → 不修改 `usagestats/maintenance_skip.go`。

`UcgFeedFollowingReq` 增加：

```go
Lat *float64 `json:"lat" in:"query"`
Lng *float64 `json:"lng" in:"query"`
```

Controller：`FeedFollowing` 将 `req.Lat`/`req.Lng` 传入 `ListFollowingFeed(ctx, wxID, page, pageSize, lat, lng)`。

### D5：snapshot miss 回源（采用，与 geo change 一致）

`assembleFeedPosts` → `loadPostSnapshots` miss → `backfillPostSnapshot` → 写 Redis snapshot。失败时跳过该帖（best-effort），不阻塞整页。

### D6：Flutter 坐标策略（采用）

对齐推荐 Tab 模式：

- **下拉刷新 / 切换至关注 Tab**：`tryGetCurrentCoords()`，传入 `fetchFollowingFeed`。
- **上拉加载更多**：复用内存中已得坐标，**不**重复定位。
- **无定位权限**：不传 lat/lng，Feed 正常，item 无 `distanceMeters`。
- Masonry card 已有 `formatDistance(distanceMeters)` badge，无需 UI 结构变更。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| snapshot 未 backfill 导致关注 Feed 缺帖 | `backfillPostSnapshot` 回源；依赖 geo change backfill 已执行 |
| Redis 不可用时组装失败 | 与推荐 Feed 相同错误语义；运维 runbook Redis 恢复 |
| 翻页坐标与刷新坐标不一致 | 产品接受（与推荐 Feed 首屏定位语义类似）；刷新重新定位 |
| MySQL 仍承担 followee IN 查询与 COUNT | 关注人数通常有限；本 change 不优化该路径 |

## Migration Plan

1. **前置**：确认 `ucg-feed-geo-composite-score` 已部署（snapshot/liked 写路径 + backfill）。
2. 部署本变更：`ListFollowingFeed` + API query + Flutter 传参。
3. 验收：关注 Feed 列表字段与改前一致；带 lat/lng 时 `distanceMeters` 出现；likedByMe 正确；total/page 分页正常。
4. 回滚：恢复 `postsFromResult` 路径（单文件 revert `ListFollowingFeed`）；无 Redis 数据迁移。

## Open Questions

（无阻塞项。usage 统计：同路由 query 扩展，**不**新增登记。）
