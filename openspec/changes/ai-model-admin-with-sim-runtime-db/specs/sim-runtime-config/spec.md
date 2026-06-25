## ADDED Requirements

### Requirement: sim-user-service SHALL persist runtime task and interval configuration in database

除既有 `sim_config.enabled`、`max_sim_users` 外，sim 库 MUST 持久化以下运行时项（列或 JSON 字段，语义等价即可）：

- **taskSwitches**：`register`、`comment`、`postImage`、`postVideo`、`chat`、`follow`、`videoPoll`（bool）
- **intervals**（秒或 duration）：`register`、`comment`、`postImage`、`postVideo`、`chat`、`follow`、`videoPollIdle`、`videoPollActive`、`startupStaggerMax`、`ephemeralChatLoop`、`ephemeralChatWindow`
- **rateLimit**：`ucgRateLimitRps`、`ucgRateLimitBurst`

`LoadRuntimeFromDB(ctx)` MUST 组装进程内 `RuntimeFlags`；**DB 优先**，env（`SIM_TASK_*`、`SIM_INTERVAL_*` 等）仅作迁移期兜底。进程级 **`SIM_USER_SERVICE_ENABLED=false`** MUST 仍为硬闸，阻止 scheduler 启动（保留 env，不迁入 DB）。

#### Scenario: DB 运行时覆盖 env

- **WHEN** DB 中 `interval_comment=7200` 且 env `SIM_INTERVAL_COMMENT=3600`
- **THEN** 调度 MUST 使用 7200 秒

#### Scenario: 进程总闸仍读 env

- **WHEN** `SIM_USER_SERVICE_ENABLED=false` 且 DB `enabled=true`
- **THEN** scheduler MUST NOT 启动

### Requirement: sim-user-service SHALL support scheduler reload on runtime config save

sim-user-service MUST 提供 **SchedulerManager**：持有可 cancel 的 scheduler context 与 WaitGroup。`PUT /sim/admin/api/config`（扩展 body 含 taskSwitches/intervals/rateLimit）在持久化 DB 后 MUST：

- 若变更含 **调度类字段**（taskSwitches、intervals、rateLimit、ephemeral 参数、`enabled`）→ MUST `Reload`：Stop（cancel + wait）→ LoadRuntimeFromDB → Start（`skipStagger=true`，跳过长错峰 `startupStaggerMax`）
- 若 **仅** `maxSimUsers` 变更 → MAY 跳过 Reload（注册任务 tick 内已读 config）
- LLM lane PUT MUST NOT 触发 Reload

#### Scenario: 修改 comment 周期触发 reload

- **WHEN** PUT config 将 `interval_comment` 从 6h 改为 3h
- **THEN** 响应 MUST 含 `scheduleReloaded=true` 且新 goroutine MUST 使用 3h 周期

#### Scenario: 仅改 maxSimUsers 不 reload

- **WHEN** PUT 仅变更 `maxSimUsers`
- **THEN** 响应 MAY 含 `scheduleReloaded=false` 且 scheduler goroutine MUST NOT 全量重启

#### Scenario: 热 reload 跳过长错峰

- **WHEN** Admin 触发 Reload
- **THEN** 各任务首轮 MUST NOT 等待完整 `startupStaggerMax`（MAY 使用 0～30s 短延迟）

### Requirement: sim config PUT response SHALL describe save effects

`PUT /sim/admin/api/config` 响应 MUST 扩展：

- `scheduleReloaded`（bool）
- `effects[]`：`kind`（如 `scheduler_reloaded`、`task_interval_changed`、`ephemeral_may_continue`）、可选 `task`、`message`
- `taskSchedule[]`：每任务 `name`、`enabled`、`intervalSec`、`lastRunAt`（来自 `sim_task_run` 若存在）、`nextRunHint`（人类可读，如「约 3h 后」或「保存后立即进入新周期等待」）

系统 MUST NOT 保证强杀进行中的 LLM 调用或已 spawn 的 E1 聊天 goroutine；`effects` MUST 可表达「进行中任务可能跑完旧配置」。

#### Scenario: 关闭 chat 任务的效果提示

- **WHEN** PUT 将 `task_chat=false`
- **THEN** 响应 `effects` MUST 含 chat 相关提示且 `taskSchedule` 中 chat MUST 反映 disabled

#### Scenario: 修改 interval 的下一跑提示

- **WHEN** PUT 缩短 postImage 周期且该任务有 `lastRunAt`
- **THEN** `taskSchedule` 中 postImage 的 `nextRunHint` MUST 基于 `lastRunAt + 新 interval` 估算

### Requirement: docker compose env SHALL omit DB-backed sim runtime and LLM variables

`manifest/docker/docker-compose.microservices.yml` 与 `.env.example` MUST 移除已 DB 化的 sim 变量，至少包括：`SIM_LLM_*`、`SIM_TASK_*`、`SIM_INTERVAL_*`、`SIM_EPHEMERAL_*`、`SIM_STARTUP_STAGGER_MAX`、`SIM_UCG_RATE_LIMIT_*`、`SIM_VIDEO_POLL_ENABLED`。MUST 保留：`SIM_DB_LINK`、`SIM_USER_SERVICE_ENABLED`、`SIM_ADMIN_PASSWORD`、API Key 类变量。

#### Scenario: compose 无 SIM_LLM env

- **WHEN** 运维查看 microservices compose 中 sim-user-service environment
- **THEN** MUST NOT 含 `SIM_LLM_TEXT_PROVIDER` 等 LLM lane env 块
