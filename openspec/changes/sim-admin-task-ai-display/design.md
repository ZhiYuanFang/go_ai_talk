## Context

sim 调度任务（T1–T6、P1）在 `internal/services/simuser/tasks.go` 中通过 `aimodel.Invoke` / `GenerateImage` / `SubmitVideoGeneration` / `PollVideoGeneration` 调用不同 lane。LLM 配置已在 `sim_llm_lane_config` 与 ai-model-admin 统一管理；sim-admin 已有任务状态表与 runtime 只读区，但缺少「任务 ↔ AI」视图。

## Goals / Non-Goals

**Goals:**

- 单一数据源：Go 侧 `taskLLMUsage` 映射与 `tasks.go` 保持一致，避免前端硬编码漂移。
- status 刷新时同步展示当前生效 provider/model（读 DB lane store，与 runtime 一致）。
- UI 可读：每条 lane 带用途标签（昵称、头像、评论、E1 回复等）。

**Non-Goals:**

- 不在 sim-admin 编辑 lane 或并发（仍走 ai-model-admin）。
- 不展示 API Key、inflight/waiters（ai-model-admin 职责）。
- 不为 Prompt 模板类型（`register_nickname` 等）单独建映射表（调度任务粒度即可）。

## Decisions

### 1. 映射放后端 `task_llm.go`，非前端静态表

**选择**：`simuser.taskLLMUsage` + `BuildTaskAIModelCatalog(lanes)`。

**理由**：`tasks.go` 变更 lane 时同包维护映射，评审可 grep；前端只渲染 API 字段。

**备选**：前端 `runnableTasks` + 静态 lane 列表 — 拒绝，易与 Go 漂移。

### 2. 字段挂在 status 与 runtime，而非仅 runtime

**选择**：`taskAiModels map[string][]{laneKey, usage, provider, model}` 同时出现在 `GET /status` 与 `GET /runtime`。

**理由**：状态表每 5s 轮询 status；含 catalog 后改 ai-model-admin 后刷新状态即可见，无需额外请求。

**备选**：仅 runtime — 手动执行轮询时 AI 列不更新。

### 3. follow 与 chat_scan 语义

| 任务 | 展示 |
|------|------|
| `follow` | 空数组 → UI「—」 |
| `chat_scan` | 仅 E1 goroutine 用 `simText`；扫描本身无 LLM，usage 标「E1 回复」 |
| `video_poll` | `simVideoGen`（轮询）；提交在 T4 |

### 4. 任务 → lane 对照（实现契约）

```
register          → simText(昵称), simImageGen(头像)
comment           → simVision(评论)
post_image        → simText(文案), simImageGen(配图)
post_video_submit → simText(文案), simVideoGen(提交生成)
video_poll        → simVideoGen(轮询结果)
chat_scan         → simText(E1 回复)
follow            → (无)
```

## Risks / Trade-offs

- **[映射与 tasks.go 不同步]** → 代码评审检查；映射与 `Run*Task` 同目录。
- **[lane 未配置时 provider/model 为空]** → UI 显示「（未配置）」；不阻塞状态表其他列。
- **[PollVideoGeneration 不经 lane gate]** → 展示仍标 simVideoGen，与 T4 提交侧配置一致；轮询 API 用全局 zhipu key 为实现细节，不在本页展开。

## Migration Plan

1. 部署 `sim-user-service` + gateway-app（含静态页）。
2. 无需 DB 迁移。
3. 回滚：移除 API 字段与 UI 列即可，无数据副作用。

## Open Questions

（无）
