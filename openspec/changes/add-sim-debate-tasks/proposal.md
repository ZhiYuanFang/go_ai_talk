## Why

UCG 辩论帖要求「先投票再评论」，现有 sim T2 评论任务经 `posts/sample` 随机抽帖时可能命中辩论帖，导致 `请先投票后再发表论点` 失败且无法养热辩论区。同时广场缺少稳定的辩论话题供给与 sim 侧论点互动。

## What Changes

- **ucg internal `posts/sample`**：新增 `excludeDebate` / `onlyDebate` 互斥过滤（与 `isDebatePost` 语义一致）；响应增加 `debateLeft`、`debateRight`（辩论帖时非空）。
- **T2 评论**：sample 请求 MUST `excludeDebate=true`（保留既有 `excludeMediaTypes=[2]`）；防御性跳过若仍返回辩论帖。
- **T7 辩论发帖**（`post_debate`）：每 12h（± jitter）随机 sim 用户经 LLM 生成话题正文与左右立场（各 ≤5 字），`POST /ucg/app/api/posts` 提交审核。
- **T8 辩论论点**（`debate_comment`）：每 1h（± jitter）随机 sim 用户经 `onlyDebate` sample 抽帖，先 `POST vote` 再 `POST comments`（论点 ≤10 字）；MUST NOT 对帖作者本人执行（authorWxId ≠ simWxId）。
- **sim runtime / admin**：持久化与展示 T7/T8 开关、周期、prompt、`taskAiModels`、手动执行。

**usage 统计**：无新增 App HTTP 路由；sim 经既有 gateway App API 发帖/投票/评论，**无需**变更 `maintenance_skip.go`。

## Capabilities

### New Capabilities

（无独立新 capability 目录；T7/T8 归入既有 sim 任务规格增量。）

### Modified Capabilities

- `ucg-sim-feed-sample`：sample API 辩论过滤与响应字段。
- `sim-user-service`：T2 排除辩论帖；新增 T7、T8 任务语义与默认周期。
- `sim-runtime-config`：`taskPostDebate`、`taskDebateComment` 及对应 interval 字段/env 默认。
- `sim-user-admin`：admin API/UI 覆盖 T7/T8 与 `post_debate` / `debate_comment` prompt。

## Impact

- **ucg-service**：`post_sample_internal.go`、internal controller、api/v1 类型。
- **sim-user-service**：`tasks.go`、`scheduler_manager.go`、`runtime_config.go`、`runtime.go`、`config_admin.go`、`schema.go`、`task_llm_temp.go`、admin controller/HTML。
- **OpenSpec**：背景循环任务已在 proposal/design 声明宿主与周期，符合 `background-loop-task-governance`。
