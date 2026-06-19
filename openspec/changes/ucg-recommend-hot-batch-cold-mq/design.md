## Context

推荐分更新现状：

- **路径 A（MQ）**：`post.published` / `liked` / `unliked` / `comment.*` → consumer → `RecomputeRecommendScore`；`unpublished` → DELETE。
- **路径 B（reconciler）**：每 60s 分页扫 72h 热区，无互动也重算（时间衰减）。
- **Feed**：`ORDER BY r.score DESC, published_at DESC`；无 recommend 行时 `score` 为 NULL。

运维痛点：共享 MySQL 上 test+prod+sim 叠加，热区 MQ 重算（每次互动 2 DB op + Redis 节流）与 60s reconciler 构成持续写压。产品确认：热区排序可非实时；新帖需可见置顶；冷区老帖互动仍须能翻红。

## Goals / Non-Goals

**Goals:**

- 热区 reconciler 默认 1h tick，降低背景写库频率。
- 热区互动不触发 MQ 重算；冷区互动仍触发。
- 新帖不发 `post.published` recommend MQ；Feed 未算分帖置顶直至 reconciler 首次 UPSERT。
- reconciler 与 `UCG_RECOMMEND_MQ_CONSUMER_ENABLED` 解耦。
- 删帖/下架仍清除 `ucg_post_recommend` 行。

**Non-Goals:**

- 不优化 `ListRecommendFeed` / `postsFromResult` N+1 读路径（另开 change）。
- 不引入冷区分页 reconciler。
- 不改推荐公式（`wNew/tau/wLike/wComment`）。
- 不在 Admin 页暴露间隔编辑。

## Decisions

### D1：热/冷分流判据（采用）

**决策**：在 recommend MQ consumer（或统一 helper `shouldRecomputeRecommendOnEvent(ctx, postID)`）内读取 `ucg_post.published_at`，与 `hotCutoff = now - hotWindowHours` 比较（与 reconciler `round_hot_cutoff` 语义一致，使用 `LoadRecommendConfig().HotWindowHours`）。

- `published_at >= hotCutoff` → **热区**：`liked/unliked/comment.*` 消息 Ack 跳过，不 `Recompute`。
- `published_at < hotCutoff`（或 `published_at==0` 回退 `created_at`）→ **冷区**：执行现有 throttle + `RecomputeRecommendScore`。

**理由**：保留规格「冷区靠 MQ 翻红」；削减热区（含 sim 新帖）互动写库。

**备选（未采用）**：全删 MQ — 冷区无法翻红，产品拒绝。

### D2：新帖曝光 — Feed 读侧置顶（采用）

**决策**：`ListRecommendFeed` 排序改为：

1. `(r.post_id IS NULL) DESC` — 尚无 recommend 行的帖置顶；
2. `p.published_at DESC` — 置顶区内按发布时间；
3. `r.score DESC` — 已算分帖按分数；
4. `p.id DESC` — 稳定次序。

停止 `publishPostCAS` 内 `PublishPostPublished` 调用。

**理由**：零额外写库；「置顶直到 reconciler 扫到」= 直到存在 recommend 行。

### D3：reconciler 默认 1h + env 覆盖（采用）

**决策**：

- yaml `hotScanIntervalSeconds: 3600`。
- 新增 env `UCG_RECOMMEND_HOT_SCAN_INTERVAL_SECONDS`（正整数秒），在 `LoadRecommendConfig` 中优先于 yaml；缺省回退 yaml，再回退 3600。

### D4：reconciler 与 MQ consumer 解耦（采用）

**决策**：`StartUcgMQConsumers` 内 **始终** `StartRecommendHotReconciler(ctx)`（进程启动即跑，与 Rabbit 连接无关）。`UCG_RECOMMEND_MQ_CONSUMER_ENABLED` 仅控制是否订阅 `ucg.recommend.score.q`。

审核 consumer 与 recommend consumer 仍共用单 AMQP connection（规格不变）；仅 recommend 订阅可关。

### D5：unpublished 清理（采用）

**决策**：在 `DeletePost` 与 `publishPostCAS` 驳回/下架等已调用 `PublishPostUnpublished` 的路径，**同步** `RemoveRecommendScore(ctx, postID)`；可保留 MQ publish 作冗余或移除 publish 仅留同步（实现时二选一，优先同步 + 停发 MQ 减队列噪音）。

### D6：`post.published` MQ 停发

**决策**：移除 `publishPostCAS` 成功后的 `PublishPostPublished`；consumer 对 `published` routing key 若仍收到遗留消息，按热区 skip 处理（或 Ack 无 op）。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 热区帖互动后最多等一整轮 reconciler 才反映到 score | 产品已接受；热区窗口内帖量小时一轮可能 1h，帖量大时按页推进 |
| 热区边界帖（刚满 72h）短期只靠最后一次热区分数 | 划入冷区后首次互动走 MQ 更新 |
| 关 consumer 后冷区不翻红 | consumer 默认仍 enabled；仅热区 skip |
| Feed 置顶帖过多（大量未算分） | reconciler 逐批消化；置顶按 `published_at` 有序 |
| `UCG_RECOMMEND_MQ_CONSUMER_ENABLED=false` 误关 | runbook 注明冷区翻红依赖 consumer |

## Migration Plan

1. 部署新 `ucg-service` 镜像（含代码 + yaml 默认 3600）。
2. 无需 DB migration；既有 `ucg_post_recommend` 行不变。
3. 部署后观察：`Threads_running`、recommend 队列深度、Feed 首屏是否含新帖。
4. 回滚：恢复旧镜像与 60s yaml；Feed 排序回退。

## Open Questions

（无阻塞项。）
