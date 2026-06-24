## Why

`ucg-feed-geo-composite-score` 已为推荐 Feed 引入 Redis 帖子/作者 snapshot、`likedByMe` SET 与 `distanceMeters` 组装能力，但**关注 Feed** 仍走 `postsFromResult` 全量 MySQL JOIN + N+1 profile/like 查询，无法在列表展示距离，且与推荐 Feed 读路径不一致。产品需要在**不改变关注 Feed 分页语义**（仍 `page`/`total`、仍按 `published_at DESC`）的前提下，复用既有 snapshot 组装与距离计算，降低 DB 压力并统一客户端距离 badge 体验。

## What Changes

- **Hybrid B-lite 架构**：MySQL 保留 `ucg_follow` 关注列表与 `ucg_post` 分页（`published_at DESC`，`page`/`total`）；Redis **仅替换** `postsFromResult` 组装层，复用 `assembleFeedPosts`（post snapshot、profile snapshot、liked SET、`distanceMeters`）。
- **无新 Redis 键**：不建 author ZSET、followees SET 或关注 timeline；沿用 `ucg-feed-geo-composite-score` 已有 snapshot/liked 键。
- **无 cursor**：关注 Feed 不分页改造，不下发 `nextCursor`/`hasMore` 替代 `total`。
- **API 扩展**（同路由，非新 endpoint）：`GET /ucg/app/api/feed/following` 增加可选 query `lat`、`lng`；item 在 viewer 与帖均有坐标时含 `distanceMeters`；响应仍为 `UcgPageRes` `{ list, total, page, pageSize }`。**不计入**新 usage 统计（query 扩展 only）。
- **snapshot miss**：与推荐 Feed 相同，`backfillPostSnapshot` best-effort 回源 MySQL + profile API 后回填。
- **Flutter**（`flutter_ai_talk`）：`fetchFollowingFeed` 传 lat/lng；关注 Tab 刷新时 `tryGetCurrentCoords`，翻页复用坐标；瀑布流 card 已有 distance badge，无结构 UI 变更。

## Capabilities

### New Capabilities

- `ucg-following-feed`：关注 Feed 读路径——MySQL 分页取 postId + Redis snapshot 组装 + 可选距离；明确 Non-Goals（无 Redis timeline、无 cursor）。

### Modified Capabilities

- `ucg-app-http-api`：`GET /ucg/app/api/feed/following` 可选 `lat`/`lng` query；列表 item 含 `distanceMeters`（与推荐/详情一致语义）；分页契约保持 `{ list, total, page, pageSize }`。

## Impact

- **go_ai_talk**
  - `internal/services/ucg/feed.go` — `ListFollowingFeed` 改为 DB 取 postId 后 `assembleFeedPosts`；移除 `postsFromResult` 调用
  - `internal/controller/ucg_app_api.go` — `FeedFollowing` 解析并传递 viewer 坐标
  - `api/v1/ucg_app_http.go` — `UcgFeedFollowingReq` 增加可选 `Lat`/`Lng`
  - 依赖 `ucg-feed-geo-composite-score` 已落地的 snapshot/liked/cachekit 基础设施；**无** `keys_ucg.go` 新增键
  - **无** usage 统计变更；**无** 新路由
- **flutter_ai_talk**（`d:\work\flutter_ai_talk`）
  - `app/lib/ucg/data/ucg_repository.dart` — `fetchFollowingFeed` 可选 lat/lng
  - `app/lib/ucg/ui/ucg_square_tab.dart` — 关注模式定位与翻页坐标复用
  - 验证 `fetchPost`/详情路径在关注场景下 distance 字段可用（geo change 已覆盖则仅验收）
- **约束**：Redis 经 `cachekit` + `WithObserver`；OpenSpec 中文文档；**无**新增 `*_test.go`；**无**新增 background ticker
