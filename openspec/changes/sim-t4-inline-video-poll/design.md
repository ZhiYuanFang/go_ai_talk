## Context

现状：

- **T4** `RunPostVideoSubmitTask`：random sim → simText → `SubmitVideoGeneration` → `InsertVideoJob(pending)` → `RecordTaskRun(success)`（仅表示提交 OK）。
- **P1** `RunVideoPollTask`：`runAdaptivePeriodicTracked`（idle 10m / active 2m）扫 `ListPendingVideoJobs` → `PollVideoGeneration`（已走 `paas/v4/async-result`）→ 成功发帖 / 失败 skipped；不写 `sim_task_run`。
- Admin 运行时配置在 `sim_config.runtime_json`，经 `PUT /sim/admin/api/config` 在线生效；`sim-admin.html` 可编辑任务开关与周期。
- 手动执行：`StartManualRunAsync` 占用 `manualBusy` 至 `RunTaskByName` 返回；前端 `runningTasks` 在 `lastRunAt` 变化后清除「执行中…」（当前 T4 提交即更新 lastRunAt，与业务语义不符）。

约束：`AGENTS.md` 背景任务须 OpenSpec 批准；本变更为 **T4 触发的有界 goroutine**（deadline = `postVideoPollMaxWait`），非 ticker 扫表。

## Goals / Non-Goals

**Goals:**

- 一条 T4 流水线覆盖 submit → poll → post；超时即失败。
- 全局单飞，轮询期间 T4 不得再次执行。
- 启动时丢弃未完成 job（B），不 resume。
- Admin 可在线配置轮询间隔与最大等待。
- 手动执行 T4 时，按钮「执行中…」持续到流水线结束（靠延后 `RecordTaskRun` + 前端 status 轮询）。

**Non-Goals:**

- 多视频并发（全局仅一条流水线）。
- 重启后恢复轮询（不做方案 A）。
- 新增 status API 字段暴露 inFlight。
- 修改 Zhipu 视频生成 API 或 LLM lane 配置。
- Redis 读缓存。

## Decisions

### 1. T4 内联 poll goroutine（废除 P1）

**选择**：`InsertVideoJob` 成功后设 `videoPostInFlight=true`，启动 `spawnVideoPoll(sess, job, flags)`；轮询逻辑自 P1 迁入（upload + `CreatePost`），使用 submit 时已持有的 `loginSession`，不再 `accountForWx` 扫 list。

**调度路径**：`go spawnVideoPoll(...)`，不阻塞 scheduler tick。

**手动路径**：`isManualRun(ctx)` 时在同 goroutine 内 `spawnVideoPoll` **同步执行**（或 `<-done`），使 `StartManualRunAsync` 的 `defer EndManualRun` 在流水线结束后才释放。

**备选**：保留 P1 作兜底 — 否决（与用户「不留 P1」冲突）。

### 2. 全局单飞互斥

**选择**：包级 `videoPollMu` + `videoPostInFlight bool`。T4 入口首检：true → skip，`RecordTaskRun(true, "video poll in progress")` 或仅 debug log（与现「已有 pending job」一致记 success+msg）。

**保留** `HasPendingVideoJob(wxID)` 作 per-wx 双保险；启动 B 清理后应极少命中。

### 3. 轮询参数与配置存储

| API / JSON 键 | DB 字段 | 默认 | 说明 |
|---------------|---------|------|------|
| `intervals.postVideoPollInterval` | `intervalPostVideoPollSec` | 120（2m） | 两次 async-result 间隔 |
| `intervals.postVideoPollMaxWait` | `intervalPostVideoPollMaxWaitSec` | 1800（30m） | submit 起总 deadline |

**迁移**：读取旧 `intervalVideoPollActiveSec` 若新字段缺失则作 interval 初值；忽略 idle；删除 `videoPoll` bool。

**热更新**：变更 poll 参数 **不** 触发 scheduler Reload（进行中的 goroutine 使用启动时快照的 flags）。

**env 兜底**（仅 seed / 空 DB）：`SIM_POST_VIDEO_POLL_INTERVAL`、`SIM_POST_VIDEO_POLL_MAX_WAIT`；移除 `SIM_VIDEO_POLL_*`。

### 4. 启动清理（方案 B）

**选择**：`InitScheduler` 前（或 `Start` 内）执行：`UPDATE sim_video_job SET status='skipped' WHERE status IN ('pending','processing')`（可选 `last_error='startup discard'`）。

**理由**：模拟场景可接受丢失；避免僵尸 job 阻塞 `HasPendingVideoJob`。

### 5. RecordTaskRun 语义

**选择**：仅在流水线结束写一次 — 成功发帖 `true`；API failed / upload failed / timeout `false` + 错误摘要。submit 阶段失败仍在 tick 内立即写 `false`。

**不再** 在 submit 成功时写 success。

### 6. Admin UI：T4「执行中」无新 API 字段

**选择**：

- 后端：手动 T4 同步等待 poll 结束 + 延后 `RecordTaskRun`。
- 前端：`maybeClearRunningFromStatus` 仍依据 `lastRunAt` 变化清除「执行中…」；对 `post_video_submit` **延长** status 轮询上限（例如 `ceil(maxWait/5s)+6`，或固定 ≥40min），避免默认 2min 后按钮误恢复可点。

**不** 在 `GET /sim/admin/api/status` 增加 `videoPostInFlight`。

### 7. 删除 P1 连带项

- `scheduler_manager.go`：移除 `runAdaptivePeriodicTracked` for video_poll。
- `manual_run.go`：`RunnableTaskNames` 去掉 `video_poll`。
- `config_admin.go`：`taskSchedule` 表去掉 P1 行。
- `task_llm.go`：去掉 `video_poll` 键。
- `sim-admin.html`：任务开关/周期/手动执行/runtime 只读区同步删除 P1，新增 T4 poll 两项。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 进程 crash 中途中断 | B 启动丢弃；接受丢失 |
| 全局单飞吞吐低 | T4 调度周期 6.5h，模拟场景可接受 |
| 手动执行轮询很久 UI 假释放 | 延长 status poll 周期上限 |
| Admin 改 maxWait 不影响进行中 poll | design 已说明快照语义 |
| 与 `sim-gentle-polling` P1 自适应决策冲突 | 本 change 显式 supersede |

## Migration Plan

1. 部署 sim-user-service + 更新 `sim-admin.html`。
2. 首次启动自动 skipped 旧 pending job。
3. Admin 保存一次 config（可选）以写入新 poll 字段默认值。
4. 回滚：旧镜像恢复 P1；遗留 skipped job 无需处理。

## Open Questions

（无 — 全局单飞、B 清理、按钮执行中均已确认。）
