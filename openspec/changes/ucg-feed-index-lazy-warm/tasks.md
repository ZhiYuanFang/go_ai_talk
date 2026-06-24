## 1. 配置与 Redis 键

- [x] 1.1 在 `FeedConfig` / `LoadFeedConfig` 增加 `indexAutoWarmEnabled`、`indexWarmBatchSize`、`indexWarmMaxPosts`、`indexWarmLockSeconds` 及默认值（见 design D4）
- [x] 1.2 在 `manifest/config/config.ucg-service.yaml` 的 `ucg.feed` 下登记上述字段及中文注释
- [x] 1.3 在 `internal/platform/cachekit/keys_ucg.go` 登记 `ucg:feed:index:warm:lock` builder 及 TTL 语义注释

## 2. 索引 warm 核心逻辑（go_ai_talk）

- [x] 2.1 新增 `feed_index_warm.go`：`isFeedIndexCold`（ZCARD==0 + MySQL published COUNT>0）、`ensureFeedIndexWarm`（SET NX 锁、分页 SELECT published、`syncPublishedPostRedis`、cap/batch、defer 释放锁）
- [x] 2.2 未获锁方：短退避（~200ms）后重读 ZCARD，仍 0 则跳过 warm 继续空 Feed（不阻塞）
- [x] 2.3 在 `ListRecommendFeed` 调用 `collectFeedCandidates` 之前调用 `ensureFeedIndexWarm`（`indexAutoWarmEnabled=false` 时 no-op）
- [x] 2.4 输出结构化日志：`feed_index_warm_start` / `feed_index_warm_done`（posts_ok、posts_fail、duration_ms、zcard_after）

## 3. Runbook 与运维文档

- [x] 3.1 在 `docs/runbooks/release-deploy-and-run.md`（或 UCG Feed 相关章节）补充：lazy warm 与 `cmd/ucg-feed-backfill` 关系、验收命令（`ZCARD ucg:recommend:score`、冷启动首请求）、2G 小机关闭/降 cap 说明
- [x] 3.2 在 `manifest/docker/env/.env.test`（若适用）为 warm 相关 env 覆盖项补充 `# 含义：` 注释（若实现 env 覆盖）

## 4. Flutter 客户端（兄弟仓 flutter_ai_talk）

- [x] 4.1 为 `/feed/recommend` 读请求配置 ≥45s 超时（`UcgApiClient.get` 或 `ApiClient` 层 per-request / UCG 专用 timeout，避免 warm 中途断开）
- [x] 4.2 确认 `ucg_square_tab.dart` 首屏 `_loading` 期间不展示「暂无动态」（`_initialLoaded` 门控）；若 Masonry 空列表 loading 体验不佳，首屏改为居中 `CircularProgressIndicator`
- [x] 4.3 可选：推荐 Tab 首屏 `items.isEmpty && !hasMore` 且请求耗时 <3s 时单次延迟 2s 自动 `_load(refresh:true)` 兜底（竞态未 warm 完成）

## 5. 验收

- [x] 5.1 测试环境：清空 recommend ZSET 后首请求 Feed 返回帖子（cap 内）；日志含 warm_done；`ZCARD`>0
- [x] 5.2 `indexAutoWarmEnabled=false` 时行为与变更前一致（空索引则空 Feed）
- [x] 5.3 Flutter：冷启动首屏可等待 warm 完成，无过早超时或误显「暂无动态」
