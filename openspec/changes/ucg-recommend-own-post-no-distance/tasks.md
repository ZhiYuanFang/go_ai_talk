## 1. 后端推荐 Feed 响应（go_ai_talk）

- [x] 1.1 `ListRecommendFeed`（`feed.go`）：在 `assembleFeedPosts` 返回后，对 `viewerWxID > 0 && item.AuthorWxId == uint64(viewerWxID)` 的 item 清空 `DistanceMeters`（或等价：仅 recommend 路径传参 omit，**不**改 `ListFollowingFeed`）
- [x] 1.2 确认 `collectFeedCandidates` / `computeFinalScore` **未**改动（本人帖 `distanceTerm` 仍参与排序）
- [x] 1.3 `go build ./...` 通过

## 2. Flutter 客户端兜底（flutter_ai_talk，`d:\work\flutter_ai_talk`）

- [x] 2.1 `ucg_models.dart`：`UcgPost` 增加 `shouldShowDistance(String? currentUserId)`（或等价）：`distanceDisplay` 非空且 `authorId != currentUserId` 时为 true
- [x] 2.2 `ucg_masonry_feed_card.dart`：距离角标条件改为 `post.shouldShowDistance(currentUserId)`（从 Riverpod/auth 取当前 wxId）
- [x] 2.3 `ucg_post_detail_screen.dart`：meta 行距离展示同样使用 `shouldShowDistance`，本人帖不展示距离文案
- [x] 2.4 兄弟仓 OpenSpec：创建 `openspec/changes/ucg-recommend-own-post-no-distance/specs/ucg-square-feed/spec.md`（ADDED：Feed 卡片/详情对本人帖 MUST NOT 展示距离）
- [x] 2.5 `dart analyze` 变更文件通过

## 3. 手工验收

- [ ] 3.1 测试栈：用户 A 发带坐标帖；A 刷推荐 Feed（带 lat/lng）→ 本人帖 JSON **无** `distanceMeters`，他人帖 **有**
- [ ] 3.2 App 推荐 Tab：本人帖卡片 **无** 距离角标；他人帖 **有**
- [ ] 3.3 从推荐进入本人帖详情：meta **无** 距离；他人帖详情 **有**（带坐标时）
- [ ] 3.4 关注 Feed 行为与变更前一致（out of scope，回归 smoke）
- [ ] 3.5 `GET /posts/{id}` 本人帖详情 API 仍可按现规则返回距离（若产品仍返回，客户端详情已按 author 隐藏）
