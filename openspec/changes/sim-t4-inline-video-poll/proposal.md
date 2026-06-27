## Why

当前 T4（`post_video_submit`）仅在视频生成 **提交成功** 时记 `sim_task_run` 成功，实际发帖依赖独立 P1（`video_poll`）定时扫 `sim_video_job`，导致：提交与发帖语义割裂、首次轮询可能延迟数分钟、无总超时、运维难以从任务成功数判断视频是否真正发布。模拟场景下应将「提交 → async-result 轮询 → 上传发帖」合并为一条 T4 流水线，并去掉 P1。

## What Changes

- T4 提交 `SubmitVideoGeneration` 后 **立即** 启动有界 goroutine，按间隔调用 `paas/v4/async-result/{task_id}` 轮询；成功则 OSS + 发帖，失败或超时则 `sim_video_job=skipped`，整次发布记失败。
- **全局单飞**：任意时刻仅允许一条 T4 视频流水线（内存 `videoPostInFlight`）；轮询未结束前调度 tick 与手动触发 MUST skip/拒绝。
- **删除 P1**：移除 `video_poll` 调度任务、`SIM_VIDEO_POLL_ENABLED` / idle-active 双间隔、手动执行入口与 Admin 开关。
- **进程启动清理（方案 B）**：将 `sim_video_job` 中 `pending`/`processing` 一律标 `skipped`（模拟场景可接受丢失），不恢复轮询。
- **Admin 可配 T4 内轮询参数**（DB `runtime_json`）：`postVideoPollInterval`（轮询间隔）、`postVideoPollMaxWait`（最大等待）；替换原 P1 idle/active 字段。
- **sim-admin.html**：去掉 P1 相关表单项；T4 行手动「执行」按钮在 **整条流水线结束**（`sim_task_run` 更新）前保持「执行中…」；**不**新增 `videoPostInFlight` 等 status API 字段。
- `RecordTaskRun("post_video_submit")` 仅在流水线结束（成功发帖或失败/超时）时写入。
- 不新增 Redis；不新增 `*_test.go`。

## Capabilities

### New Capabilities

（无 — 在既有 sim-user-service / sim-runtime-config / sim-user-admin 能力上增量修改。）

### Modified Capabilities

- `sim-user-service`：T4 内联 async-result 轮询；全局单飞；删除 P1；启动丢弃未完成 job；手动/调度共用 `RunPostVideoSubmitTask`。
- `sim-runtime-config`：runtime_json 新增 `postVideoPollInterval`/`postVideoPollMaxWait`；移除 `videoPoll` 开关与 `videoPollIdle`/`videoPollActive`。
- `sim-user-admin`：Admin API 与 `sim-admin.html` 同步上述字段；移除 P1 展示与手动执行；T4 按钮「执行中」覆盖轮询等待期。

## Impact

- **代码**：`internal/services/simuser/`（tasks、scheduler_manager、runtime_config、runtime_api、config_admin、manual_run、store、task_llm）、`api/v1/sim_admin_http.go`、`internal/controller/sim_admin_api.go`、`resource/public/sim-admin.html`。
- **进程**：仅 **sim-user-service**（及静态页经 gateway-app 分发）。
- **DB**：无表结构迁移；`runtime_json` 字段语义变更；启动 UPDATE `sim_video_job`。
- **OpenSpec 基线**：对照 v2.0.5；本变更 delta 挂于 `sim-t4-inline-video-poll`。
- **App usage 统计**：无新增 App HTTP 接口，无需确认。
- **Redis**：不涉及。
