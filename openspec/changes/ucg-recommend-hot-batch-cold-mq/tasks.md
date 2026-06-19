## 1. 配置与启动

- [x] 1.1 `config.ucg-service.yaml`：`hotScanIntervalSeconds` 默认改为 `3600`
- [x] 1.2 `LoadRecommendConfig`：支持 env `UCG_RECOMMEND_HOT_SCAN_INTERVAL_SECONDS` 覆盖；默认回退 3600
- [x] 1.3 `ucg_mq_runner.go`：`StartRecommendHotReconciler` 与 `UCG_RECOMMEND_MQ_CONSUMER_ENABLED` 解耦，进程启动始终跑 reconciler

## 2. 热/冷分流与 MQ

- [x] 2.1 实现 `isPostInRecommendHotZone(ctx, postID)`（或等价），判据与 `hotWindowHours` 一致
- [x] 2.2 `recommend_mq_consumer`：`liked/unliked/comment.*` 热区 Ack 跳过、冷区 throttle + `Recompute`；`published` Ack 跳过
- [x] 2.3 移除 `publishPostCAS` 中 `PublishPostPublished`；`DeletePost`/下架路径同步 `RemoveRecommendScore`（评估是否停发 `PublishPostUnpublished`）

## 3. Feed 排序

- [x] 3.1 `ListRecommendFeed`：未算分帖置顶 + `published_at` / `score` / `id` 排序（见 design D2）
- [x] 3.2 确认分页 `total` 与排序语义一致（仍仅 published 帖）

## 4. 文档与验收

- [x] 4.1 `docs/runbooks/release-deploy-and-run.md`：补充热区批算 + 冷区 MQ 翻红、默认 1h、共享 MySQL 调参说明
- [x] 4.2 `manifest/docker/.env.example`：注释 `UCG_RECOMMEND_HOT_SCAN_INTERVAL_SECONDS`（可选）
- [x] 4.3 手工验收：热区点赞不触发 Recompute（日志/MQ 无写库）；冷区老帖点赞后 score 上升且 Feed 可前排；新帖过审 Feed 置顶直至 recommend 行出现；`UCG_RECOMMEND_MQ_CONSUMER_ENABLED=false` 时 reconciler 仍启动
