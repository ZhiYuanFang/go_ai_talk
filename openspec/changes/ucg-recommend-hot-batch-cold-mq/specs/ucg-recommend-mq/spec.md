## MODIFIED Requirements

### Requirement: 推荐分 AMQP consumer MUST 驻留 ucg-service 且 SHALL 单帖更新

`ucg-service` MUST 经 AMQP push（`autoAck=false`）消费 `ucg.recommend.score.q`，按 routing key 分发：

- `liked` / `unliked` / `comment.published` / `comment.removed` → 若帖处于**冷区**（`published_at` 早于 `now - hotWindowHours`，与热区 reconciler 窗口一致），MUST `RecomputeRecommendScore(postId)` UPSERT `ucg_post_recommend`；若帖处于**热区**，MUST Ack 跳过本次重算（不得 Nack 重试）。
- `published` → MUST Ack 跳过且不 UPSERT（新帖 score 由热区 reconciler 写入；曝光由 Feed 未算分置顶保证）。
- `unpublished` → `DELETE FROM ucg_post_recommend WHERE post_id=postId`（或写路径已同步删除时 Ack 成功即可）。

MUST NOT 在 consumer 内对全部 published 帖做无 LIMIT 全表扫描。

#### Scenario: Cold zone like triggers recompute

- **WHEN** consumer 收到 `ucg.post.liked` 且 `postId` 对应帖 `published_at` 早于热区 cutoff
- **THEN** 系统 MUST 执行 `RecomputeRecommendScore` 且 MUST Ack

#### Scenario: Hot zone like skips recompute

- **WHEN** consumer 收到 `ucg.post.liked` 且帖 `published_at` 在热区窗口内
- **THEN** 系统 MUST NOT 调用 `RecomputeRecommendScore` 且 MUST Ack

#### Scenario: unpublished 删除推荐行且永远 Ack

- **WHEN** consumer 收到 `ucg.post.unpublished` 且 `postId` 合法
- **THEN** 系统 MUST 执行 DELETE；`RowsAffected=0` MUST 视为成功且 MUST NOT 报错；处理完成后 MUST Ack

### Requirement: like 类事件 throttle MUST 仅限制重算频率且 SHALL 允许短期误差

对 **冷区** `ucg.post.liked`、`ucg.post.unliked`、`ucg.comment.published`、`ucg.comment.removed`，consumer MUST 对 `postId` 使用 Redis 单 key `SET NX EX` throttle（默认 500ms）。窗口内 NX 失败 MUST 跳过本次重算并 Ack。throttle MUST 仅保证 500ms 内最多 1 次 `RecomputeRecommendScore`；MUST NOT 保证反映每一次 like/unlike 方向变化。Publisher MUST NOT 在发送侧合并事件。

热区事件因跳过 Recompute，MUST NOT 依赖 throttle 收敛（由 reconciler 负责）。

#### Scenario: Cold zone 500ms 内多次 like 只重算一次

- **WHEN** 冷区帖 `postId` 在 500ms 内收到 3 条 `ucg.post.liked`
- **THEN** consumer MUST 最多执行 1 次 `RecomputeRecommendScore`，且 3 条消息 MUST 均 Ack

### Requirement: 热区 reconciler MUST 分页增量且轮首固定 hotCutoff

系统 MUST 使用 MySQL 表 `ucg_recommend_hot_scan_state`（含 `last_post_id` 与 `round_hot_cutoff`）驱动热区分页 reconciler。当 `last_post_id=0` 开始新轮时 MUST 计算并持久化 `round_hot_cutoff`；同一轮后续分页 MUST 使用已存 `round_hot_cutoff` 且 MUST NOT 在分页过程中用 `NOW()` 重新计算 cutoff。每 tick MUST 使用 `published_at >= round_hot_cutoff AND id > last_post_id ORDER BY id LIMIT pageSize`；MUST NOT 一次加载全部 published 帖。

默认 tick 间隔 MUST 为 **3600s**（1 小时），可由 env `UCG_RECOMMEND_HOT_SCAN_INTERVAL_SECONDS` 覆盖。

#### Scenario: 热区一轮扫完重置 cursor

- **WHEN** 某 tick 返回行数小于 `pageSize`
- **THEN** reconciler MUST 将 `last_post_id` 置 0；下轮开始 MUST 重新计算 `round_hot_cutoff`

#### Scenario: Default interval one hour

- **WHEN** 未设置 `UCG_RECOMMEND_HOT_SCAN_INTERVAL_SECONDS` 且 yaml 为变更后默认值
- **THEN** reconciler 日志 MUST 显示 interval 为 1h 量级

### Requirement: 热区 reconciler MUST 周期性重算即使无互动

在冷区零兜底前提下，热区 reconciler MUST 对每个扫到的 published 热区帖执行 `RecomputeRecommendScore`，即使该帖在扫描周期内无任何 like/comment 事件。该行为 MUST 用于热区时间衰减与热区互动改由 reconciler 收敛后的 score 更新。

#### Scenario: 无互动热区帖仍更新 score

- **WHEN** 热区 reconciler 扫描到 `postId` 且该帖在上一扫描周期内 like/comment 计数未变
- **THEN** 系统 MUST 仍执行 `RecomputeRecommendScore(postId)` 以反映 `exp(-age/τ)` 变化

### Requirement: 冷区 MUST NOT 运行分页 reconciler

`published_at < round_hot_cutoff`（冷区）的帖子 MUST 通过 MQ 互动事件更新推荐分（见上文冷区分流）；系统 MUST NOT 为冷区启动定时全量或分页扫表任务。

#### Scenario: 冷区帖靠互动翻红

- **WHEN** 冷区帖收到 `ucg.post.liked` 且 throttle 允许重算
- **THEN** 系统 MUST 更新该帖 `ucg_post_recommend.score` 且 Feed MUST 可将该帖按 score 前排展示

### Requirement: ucg-service 热区 reconciler SHALL 独立于 recommend MQ consumer 开关

`StartRecommendHotReconciler` MUST 在 `ucg-service` 启动时运行（与 `UCG_RECOMMEND_MQ_CONSUMER_ENABLED` 无关）。`UCG_RECOMMEND_MQ_CONSUMER_ENABLED=false` MUST 仅停止订阅 `ucg.recommend.score.q`，MUST NOT 停止热区 reconciler。

#### Scenario: Consumer disabled reconciler still runs

- **WHEN** `UCG_RECOMMEND_MQ_CONSUMER_ENABLED=false` 且进程启动
- **THEN** 日志 MUST 含 `[ucg-recommend-hot] started` 且 MUST NOT 订阅 recommend 队列

### Requirement: 发帖过审 MUST NOT 发布 recommend post.published 事件

帖 `pending_audit → published` 成功后，系统 MUST NOT 调用 `PublishPostPublished` / 向 `ucg.recommend.score.q` 发送 `published` 路由事件。该帖首次 score MUST 由热区 reconciler 写入。

#### Scenario: Publish CAS does not emit recommend published

- **WHEN** `publishPostCAS` 成功将帖设为 published
- **THEN** MUST NOT 向 recommend MQ 发送 `post.published` 事件
