## 1. Backend API（query 扩展）

- [x] 1.1 `api/v1/ucg_app_http.go`：`UcgFeedFollowingReq` 增加可选 query `Lat`/`Lng`（`*float64`，`in:"query"`）
- [x] 1.2 `internal/controller/ucg_app_api.go`：`FeedFollowing` 将 `req.Lat`/`req.Lng` 传入 service 层
- [x] 1.3 确认 **不** 修改 `usagestats/maintenance_skip.go`（同路由 query 扩展，非新 endpoint）

## 2. Backend service（Hybrid B-lite 读路径）

- [x] 2.1 `ListFollowingFeed` 签名扩展：增加可选 `viewerLat, viewerLng *float64`；入口校验 wxId（保持现有未授权语义）
- [x] 2.2 MySQL 阶段：保留 `ucg_follow` → followee ids；`ucg_post` 过滤 published + author IN followees；`Count()` 得 total；分页 `OrderDesc(published_at)` 仅取 postId 列表（最小字段）
- [x] 2.3 组装阶段：解析 `ValidViewerCoords`；构造空 `distKm` 的 `feedCandidate` 列表；调用 `assembleFeedPosts` 替换 `postsFromResult`
- [x] 2.4 返回 `PageResult{List, Total, Page, PageSize}`；更新 `ListFollowingFeed` 注释（移除「本 change 不改造」表述）
- [x] 2.5 关键路径补充中文注释：MySQL 分页 vs Redis 组装职责、snapshot miss 行为

## 3. 依赖与验收（go_ai_talk）

- [x] 3.1 确认前置变更 `ucg-feed-geo-composite-score` 已部署（snapshot/liked 写路径 + backfill）
- [x] 3.2 手工验收：`GET /feed/following?page=1` 返回 `{list,total,page,pageSize}`；带 `lat/lng` 时 item 含 `distanceMeters`；`likedByMe` 与改前一致；无 lat/lng 时无 distance 字段
- [x] 3.3 手工验收：snapshot 缺失帖经 backfill 仍可出现在列表（抽样）
- [x] 3.4 grep 确认本 change **无** 新增 Redis 键 builder、**无** `g.Redis()` bypass、**无** 新增 `*_test.go`

## 4. flutter_ai_talk（`d:\work\flutter_ai_talk`）

- [x] 4.1 `app/lib/ucg/data/ucg_repository.dart`：`fetchFollowingFeed` 增加可选参数 `double? lat, double? lng`；query 传入 `/feed/following`（保留 `page`/`pageSize`）
- [x] 4.2 `app/lib/ucg/ui/ucg_square_tab.dart`：关注模式 `_load(refresh: true)` 时 `tryGetCurrentCoords()` 并传入 `fetchFollowingFeed`；上拉加载更多复用内存坐标（与推荐 Tab 模式对齐，不重复定位）
- [x] 4.3 确认 `UcgPost`/`parsePagedPosts` 已解析 `distanceMeters`（geo change 应已具备）；关注路径返回的 item 能驱动 masonry distance badge
- [x] 4.4 验证 `fetchPost` / `UcgPostDetailScreen`：从关注 Feed 进入详情时，若 seedPost 含 `distanceMeters` 或详情 query 带 coords，距离展示正常（无结构 UI 变更）
- [x] 4.5 手工验收：关注 Tab 下拉刷新与翻页均正常；有定位权限时 card 左下角 distance badge 显示；无权限时列表仍可用

## 5. 文档与评审

- [x] 5.1 归档前对照 delta specs（`ucg-following-feed`、`ucg-app-http-api`）做 Requirement/Scenario 验收
- [x] 5.2 确认 Non-Goals：无 cursor、无新 Redis 键、无 Redis timeline、无新 background ticker
