## Why

UCG 推荐分当前采用「热区 reconciler 每 60s 扫表 + MQ 即时重算（发帖/点赞/评论）」双路径。在 2C2G 同机双栈、共享 MySQL（`max_connections` 约 100）及 sim 放大量互动时，热区 MQ 重算与高频 reconciler 叠加，导致 UCG 背景写库与连接争抢，间接拖慢生产 recommend/push 等接口。产品侧接受热区排序非实时（≤1 轮 reconciler 延迟），但冷区老帖仍需凭新互动翻红，不能封死 72h 外帖子。

## What Changes

- 将热区 reconciler 默认 tick 从 **60s 调整为 3600s（1h）**（`ucg.recommend.hotScanIntervalSeconds`）。
- **热区**（`published_at` 在 `hotWindowHours` 内）：点赞/评论/取消赞/发帖过审 **不再** 触发 MQ `RecomputeRecommendScore`；由 reconciler 批量收敛（含时间衰减）。
- **冷区**（`published_at` 早于热区 cutoff）：保留 MQ 互动事件 → `RecomputeRecommendScore`，保障老帖翻红。
- 新发帖：停止 `post.published` recommend MQ；Feed 对 **尚无 `ucg_post_recommend` 行** 的帖置顶，直至 reconciler 首次算分。
- `ListRecommendFeed` 排序：未算分帖优先，再按 `score` / `published_at`。
- 下架/删帖：保留 recommend 行清理（`RemoveRecommendScore`），优先改为写路径同步调用，减少对 `post.unpublished` MQ 的依赖。
- **解耦** `StartRecommendHotReconciler` 与 `UCG_RECOMMEND_MQ_CONSUMER_ENABLED`：关闭 MQ consumer 时 reconciler 仍运行。
- 可选 env：`UCG_RECOMMEND_HOT_SCAN_INTERVAL_SECONDS` 覆盖 yaml 间隔（运维调参，非 Admin 页）。

## Capabilities

### New Capabilities

（无独立新 capability。）

### Modified Capabilities

- `ucg-recommend-mq`：热区互动不再要求 MQ 即时重算；冷区互动仍 MUST 经 MQ 更新；`post.published` 不再发 recommend 事件；reconciler 与 consumer 开关解耦。
- `ucg-recommend-feed`：Feed 须对未算分 published 帖置顶展示，直至 reconciler 写入 score。

## Impact

- `manifest/config/config.ucg-service.yaml` — `hotScanIntervalSeconds` 默认值
- `internal/services/ucg/recommend_hot_reconciler.go`、`recommend_worker.go`、`recommend_mq_consumer.go`、`recommend_publisher.go`
- `internal/services/ucg/feed.go` — `ListRecommendFeed` 排序
- `internal/services/ucg/audit_post.go`、`post.go` — 移除/保留 publish 调用
- `internal/services/ucg/ucg_mq_runner.go` — reconciler 启动条件
- `docs/runbooks/release-deploy-and-run.md` — 共享 MySQL 推荐调参说明
- 无 App 接口契约破坏性变更；Feed 排序语义对客户端透明
- RabbitMQ `ucg.recommend.score.q` 仍存在，冷区/unpublished 消息减少
