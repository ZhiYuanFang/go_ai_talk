## Why

sim T2 评论经 `simVision` 结合封面与正文生成评论，视频帖仅能提供首帧 snapshot，识图质量弱、评论易偏题，且视频内容由 T4 任务覆盖。须在选帖阶段排除 `mediaType=video`，使 T2 仅在纯文字帖与图文帖上评论，避免浪费 LLM 与产生低质量互动。

## What Changes

- ucg internal `posts/sample` 请求体增加可选 `excludeMediaTypes`（整型数组）；random（及 latest 若传该字段）抽样 SQL MUST 排除 listed `media_type`。
- sim T2 调用 sample 时传 `excludeMediaTypes: [2]`（`MediaTypeVideo`）。
- sim 侧可选防御：若仍收到 `mediaType=2` 则 skip 并 warning，不发表评论。
- 当排除后无候选帖时，sample 返回空 list，T2 记失败「无已发布帖」（与现网空广场一致）。
- 不新增 `*_test.go`；不改变 T3/T4 行为。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `ucg-sim-feed-sample`：sample API 支持 `excludeMediaTypes` 过滤；random 探测 MIN/MAX 与 probe 均应用同一 filter。
- `sim-user-service`：T2 sample 请求 MUST 排除视频帖；MUST NOT 对 `mediaType=2` 帖发表评论。

## Impact

- **代码**：`internal/services/ucg/post_sample_internal.go`、`internal/controller/ucg_internal_posts_sample.go`、`api/v1/ucg_internal_posts_sample_http.go`、`internal/services/simuser/tasks.go`。
- **进程**：`ucg-service`、`sim-user-service`；无 DB 迁移。
- **兼容**：未传 `excludeMediaTypes` 时行为与变更前一致。
