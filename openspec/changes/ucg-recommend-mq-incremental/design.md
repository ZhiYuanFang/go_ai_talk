## Context

- **现状**：`computeRecommendScore` + `RefreshRecommendScores` 每 `refreshIntervalSeconds` 对全部 published 帖 `All()` 后 UPSERT `ucg_post_recommend`；Feed 读 `ListRecommendFeed` JOIN 该表排序。
- **公式**：`score = WNew×exp(-age/τ) + WLike×log(1+likes) + WComment×log(1+comments)` — 时间项随 age 连续变化（热区明显），互动项仅 like/comment 变化时变。
- **已有**：UCG 审核 AMQP consumer（4 队列）；Publisher HTTP 15672。
- **约束**：UCG 域 only；禁止全表刷新；不新增 `*_test.go`；Redis throttle 为写入侧去重，非读缓存。

## Goals / Non-Goals

**Goals:**

- **一致性模型**：MQ consumer = 低延迟近似（允许短期 score 误差）；热区 reconciler = 热区权威收敛（时间衰减 + 读库重算）；冷区仅 MQ。
- MQ 事件驱动单帖 score 更新；删帖/驳回统一 `ucg.post.unpublished`。
- Like/unlike/comment：**每条仍 Publish**；consumer 内 **500ms/postId 单 key throttle**（只保证窗口内最多算一次，**不保证**反映每一次 like/unlike 方向变化）。
- **热区**分页 reconciler（MySQL cursor + **轮首固定 hotCutoff**）；**冷区零兜底**。
- Audit + recommend **1 AMQP connection**，每队列独立 channel + manual ack。
- 删除全表 `RefreshRecommendScores` / 原 `StartRecommendWorker` ticker。

**Non-Goals:**

- Publisher 改 AMQP；冷区分页；Redis ZSET Feed；改 ranking 公式；新 gateway-app API。
- Throttle 精确跟踪每一次互动方向（由热区 reconciler 最终一致）。

## Decisions

### 1. 事件与队列

| routing key | 载荷 | Consumer 行为 | Throttle |
|-------------|------|---------------|----------|
| `ucg.post.published` | `{ postId }` | Recompute | 否 |
| `ucg.post.unpublished` | `{ postId }` | DELETE recommend 行 | 否 |
| `ucg.post.liked` | `{ postId }` | Recompute | **是** |
| `ucg.post.unliked` | `{ postId }` | Recompute | **是** |
| `ucg.comment.published` | `{ postId, commentId }` | Recompute(postId) | **是（postId）** |
| `ucg.comment.removed` | `{ postId, commentId }` | Recompute(postId) | **是（postId）** |

- Queue：`ucg.recommend.score.q`；Binding：`ucg.recommend.#` 或逐 key。

### 2. Publish 挂点

（同前：publishPostCAS → published；published 驳回 / DeletePost / admin 驳回 → unpublished；like/unlike；comment 过审/删除。）

- pending 驳回（从未 published）：不发 unpublish。
- Publish 失败：log warning，不阻断主路径。

### 3. Throttle：只限频，不保证方向

```
key = ucg:recommend:throttle:{postId}   // 单 key，不分 inc/dec
SET key "1" NX EX 500ms
  NX 成功 → RecomputeRecommendScore(postId)
  NX 失败 → 跳过
→ Ack
```

- **语义**：500ms 内同一 `postId` **最多 1 次**重算；**不保证** like 后立即 unlike 等方向变化被反映。
- **允许短期误差**；热区 reconciler 读库重算，在「扫完一轮热区」时间尺度内 **最终一致**。
- 配置：`ucg.recommend.likeThrottleMs`（默认 500）。
- **不 throttle**：`published`、`unpublished`。

### 4. `unpublished` 与 DELETE 幂等

```sql
DELETE FROM ucg_post_recommend WHERE post_id = ?
```

| 规则 | 说明 |
|------|------|
| `RowsAffected == 0` | **正常**，MUST NOT 报错 |
| `RowsAffected >= 1` | 正常 |
| 处理完成 | **`ucg.post.unpublished` 永远 Ack**（不 Nack requeue） |

