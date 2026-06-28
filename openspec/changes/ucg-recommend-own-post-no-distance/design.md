## Context

`ListRecommendFeed` 经 `assembleFeedPosts` 为每条有坐标帖计算 haversine 并填充 `distanceMeters`。推荐池不排除作者本人帖，故登录用户刷推荐 Tab 可能看到自己的帖并出现 `0m` 等无意义距离。`assembleFeedPosts` 亦被关注 Feed 复用，但本变更 **仅约束推荐 Feed 响应**；composite 排序（`distanceTerm`）**不在范围内**。

## Goals / Non-Goals

**Goals:**

- `GET /ucg/app/api/feed/recommend`：viewer 为帖子作者时，响应 item **omit** `distanceMeters`。
- Flutter 推荐 Tab 瀑布流与详情：本人帖 **不展示** 距离角标/文案；客户端以 `authorId == 当前 wxId` 兜底。
- 规格增量覆盖 `ucg-recommend-feed` 与 `ucg-app-http-api`（Feed 距离例外）。

**Non-Goals:**

- 修改 `computeFinalScore` / `distanceTerm`（本人帖仍可参与距离加权排序）。
- 修改 `GET /feed/following`、`GET /posts/{id}` 距离语义。
- 从推荐池排除本人帖。
- 新增 Redis 键或 DB 变更。

## Decisions

### D1：推荐 Feed 响应层 strip，不改共用组装排序

**选择**：在 `ListRecommendFeed` 返回前，对 `list` 中 `AuthorWxId == viewerWxID` 的 item 清空 `DistanceMeters`（或组装时传入 `omitOwnDistance=true` 且 **仅** `ListRecommendFeed` 传 true）。

**理由**：关注 Feed 与详情保持现行为；最小侵入；排序路径零改动。

**备选**：改 `assembleFeedPosts` 全局 skip 本人距离 → 会波及关注 Feed，超出 scope。

### D2：排序仍含 distanceTerm

**选择**：不在 `collectFeedCandidates` / `computeFinalScore` 对本人帖置零。

**理由**：负责人明确「仅响应去掉距离」；本人帖在发帖坐标附近时排序加成通常可忽略。

### D3：Flutter 客户端双保险

**选择**：`UcgPost` 增加 `shouldShowDistance(String? currentUserId)`（或等价 getter），在 `ucg_masonry_feed_card.dart` 与 `ucg_post_detail_screen.dart` 使用；条件：`distanceDisplay.isNotEmpty && authorId != currentUserId`。

**理由**：旧服务端未升级时仍可能返回本人帖距离；推荐 Tab 体验一致。

**范围**：推荐 Tab 卡片必改；详情若从推荐进入且为本人帖同样隐藏（全局按 author 判断即可，无需区分入口）。

### D4：Flutter OpenSpec

**选择**：在 `flutter_ai_talk` 兄弟仓 `openspec/changes/ucg-recommend-own-post-no-distance/specs/ucg-square-feed/spec.md` 镜像 ADDED 要求（与 go_ai_talk tasks 同名 change 便于追踪）。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 本人帖在推荐里仍可能因 distanceTerm 略靠前 | 产品已接受；后续可独立 change 调整排序 |
| 关注 Feed 仍可能对本人帖返回距离（若自关注等边缘情况） | 本 change 明确 out of scope |
| 详情页从「我的动态」进入仍可能显示距离 | 不变；仅推荐路径 + 全局 author 判断在详情也隐藏本人距离，与产品一致 |

## Migration Plan

1. 部署 ucg-service（响应 strip）。
2. 部署 App（客户端兜底）。
3. 无数据迁移；回滚仅还原两处逻辑。

## Open Questions

（无）
