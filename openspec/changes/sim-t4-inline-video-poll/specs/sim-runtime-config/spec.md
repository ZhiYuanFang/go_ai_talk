## MODIFIED Requirements

### Requirement: sim-user-service SHALL persist runtime task and interval configuration in database

除既有 `sim_config.enabled`、`max_sim_users` 外，sim 库 MUST 持久化以下运行时项（列或 JSON 字段，语义等价即可）：

- **taskSwitches**：`register`、`comment`、`postImage`、`postVideo`、`chat`、`follow`（bool）；**MUST NOT** 含 `videoPoll`
- **intervals**（秒或 duration）：`register`、`comment`、`postImage`、`postVideo`、`chat`、`follow`、`postVideoPollInterval`、`postVideoPollMaxWait`、`startupStaggerMax`、`ephemeralChatLoop`、`ephemeralChatWindow`；**MUST NOT** 含 `videoPollIdle`、`videoPollActive`
- **rateLimit**：`ucgRateLimitRps`、`ucgRateLimitBurst`

`LoadRuntimeFromDB(ctx)` MUST 组装进程内 `RuntimeFlags`；**DB 优先**，env 仅作迁移期兜底。进程级 **`SIM_USER_SERVICE_ENABLED=false`** MUST 仍为硬闸。

读取旧 `runtime_json` 时：若缺失 `intervalPostVideoPollSec` 且存在 `intervalVideoPollActiveSec`，MAY 将其作为 poll interval 初值；`videoPoll` 与 idle 字段 MUST 忽略。

#### Scenario: DB 运行时覆盖 env

- **WHEN** DB 中 `interval_post_video_poll=60` 且 env 兜底为 120
- **THEN** 新启动的 T4 poll goroutine MUST 使用 60 秒间隔

#### Scenario: 进程总闸仍读 env

- **WHEN** `SIM_USER_SERVICE_ENABLED=false` 且 DB `enabled=true`
- **THEN** scheduler MUST NOT 启动

### Requirement: sim-user-service SHALL support scheduler reload on runtime config save

`PUT /sim/admin/api/config` 持久化后：

- 变更 **taskSwitches**、**T4 调度周期**（`postVideo`）、**rateLimit**、ephemeral、**enabled** → MUST Reload scheduler
- 变更 **仅** `postVideoPollInterval` / `postVideoPollMaxWait` → MUST NOT 触发 scheduler Reload（进行中的 poll 使用启动时快照）
- 变更 **仅** `maxSimUsers` → MAY 跳过 Reload
- LLM lane PUT MUST NOT 触发 Reload

#### Scenario: 修改 postVideoPollMaxWait 不 reload

- **WHEN** PUT 仅变更 `postVideoPollMaxWait`
- **THEN** 响应 MUST 含 `scheduleReloaded=false`

#### Scenario: 修改 postVideo 调度周期触发 reload

- **WHEN** PUT 变更 `interval_post_video`
- **THEN** 响应 MUST 含 `scheduleReloaded=true`

## REMOVED Requirements

### Requirement: videoPoll task switch and idle/active intervals in runtime config

**Reason**: P1 已删除；轮询参数并入 T4 专用字段。

**Migration**: Admin 表单移除 P1 开关与 idle/active；改用 `postVideoPollInterval`、`postVideoPollMaxWait`。
