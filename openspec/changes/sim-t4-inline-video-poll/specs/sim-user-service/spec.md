## ADDED Requirements

### Requirement: T4 post video SHALL inline poll async-result until post or failure

`RunPostVideoSubmitTask`（调度与手动共用）在 `SubmitVideoGeneration` 成功且 `InsertVideoJob` 写入 `pending` 后 MUST 启动视频结果轮询。轮询 MUST 调用智谱 `GET /paas/v4/async-result/{task_id}`（经 `aimodel.PollVideoGeneration`）。轮询 MUST 使用 submit 阶段已获得的 `loginSession`，MUST NOT 再经分页 list 线性查 account。

- **success**：下载视频 → UCG media 上传 → `POST /ucg/app/api/posts`（`submit=true`）→ `sim_video_job=done` → `RecordTaskRun("post_video_submit", true, ...)`
- **failed**（上游明确失败）：`sim_video_job=skipped` → `RecordTaskRun(..., false, ...)`
- **processing / pending**：在 `postVideoPollInterval` 后重试，直至 `now >= submitTime + postVideoPollMaxWait` → 超时视为发布失败 → `skipped` + `RecordTaskRun(..., false, ...)`

MUST NOT 在 submit 成功时单独写 `RecordTaskRun` success。

#### Scenario: Poll success posts video

- **WHEN** T4 提交后 async-result 返回 success 且上传发帖 OK
- **THEN** job MUST 为 `done` 且 `sim_task_run` MUST 记 post_video_submit 成功

#### Scenario: Poll timeout fails task

- **WHEN** 自 submit 起超过 `postVideoPollMaxWait` 仍为 processing
- **THEN** job MUST 为 `skipped` 且 `sim_task_run` MUST 记 post_video_submit 失败

### Requirement: T4 video pipeline SHALL be globally single-flight

任意时刻 sim-user-service 进程内 MUST 最多一条进行中的 T4 视频流水线（submit + poll + post）。`videoPostInFlight`（或等价机制）为 true 时：

- 调度 tick MUST skip 新 submit（MAY 记 success + 说明「video poll in progress」）
- 手动 `POST .../tasks/post_video_submit/run` MUST 拒绝或返回「任务正在执行中」（与 `manualBusy` 语义一致）

流水线结束（成功/失败/超时）后 MUST 清除 inFlight。

#### Scenario: Scheduler skips while poll active

- **WHEN** 上一 T4 流水线仍在轮询且调度 tick 到达
- **THEN** MUST NOT 再次 SubmitVideoGeneration

#### Scenario: Manual run rejected while poll active

- **WHEN** 管理员在流水线进行中再次点击 T4 手动执行
- **THEN** API MUST 返回任务忙错误

### Requirement: sim-user-service SHALL discard pending video jobs on startup

进程 scheduler 启动前（或等效启动钩子）MUST 将 `sim_video_job` 中 `status IN ('pending','processing')` 更新为 `skipped`。MUST NOT 为遗留 job 恢复轮询 goroutine。

#### Scenario: Startup clears stale jobs

- **WHEN** sim-user-service 重启且 DB 存在 pending job
- **THEN** 启动后这些 job MUST 为 skipped 且无恢复轮询

## REMOVED Requirements

### Requirement: P1 video_poll periodic task

**Reason**: 轮询并入 T4 内联 goroutine；模拟场景不需要独立扫表调度。

**Migration**: 删除 `video_poll` scheduler、手动任务与 Admin 开关；运维改用 T4 内 `postVideoPollInterval` / `postVideoPollMaxWait`。
