## 1. Backend platform（cachekit GEO/ZSET 批量）

- [ ] 1.1 `keys_ucg.go`：登记 `UCGFeedGeoKey`、`UCGRecommendScoreKey`、`UCGPostSnapshotKey`、`UCGProfileSnapshotKey`、`UCGUserLikedPostsKey`、`UCGFeedSessionKey`（含中文注释：TTL、失效语义）
- [ ] 1.2 `cachekit`：封装 GEOSEARCH/GEORADIUS WITHDIST、ZSCORE/ZADD/ZREM 批量、SET/SADD/SISMEMBER pipeline（均经 `WithObserver`）
- [ ] 1.3 `config.ucg-service.yaml`：新增 `ucg.feed.wDist`（0.5）、`distDecayKm`（10）、`radiusStepsKm`（50,100,200,500,0）、`candidateBatchSize`（200）、`sessionTtlMinutes`（30）、`snapshotTtlDays`（7）
- [ ] 1.4 `LoadFeedConfig`（或扩展 `LoadRecommendConfig`）读取上述配置

## 2. Backend ucg service（snapshots、scores、feed、social likes SET）

- [ ] 2.1 实现 post/profile snapshot 读写（JSON 序列化、CDN URL 填充、server-only lat/lng）
- [ ] 2.2 实现 `RecomputeRecommendScore` 输出 ZADD Redis ZSET（移除 MySQL `ucg_post_recommend` UPSERT）
- [ ] 2.3 `publishPostCAS`：ZADD + GEOADD（有坐标）+ snapshot 刷新；`DeletePost`/unpublished：ZREM score/geo + DEL snapshot
- [ ] 2.4 `UpdatePost`/`CreatePost`：持久化可选 lat/lng；publish 路径同步 GEO
- [ ] 2.5 like/unlike：`SADD`/`SREM` `ucg:user:{wxId}:liked-posts`；保留现有 like_count MySQL 更新
- [ ] 2.6 实现 composite score 计算 helper（`finalScore = baseScore + wDist*exp(-distKm/distDecayKm)`）
- [ ] 2.7 实现 Feed session（生成 sessionId、SET seen postIds TTL、cursor 编解码冻结 lat/lng/radius/offset）
- [ ] 2.8 重写 `ListRecommendFeed`：GEO 半径阶梯 + ZSET 无 geo 帖合并 + session 去重 + cursor 分页 + snapshot 组装 + batch likedByMe；snapshot miss 回源 MySQL
- [ ] 2.9 `GetPost`：可选 viewer lat/lng 返回 `distanceMeters` 字符串
- [ ] 2.10 MQ consumer / hot reconciler：`unpublished` 清理 Redis；重算写 ZSET only
- [ ] 2.11 DTO：`UcgPostItem` 增加 `DistanceMeters string \`json:"distanceMeters,omitempty"\``；Feed 响应改为 `{list, hasMore, nextCursor}`

## 3. API v1/v2 变更

- [ ] 3.1 `api/v2/ucg_app_http.go`（或等价）：`POST /ucg/app/api/v2/posts` 含可选 lat/lng
- [ ] 3.2 `UcgFeedRecommendReq`：增加 query `lat`、`lng`、`cursor`；废弃/忽略 `page` 当有 cursor
- [ ] 3.3 `UcgPostUpdateReq`：body 可选 lat/lng；`UcgPostGetReq`：query 可选 lat/lng
- [ ] 3.4 `internal/controller/ucg_app_api.go`：wire 新 handler；Feed 返回 cursor 结构
- [ ] 3.5 `api/v1` apiregistry：登记 v2 posts summary；**确认** `POST /ucg/app/api/v2/posts` **计入** usage（**不**写入 `maintenance_skip.go`）
- [ ] 3.6 手工验收：v2 create 2xx 出现在 usage 统计；v1 PUT/GET feed/get 仍按原策略统计

## 4. Backfill / migration runbook

- [ ] 4.1 编写 backfill 脚本/命令（`cmd/` 或 `hack/`）：分页扫描 published 帖 → ZADD + GEOADD + post/profile snapshot
- [ ] 4.2 backfill：`ucg_post_like` → 重建 `ucg:user:{wxId}:liked-posts` SET
- [ ] 4.3 `docs/runbooks/release-deploy-and-run.md`：补充 backfill 执行顺序、Redis 键验收、停写 `ucg_post_recommend` 说明
- [ ] 4.4 部署后验收：抽样 Feed 与旧 MySQL 排序大致一致（无坐标场景）；有坐标时距离字段正确

## 5. flutter_ai_talk（`d:\work\flutter_ai_talk`）

- [ ] 5.1 `pubspec.yaml`：添加 `geolocator` 依赖；iOS/Android 定位权限配置
- [ ] 5.2 `ucg_models.dart`（或等价）：`UcgPost` 增加 `String? distanceMeters`；新增 `formatDistance(meters)` helper（m/km 展示）
- [ ] 5.3 `ucg_repository.dart`：`createPost` → `POST /ucg/app/api/v2/posts` 带可选 lat/lng；`updatePost` 带 lat/lng；`fetchRecommendedFeed` 改为 cursor + lat/lng（`hasMore`/`nextCursor`，移除 `total` 依赖）
- [ ] 5.4 `fetchPost`：query 带 lat/lng；解析 `distanceMeters`
- [ ] 5.5 `ucg_square_tab.dart`：cursor 分页加载（替换 page 递增）；下拉刷新不传 cursor；获取当前位置传入 Feed
- [ ] 5.6 `ucg_compose_screen.dart`：发帖/编辑时获取 coords 传入 create/update
- [ ] 5.7 `widgets/ucg_masonry_feed_card.dart`：图片左下角 distance badge（`formatDistance`）
- [ ] 5.8 `ucg_post_detail_screen.dart`：ipLocation 右侧展示距离（有 `distanceMeters` 时）
- [ ] 5.9 手工验收：Feed 翻页无重复；距离 badge 显示；无定位权限时 Feed 仍可用（无距离字段）

## 6. 文档与评审检查

- [ ] 6.1 proposal 已标注 Redis 读缓存负责人确认；实现经 cachekit 无 bypass
- [ ] 6.2 确认无新增 `*_test.go`；无新增常驻 background ticker
- [ ] 6.3 归档前对照 v2.0.5 + 本 change delta specs 做 Requirement/Scenario 验收
