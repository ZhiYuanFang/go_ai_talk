## Why

UCG 推荐分当前由 `StartRecommendWorker` 每 300s **全表**重算所有 `status=published` 帖子并写入 `ucg_post_recommend`，随帖量增长不可扩展，且与已落地的 UCG 审核 MQ 事件化方向不一致。需改为 **MQ 事件驱动单帖更新 + 热区分页增量兜底 + 冷热分离**，并 **禁止任何形式的全表刷新**。

## What Changes

- **新增** 推荐专用 RabbitMQ routing keys（`ucg.recommend.*` / `ucg.post.published|unpublished|liked|unliked` / `ucg.comment.published|removed`）与队列 `ucg.recommend.score.q`。
- **新增** `ucg-service` AMQP recommend consumer：单帖 `RecomputeRecommendScore` / `RemoveRecommendScore`；like/unlike/comment 类事件在 consumer 内 **Redis SET NX 500ms throttle**（按 `postId`）。
- **新增** 热区分页 reconciler：MySQL 小表 `ucg_recommend_hot_scan_state` 存 `last_post_id` + **`round_hot_cutoff`（轮首固定，分页不重算）**；周期性重算即使无互动（时间衰减 + 最终一致）。
- **Throttle 语义**：500ms/postId 最多算一次；**不保证** like/unlike 方向；短期误差由热区 reconciler 收敛。
- **`unpublished`**：DELETE 允许 0 行；**永远 Ack**。
- **冷区零兜底**：超过 `hotWindow` 的帖仅依赖 MQ 事件更新，不跑分页扫表。
- **统一下架事件** `ucg.post.unpublished`：作者删帖、Green/管理端从 published 驳回等场景 Publish；consumer DELETE `ucg_post_recommend` 行。
- **删除** `RefreshRecommendScores` 全表 `All()` 与定时全表 ticker；废弃 `ucg.recommend.refreshIntervalSeconds` 全表语义。
- **重构** ucg-service AMQP：**audit 与 recommend 共用 1 个 connection**，各队列 **独立 channel**（refactor `eventkit`）。
- **Publish 路径**：推荐事件 Publisher 可先 **HTTP**（与审核一致）；Consumer **AMQP push + manual ack**。
- **Non-Goals**：改推荐公式、Feed 读路径改 Redis ZSET、history/device/voice MQ 清理、新 App gateway usage 统计路由。

## Capabilities

### New Capabilities

- `ucg-recommend-mq`：推荐分 MQ 事件契约、consumer、like throttle、热区 scan state、AMQP 与 audit 共用 connection。

### Modified Capabilities

- `ucg-recommend-feed`：禁止全表刷新；MUST 事件驱动 + 热区分页增量；冷区仅 MQ。

## Impact

- **代码**：`recommend_worker.go`（重构/拆分）、`social.go`、`audit_post.go`、`audit_comment.go`、`post.go`、`post_admin.go`、recommend publisher/consumer、热区 reconciler、`eventkit/amqp_consumer.go`（shared connection）。
- **DDL**：`ucg_recommend_hot_scan_state`；`hack/rabbitmq-init` 新 binding。
- **配置**：`config.ucg-service.yaml` — `hotWindowHours`、`hotScanPageSize`、`hotScanIntervalSeconds`、`likeThrottleMs`；移除全表 refresh 语义。
- **运行时**：ucg-service 依赖 RabbitMQ 5672（recommend consumer）+ Redis（throttle 写入，非读缓存）。
