## MODIFIED Requirements

### Requirement: sim admin API SHALL expose per-task AI model catalog

`GET /sim/admin/api/status` 与 `GET /sim/admin/api/runtime` 响应 MUST 含 `taskAiModels`（object）：键为调度任务名（至少含 `register`、`comment`、`post_image`、`post_video_submit`、`post_debate`、`debate_comment`、`chat_scan`、`follow`），值为数组。**MUST NOT** 含 `video_poll` 键。

数组元素 MUST 含 `laneKey`、`usage`（可选）、`provider`、`model`。`post_video_submit` MUST 含 simText 与 simVideoGen 相关条目。`post_debate` 与 `debate_comment` MUST 含 simText 条目。

#### Scenario: Status includes T7 T8 models

- **WHEN** GET `/sim/admin/api/status`
- **THEN** `taskAiModels.post_debate` 与 `taskAiModels.debate_comment` MUST 存在且含 simText

#### Scenario: Status returns AI for post video task

- **WHEN** 已鉴权 GET `/sim/admin/api/status`
- **THEN** `taskAiModels.post_video_submit` MUST 含 simText 与 simVideoGen 信息且 MUST NOT 存在 `video_poll` 键

### Requirement: sim admin API SHALL expose config and prompts

`sim-user-service` MUST 提供 Admin HTTP API（经 gateway 反代或直连，鉴权与 ucg-admin 对齐）：

- `GET/PUT /sim/admin/api/config` — 字段至少含 `enabled`（bool）、`maxSimUsers`（int，默认 100）
- `GET/PUT /sim/admin/api/prompts/{taskType}` — `taskType` 至少含：`register_nickname`、`register_avatar`、`comment`、`post_image_text`、`post_video_text`、`post_debate_text`、`debate_comment`、`chat_reply`；每项含 `systemPrompt`、`userPromptTemplate`
- `GET /sim/admin/api/status` — 各任务上次运行时间、成功/失败计数、pending video job 数
- `GET /sim/admin/api/users` — 分页模拟用户列表（见 ADDED Requirement）
- `POST /sim/admin/api/users/{wxId}/deactivate` — 注销单个模拟用户（见 ADDED Requirement）

Prompt 变更 MUST 在下一任务 tick 生效，无需重启进程（读 DB + 短 TTL 缓存可接受）。

#### Scenario: Debate prompts editable

- **WHEN** GET `/sim/admin/api/prompts/debate_comment`
- **THEN** MUST 返回 userPromptTemplate 且保存后下一次 T8 MUST 使用新模板

#### Scenario: Update max sim users

- **WHEN** 管理员 PUT `maxSimUsers=50`
- **THEN** 下一次 T1 MUST 在 sim 用户数 ≥50 时停止注册

#### Scenario: Update comment prompt

- **WHEN** 管理员修改 `comment` 的 `userPromptTemplate`
- **THEN** 下一次 T2 MUST 使用新模板渲染变量

### Requirement: sim-admin runtime panel SHALL display effective task schedule

`GET /sim/admin/api/runtime` 与 status 中 `intervals` 键 MUST 含 `register`、`comment`、`postImage`、`postVideo`、`postDebate`、`debateComment`、`chat`、`follow`、`postVideoPollInterval`、`postVideoPollMaxWait`、`startupStaggerMax`；**MUST NOT** 含 `ephemeralChatLoop`、`ephemeralChatWindow`、`videoPollIdle`、`videoPollActive`。

`taskSchedule[]`（config PUT 响应与 admin UI）MUST 含 `post_debate`（T7 辩论发帖）与 `debate_comment`（T8 辩论论点）项，含 `configEnabled`、`enabled`、`intervalSec`、`nextRunHint`。

#### Scenario: Admin shows T7 T8 schedule

- **WHEN** GET config 或 PUT 保存后返回 `taskSchedule`
- **THEN** MUST 含 name=`post_debate` 与 name=`debate_comment`

#### Scenario: Runtime intervals without ephemeral

- **WHEN** GET runtime
- **THEN** `intervals` MUST NOT 含 E1 键
