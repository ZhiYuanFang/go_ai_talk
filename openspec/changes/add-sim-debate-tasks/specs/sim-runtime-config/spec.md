## MODIFIED Requirements

### Requirement: sim-user-service SHALL persist runtime task and interval configuration in database

除既有 `sim_config.enabled`、`max_sim_users` 外，sim 库 MUST 持久化以下运行时项（列或 JSON 字段，语义等价即可）：

- **taskSwitches**：`register`、`comment`、`postImage`、`postVideo`、`postDebate`、`debateComment`、`chat`、`follow`（bool）；**MUST NOT** 含 `videoPoll`
- **intervals**（秒或 duration）：`register`、`comment`、`postImage`、`postVideo`、`postDebate`、`debateComment`、`chat`、`follow`、`postVideoPollInterval`、`postVideoPollMaxWait`、`startupStaggerMax`、`ephemeralChatLoop`、`ephemeralChatWindow`；**MUST NOT** 含 `videoPollIdle`、`videoPollActive`
- **rateLimit**：`ucgRateLimitRps`、`ucgRateLimitBurst`

`LoadRuntimeFromDB(ctx)` MUST 组装进程内 `RuntimeFlags`；**DB 优先**，env 仅作迁移期兜底。进程级 **`SIM_USER_SERVICE_ENABLED=false`** MUST 仍为硬闸。

读取旧 `runtime_json` 时：若缺失 `intervalPostVideoPollSec` 且存在 `intervalVideoPollActiveSec`，MAY 将其作为 poll interval 初值；`videoPoll` 与 idle 字段 MUST 忽略。缺失 `taskPostDebate` / `taskDebateComment` / 对应 interval 时 MUST 回退 env 默认（T7/T8 默认开启、12h/1h）。

#### Scenario: DB 运行时覆盖 env

- **WHEN** DB 中 `interval_post_debate=43200` 且 env 兜底为 12h
- **THEN** 新启动的 T7 goroutine MUST 使用 DB 值

#### Scenario: 进程总闸仍读 env

- **WHEN** `SIM_USER_SERVICE_ENABLED=false` 且 DB `enabled=true`
- **THEN** scheduler MUST NOT 启动

## MODIFIED Requirements

### Requirement: sim task intervals SHALL be overridable via environment variables

各背景任务周期 MUST 支持环境变量覆盖，未设置或非法值时 MUST 回退下列默认值：`SIM_INTERVAL_REGISTER=24h`、`SIM_INTERVAL_COMMENT=6h`、`SIM_INTERVAL_POST_IMAGE=3h30m`、`SIM_INTERVAL_POST_VIDEO=6h30m`、`SIM_INTERVAL_POST_DEBATE=12h`、`SIM_INTERVAL_DEBATE_COMMENT=1h`、`SIM_INTERVAL_CHAT=1h`、`SIM_INTERVAL_FOLLOW=7h`。周期执行 MUST 保留 ±10% jitter。

#### Scenario: Default intervals preserved

- **WHEN** 未设置任何 `SIM_INTERVAL_*` 环境变量
- **THEN** T1–T8 名义周期 MUST 与上表一致

#### Scenario: Custom debate comment interval

- **WHEN** `SIM_INTERVAL_DEBATE_COMMENT=30m` 且服务已启动
- **THEN** T8 两次成功执行间隔 MUST 约为 30m（含 jitter）
