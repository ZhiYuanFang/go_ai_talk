## Why

UCG 推荐 Feed 会展示 viewer 自己的已发布帖，并在带 lat/lng 请求时返回 `distanceMeters`（常见为 `0m` 或与发帖坐标相关的距离）。「距你 X」对作者本人无产品意义，且 `0m` 角标易误解。负责人确认：**仅改 API 响应 omit 距离**，**不调整** composite 排序中的 `distanceTerm`。

## What Changes

- **推荐 Feed 响应**：当 `viewerWxID > 0` 且帖子 `authorWxId == viewerWxID` 时，item **MUST NOT** 含 `distanceMeters`（即使 viewer 与帖均有坐标）；**不**跳过 haversine 用于排序（`finalScore` 语义不变）。
- **关注 Feed / 帖子详情**：本变更 **不修改**（仍可按现有规则返回距离）。
- **Flutter（兄弟仓 `flutter_ai_talk`）**：推荐 Tab 瀑布流与从推荐进入的详情，**MUST NOT** 对本人帖子展示距离角标/文案；客户端以 `authorId == 当前登录 wxId` 兜底（兼容旧服务端仍返回距离的情况）。
- **非 BREAKING**：字段省略符合既有 `omitempty` 契约；旧客户端无距离字段时本就不展示。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `ucg-recommend-feed`：新增「作者本人帖 omit `distanceMeters`」响应场景；明确排序逻辑不变。
- `ucg-square-feed`（Flutter OpenSpec delta，兄弟仓）：推荐 Feed 卡片与详情 meta 对本人帖不展示距离。

## Impact

- **go_ai_talk**：`internal/services/ucg/feed.go`（`ListRecommendFeed` 组装后或等价路径 strip 本人帖 `distanceMeters`）；`openspec` 增量规格。
- **flutter_ai_talk**：`ucg_masonry_feed_card.dart`、`ucg_post_detail_screen.dart`（或 `UcgPost` 展示 helper）；可选 `openspec/changes/<同名或镜像 change>/specs/ucg-square-feed/spec.md`。
- **API**：`GET /ucg/app/api/feed/recommend` 响应 item 语义；`GET /feed/following`、`GET /posts/{id}` 不变。
- **Redis / 排序**：无变更。
