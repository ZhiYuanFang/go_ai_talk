## Context

- `sim-gentle-polling` 已为 T2 引入 ucg internal `POST /ucg/internal/api/posts/sample`，实现为 `ORDER BY published_at DESC LIMIT n`（默认 20），sim 侧 `rand.Intn` 选一条。
- 运营期望：**全库 published 帖均可被选为评论对象**，新帖概率略高于老帖（「一点新帖权重」），而非仅最新 20 帖。
- T2 默认周期 6h（可 env 覆盖），查询频率极低；共享 MySQL 实例下仍须避免 `ORDER BY RAND()` 全表排序。
- 探索结论采用 **路径 1：幂次偏置 ID 探测**（非双区间热区方案）。

## Goals / Non-Goals

**Goals:**

- sample API 支持 `mode=random`，返回 1 条经 ID 探测抽样的 published 帖。
- 抽样覆盖全库 `status=published` 集合；锚点 R 经 `U^α`（默认 α=0.65）偏置至高 id 端，实现轻度新帖权重。
- 查询有界：MIN/MAX + 单次 `id >= R LIMIT 1`（共 2 次 SQL）；复用现有 cover 子查询字段。
- sim T2 改调 random 模式，移除客户端选帖随机逻辑。
- 保留 `mode` 缺省时的 latest 批量行为（`published_at DESC LIMIT n`），避免破坏潜在调用方。

**Non-Goals:**

- 不引入 `ORDER BY RAND()`、COUNT+OFFSET、Redis 缓存。
- 首期不暴露 α 的 env/管理端配置（包内常量即可；后续可扩展）。
- 不做 `excludeSimAuthors`、不优化 recommend Feed。
- 不新增 `*_test.go`。
- 不实现路径 2（热区双区间 `published_at` 混合）。

## Decisions

### 1. 请求契约：body 增加 `mode`

**选择**：`POST /ucg/internal/api/posts/sample` body 增加可选字段 `mode`：

| mode | 行为 |
|------|------|
| 缺省 / `latest` | 现有逻辑：`limit` 默认 20、上限 50，`published_at DESC` |
| `random` | 忽略 `limit`（或强制视为 1），返回 0 或 1 条 ID 探测结果 |

**备选**：仅用 `limit=1` 触发随机 — 否决（语义模糊，latest 也可 limit=1）。

### 2. 随机算法：幂次偏置 ID 探测（α=0.65）

**步骤**：

```
1) SELECT MIN(id), MAX(id) FROM ucg_post WHERE status = published

2) 若 min=max=0 或无行 → 返回空 list

3) U ~ Uniform(0,1)  （crypto/rand）

4) R = minId + floor((maxId - minId) * U^α)   ，α = 0.65

5) SELECT post_id, content, media_type, cover_object_key...
   FROM ucg_post p
   WHERE status = published AND id >= R
   ORDER BY id ASC
   LIMIT 1
```

**α=0.65 含义**：比均匀随机（α=1）略偏 high-id（≈新帖）；比 α=0.5 温和。单常量，首期硬编码于 `post_sample_internal.go`。

**备选**：

- `ORDER BY RAND() LIMIT 1` — 否决（帖量大时全表排序）。
- 双区间热区混合 — 否决（多一次 bounds 查询；首期不做）。

### 3. 实现位置

- 新增 `SampleRandomPublishedPost(ctx context.Context) ([]PostSampleItem, error)`于 `internal/services/ucg/post_sample_internal.go`。
- 抽取共用 `postSampleSelectFields()` / `postSampleFromRows()`，避免与 latest 路径重复 SQL 片段。
- Controller 根据 `mode` 分支调用；契约更新 `api/v1/ucg_internal_posts_sample_http.go`（`Mode string` 字段）。

### 4. sim T2 调用链

```
RunCommentTask:
  ucgInternalPost(..., { "mode": "random" }, &sample)
  post := sample.List[0]   // 无 rand.Intn
```

空 list 时仍记失败「无已发布帖」。

### 5. 随机源

ucg-service 侧使用 `crypto/rand` 生成 U，避免可预测性；与 sim 侧 `math/rand` 职责分离（选帖随机完全在 ucg 完成）。

## Risks / Trade-offs

- **[Risk] ID 偏置是 published_at 的代理，非严格时间加权** → 自增 id 与发布时间大体正相关；sim 场景可接受。若未来需精确热区语义，另开路径 2 change。
- **[Risk] MIN/MAX 在大表上可能扫较多 published 行** → T2 低频（6h）；`status` 索引存在时成本可控；仍优于 `RAND()`。
- **[Risk] id 空洞使空洞后首帖概率略高** → 探测 `id >= R` 按 id 轴区间分配概率；sim 评论分布偏差可接受。
- **[Risk] minId=maxId 单帖** → 直接探测该 id，行为正确。
- **[Risk] 与 `sim-gentle-polling` spec 并行** → 本 change delta MODIFIED 条款；归档时合并 baseline。

## Migration Plan

1. 先部署 **ucg-service**（sample random 模式）→ 再部署 **sim-user-service**（T2 改请求）。
2. 向后兼容：未传 `mode` 的调用方仍为 latest 行为。
3. 回滚：回退 sim 镜像至 `limit:20`+`rand.Intn`；或 ucg 忽略 unknown mode（若 sim 先回滚）。
4. 验收：T2 日志/任务成功；老帖（非最新 20）可被评论；MySQL 无 `ORDER BY RAND()` 慢查询。

## Open Questions

- α 是否需在后续暴露为 ucg yaml/env — 首期否，design 预留。
- latest 模式是否仍有调用方 — 当前仅 T2，保留仅为兼容。