- 幂等：重复消息、从未入推荐、已删过均视为成功。
- DB 异常：log error，仍 **Ack**（避免下架消息阻塞队列；Feed JOIN `status=published` 兜底）。

### 5. 冷热分离

- **Hot**：`published_at >= round_hot_cutoff`（见 §6）。
- **Cold**：更老；`exp(-age/τ)≈0`；**仅 MQ**；无 reconciler。
- 禁止：`Where(status=published).All()` 或无 LIMIT 全量 SELECT。

### 6. 热区分页 reconciler（时间衰减 + 最终一致）

**职责（冷区零兜底下缺一不可）：**

1. **时间衰减**：无新互动也必须 periodic `Recompute`（否则热区 score 不会自然下降）。
2. **最终一致**：纠正 throttle 漏掉的 like/unlike/comment 计数。
3. **MQ 丢失补洞**：Publish 失败时热区帖仍会被扫到。

**表 `ucg_recommend_hot_scan_state`（singleton）：**

```sql
id TINYINT PK DEFAULT 1,
last_post_id BIGINT UNSIGNED DEFAULT 0,
round_hot_cutoff BIGINT NOT NULL DEFAULT 0,  -- 当前轮快照：published_at 下限（unix）
updated_at BIGINT NOT NULL
```

**一轮（round）**：从 `last_post_id=0` 开始，到某页 `rows < pageSize` 结束并 `last_post_id=0`；**可跨多个 tick**（每 tick 通常处理一页）。

```
每个 reconciler tick:

  if last_post_id == 0:
      // 新轮开始 — 仅此一次计算 cutoff，禁止在分页过程中用 NOW() 重算
      round_hot_cutoff = unix(now) - hotWindowHours
      写入 scan_state.round_hot_cutoff

  else:
      // 续扫同一轮 — 必须使用已持久化的 round_hot_cutoff
      禁止重新 now() - hotWindow

  SELECT id FROM ucg_post
  WHERE status=2
    AND published_at >= round_hot_cutoff
    AND id > last_post_id
  ORDER BY id ASC LIMIT pageSize

  foreach id → RecomputeRecommendScore(id)   // 即使 0 互动也执行

  if rows < pageSize:
      last_post_id = 0    // 下 tick 开启新轮 + 新 round_hot_cutoff
  else:
      last_post_id = max(id)
```

- **禁止**：每个 tick / 每页用 `NOW()` 重算 hotCutoff（否则边界帖不断掉出热区，游标假前进、漏扫）。
- 配置：`hotScanPageSize`（200）、`hotScanIntervalSeconds`（60）、`hotWindowHours`（默认 = `tauHours` 72）。

### 7. AMQP 共用 connection

- `SharedAMQPRunner`：单 connection；audit 4 channel + recommend 1 channel；统一 backoff 重连。

### 8. Ack 语义（recommend consumer）

| 情况 | Ack/Nack |
|------|----------|
| `unpublished`（含 DELETE 0 行） | **永远 Ack** |
| Recompute 成功 | Ack |
| throttle 跳过、JSON 非法 | Ack |
| Recompute DB 临时错误 | Nack(requeue) |

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 500ms 内 like/unlike 方向被 throttle 吞掉 | 热区 reconciler 最终一致 |
| 误差存活时间 | ≤ 热区扫完一轮 ≈ (热区帖数/pageSize)×interval |
| 冷区老帖无时间衰减更新 | 冷区对 Feed 顶部影响极小 |
| shared connection 重构波及 audit | 同 change 内一并改 |

## Migration Plan

1. DDL（含 `round_hot_cutoff`）+ rabbitmq-init + 部署。
2. 启 recommend consumer + 热区 reconciler；停全表 worker。
3. 可选：分页 seed recommend 行（非 runtime 全表）。

## Open Questions

- （已决）单 key throttle + 最终一致；unpublished 永远 Ack；轮首固定 hotCutoff 持久化；热区 reconciler 必须周期重算。
