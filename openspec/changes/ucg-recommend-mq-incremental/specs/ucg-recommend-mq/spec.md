## ADDED Requirements

### Requirement: 推荐分更新事件 SHALL 使用已注册的 ucg.recommend routing keys

`ucg-service` 在推荐分需更新时 MUST Publish 下列 routing key 之一（经 `eventkit` 注册校验）：

- `ucg.post.published`
- `ucg.post.unpublished`
- `ucg.post.liked`
- `ucg.post.unliked`
- `ucg.comment.published`
- `ucg.comment.removed`

载荷 MUST 至少含 `postId`；comment 类事件 MUST 含 `commentId`。

#### Scenario: 审核通过后发布推荐更新

- **WHEN** 帖子 Green 审核 CAS 成功且 `status=published`
- **THEN** 系统 MUST Publish `ucg.post.published`，载荷含 `postId`

#### Scenario: 删帖或下架统一 unpublish

- **WHEN** 作者删除帖子或帖子从 published 变为 rejected/unpublished
- **THEN** 系统 MUST Publish `ucg.post.unpublished`，载荷含 `postId`

### Requirement: RabbitMQ 拓扑 SHALL 为推荐分队列绑定 topic exchange

仓库 `hack/rabbitmq-init` MUST 声明 durable 队列 `ucg.recommend.score.q` 并与 `voice.events` exchange 完成 binding（覆盖上述 routing key 或 `ucg.recommend.#`）。

#### Scenario: init 后队列可见

- **WHEN** 运维执行 rabbitmq init
- **THEN** 管理台 SHALL 可见 `ucg.recommend.score.q` 且 binding 正确

### Requirement: 推荐分 AMQP consumer MUST 驻留 ucg-service 且 SHALL 单帖更新

`ucg-service` MUST 经 AMQP push（`autoAck=false`）消费 `ucg.recommend.score.q`，按 routing key 分发：

- `published` / `liked` / `unliked` / `comment.published` / `comment.removed` → `RecomputeRecommendScore(postId)` UPSERT `ucg_post_recommend`
- `unpublished` → `DELETE FROM ucg_post_recommend WHERE post_id=postId`

MUST NOT 在 consumer 内对全部 published 帖做无 LIMIT 全表扫描。

#### Scenario: unpublished 删除推荐行且永远 Ack

- **WHEN** consumer 收到 `ucg.post.unpublished` 且 `postId` 合法
- **THEN** 系统 MUST 执行 DELETE；`RowsAffected=0` MUST 视为成功且 MUST NOT 报错；处理完成后 MUST Ack 且 MUST NOT Nack requeue

### Requirement: like 类事件 throttle MUST 仅限制重算频率且 SHALL 允许短期误差

对 `ucg.post.liked`、`ucg.post.unliked`、`ucg.comment.published`、`ucg.comment.removed`，consumer MUST 对 `postId` 使用 Redis 单 key `SET NX EX` throttle（默认 500ms）。窗口内 NX 失败 MUST 跳过本次重算并 Ack。throttle MUST 仅保证 500ms 内最多 1 次 `RecomputeRecommendScore`；MUST NOT 保证反映每一次 like/unlike 方向变化。Publisher MUST NOT 在发送侧合并事件。

#### Scenario: 500ms 内多次 like 只重算一次

- **WHEN** 同一 `postId` 在 500ms 内收到 3 条 `ucg.post.liked`
- **THEN** consumer MUST 最多执行 1 次 `RecomputeRecommendScore`，且 3 条消息 MUST 均 Ack

#### Scenario: like 后 unlike 可能被 throttle 跳过

- **WHEN** 同一 `postId` 先收到 `ucg.post.liked` 触发重算，500ms 内又收到 `ucg.post.unliked` 且 throttle NX 失败
- **THEN** consumer MAY 跳过 unlike 触发的重算且 MUST Ack；热区 reconciler MUST 在后续扫描中将 score 收敛到正确值

### Requirement: 热区 reconciler MUST 分页增量且轮首固定 hotCutoff

系统 MUST 使用 MySQL 表 `ucg_recommend_hot_scan_state`（含 `last_post_id` 与 `round_hot_cutoff`）驱动热区分页 reconciler。当 `last_post_id=0` 开始新轮时 MUST 计算并持久化 `round_hot_cutoff`；同一轮后续分页 MUST 使用已存 `round_hot_cutoff` 且 MUST NOT 在分页过程中用 `NOW()` 重新计算 cutoff。每 tick MUST 使用 `published_at >= round_hot_cutoff AND id > last_post_id ORDER BY id LIMIT pageSize`；MUST NOT 一次加载全部 published 帖。

#### Scenario: 热区一轮扫完重置 cursor

- **WHEN** 某 tick 返回行数小于 `pageSize`
- **THEN** reconciler MUST 将 `last_post_id` 置 0；下轮开始 MUST 重新计算 `round_hot_cutoff`

#### Scenario: 续扫同一轮不重算 cutoff

- **WHEN** `last_post_id > 0` 且 reconciler 执行下一页
- **THEN** 系统 MUST 继续使用表中 `round_hot_cutoff` 且 MUST NOT 更新为新的 `now - hotWindow`

### Requirement: 热区 reconciler MUST 周期性重算即使无互动

在冷区零兜底前提下，热区 reconciler MUST 对每个扫到的 published 热区帖执行 `RecomputeRecommendScore`，即使该帖在扫描周期内无任何 like/comment 事件。该行为 MUST 用于热区时间衰减与 throttle 误差的最终一致收敛。

#### Scenario: 无互动帖仍更新 score

- **WHEN** 热区 reconciler 扫描到 `postId` 且该帖在上一扫描周期内 like/comment 计数未变
- **THEN** 系统 MUST 仍执行 `RecomputeRecommendScore(postId)` 以反映 `exp(-age/τ)` 变化

### Requirement: 冷区 MUST NOT 运行分页 reconciler

`published_at < round_hot_cutoff`（冷区）的帖子 MUST 仅通过 MQ 事件更新推荐分；系统 MUST NOT 为冷区启动定时全量或分页扫表任务。

#### Scenario: 冷区帖仅靠互动更新

- **WHEN** 冷区帖收到 `ucg.post.liked` 且 throttle 允许重算
- **THEN** 系统 MUST 更新该帖 `ucg_post_recommend.score` 且 MUST NOT 依赖冷区 reconciler

### Requirement: ucg-service AMQP audit 与 recommend consumer SHALL 共用单 connection

`ucg-service` 内 UCG 审核队列 consumer 与推荐分 consumer MUST 共用 **一条** AMQP connection；每个消费队列 MUST 使用 **独立 channel** 与独立 prefetch 配置。连接断线 MUST 统一 backoff 重连并恢复全部 channel Consume。

#### Scenario: 单 connection 多 channel

- **WHEN** ucg-service 启动且 audit 与 recommend consumer 均 enabled
- **THEN** 进程 MUST 仅建立 1 条到 RabbitMQ 5672 的 AMQP connection，且 audit 4 队列与 recommend 1 队列各占用独立 channel
