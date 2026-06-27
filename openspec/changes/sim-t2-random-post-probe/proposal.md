## Why

sim T2 评论任务当前经 `posts/sample` 取 `published_at DESC` 最新 20 帖后再内存随机，导致全库老帖几乎永不被评论，与「随机选帖互动」的运营目标不符。T2 为低频任务（默认 6h），可在不引入 `ORDER BY RAND()` 全表扫描的前提下，改用主键 ID 探测实现全库可达抽样，并以幂次偏置给予新帖适度权重。

## What Changes

- ucg internal `POST /ucg/internal/api/posts/sample` 新增 `mode: "random"`（或等价语义）：返回 **1 条** 经 ID 探测抽样的已发布帖，覆盖全库 published 集合，默认幂次偏置 α=0.65（略偏新帖）。
- 保留现有 `mode` 缺省 / `latest` 行为：`limit` 批量按 `published_at DESC` 返回（供兼容，当前仅 T2 调用方）。
- sim T2 改请求 `mode=random`（或 `limit=1` + random 模式），移除客户端 `rand.Intn` 选帖逻辑。
- 不新增 Redis 读缓存；不改动 recommend Feed 真人读路径；不新增 `*_test.go`。

## Capabilities

### New Capabilities

（无 — 在既有 `ucg-sim-feed-sample` 与 `sim-user-service` 能力上增量修改。）

### Modified Capabilities

- `ucg-sim-feed-sample`：sample API 增加 random 模式语义；random 路径 MUST 使用有界 ID 探测（MIN/MAX + `id >= R LIMIT 1`），MUST NOT 使用 `ORDER BY RAND()`。
- `sim-user-service`：T2 评论 MUST 使用 sample random 模式获取单条帖子，MUST NOT 在 sim 侧对 latest-20 列表再做随机选取。

## Impact

- **代码**：`internal/services/ucg/post_sample_internal.go`、`internal/controller/ucg_internal_posts_sample.go`、`api/v1/ucg_internal_posts_sample_http.go`、`internal/services/simuser/tasks.go`。
- **进程**：`ucg-service`、`sim-user-service`；无 DB 迁移。
- **配置**：可选常量/包内默认 α=0.65；首期不暴露 env（design 可预留）。
- **OpenSpec**：delta 挂于本 change；与 in-progress `sim-gentle-polling` 归档时合并 baseline。
