## MODIFIED Requirements

### Requirement: sim-user-service SHALL persist runtime task and interval configuration in database

除既有 `sim_config.enabled`、`max_sim_users` 外，sim 库 MUST 持久化以下运行时项（列或 JSON 字段，语义等价即可）：

- **taskSwitches**：`register`、`comment`、`postImage`、`postVideo`、`chat`、`follow`（bool）；**MUST NOT** 含 `videoPoll`
- **intervals**（秒或 duration）：`register`、`comment`、`postImage`、`postVideo`、`chat`、`follow`、`postVideoPollInterval`、`postVideoPollMaxWait`、`startupStaggerMax`；**MUST NOT** 含 `videoPollIdle`、`videoPollActive`、`ephemeralChatLoop`、`ephemeralChatWindow`
- **rateLimit**：`ucgRateLimitRps`、`ucgRateLimitBurst`

`LoadRuntimeFromDB(ctx)` MUST 组装进程内 `RuntimeFlags`；**DB 优先**，env 仅作迁移期兜底。进程级 **`SIM_USER_SERVICE_ENABLED=false`** MUST 仍为硬闸。

读取旧 `runtime_json` 时：若存在 `ephemeralChatLoopSec`/`ephemeralChatWindowSec` MUST 忽略；若缺失 `intervalPostVideoPollSec` 且存在 `intervalVideoPollActiveSec`，MAY 将其作为 poll interval 初值；`videoPoll` 与 idle 字段 MUST 忽略。

#### Scenario: DB 运行时覆盖 env

- **WHEN** DB 中 `interval_chat` 为 3600 且 env 兜底不同
- **THEN** scheduler MUST 使用 DB 值

#### Scenario: 进程总闸仍读 env

- **WHEN** `SIM_USER_SERVICE_ENABLED=false` 且 DB `enabled=true`
- **THEN** scheduler MUST NOT 启动

#### Scenario: Legacy ephemeral fields ignored

- **WHEN** `runtime_json` 仍含 `ephemeralChatLoopSec`
- **THEN** LoadRuntimeFromDB MUST NOT 映射到 RuntimeFlags

### Requirement: sim-user-service SHALL support scheduler reload on runtime config save

`PUT /sim/admin/api/config` 持久化后，sim-user-service MUST 按下列规则决定是否 Reload scheduler：

- 变更 **taskSwitches**、**intervals**（含 `chat`）、**rateLimit**、**enabled** → MUST Reload scheduler
- 变更 **仅** `postVideoPollInterval` / `postVideoPollMaxWait` → MUST NOT 触发 scheduler Reload
- 变更 **仅** `maxSimUsers` → MAY 跳过 Reload
- LLM lane PUT MUST NOT 触发 Reload

#### Scenario: 修改 chat 周期触发 reload

- **WHEN** PUT 变更 `interval_chat`
- **THEN** 响应 MUST 含 `scheduleReloaded=true`

#### Scenario: 修改 postVideoPollMaxWait 不 reload

- **WHEN** PUT 仅变更 `postVideoPollMaxWait`
- **THEN** 响应 MUST 含 `scheduleReloaded=false`

## REMOVED Requirements

### Requirement: E1 ephemeral chat loop and window in runtime config

**Reason**: E1 已删除；T5 改为单次 poll-reply。

**Migration**: Admin 移除 E1 表单项；env `SIM_EPHEMERAL_CHAT_*` 不再读取；旧 JSON 字段忽略。
