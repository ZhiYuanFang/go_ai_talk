## ADDED Requirements

### Requirement: sim admin API SHALL expose read-only runtime snapshot

`sim-user-service` MUST 提供 `GET /sim/admin/api/runtime`（鉴权与现有 sim-admin API 一致）。响应 MUST 包含只读字段，反映**当前进程**生效值：

- `serviceEnabled`（bool）— 对应 `SIM_USER_SERVICE_ENABLED`
- `dbEnabled`（bool）— 对应 `sim_config.enabled`
- `databaseName`（string）— 自 `SIM_DB_LINK` 解析的库名，MUST NOT 含账号密码或 host
- `simUserCount`（int）— 当前 `is_simulated=1` 用户数；拉取失败时可为 `-1`
- `maxSimUsers`（int）— 来自 `sim_config`
- `taskSwitches`（object）— 键至少含：`register`、`comment`、`postImage`、`postVideo`、`chat`、`follow`、`videoPoll`；值为 bool，对应 `SIM_TASK_*` / `SIM_VIDEO_POLL_ENABLED`
- `intervals`（object）— 键至少含：`register`、`comment`、`postImage`、`postVideo`、`chat`、`follow`、`videoPollIdle`、`videoPollActive`、`startupStaggerMax`、`ephemeralChatLoop`、`ephemeralChatWindow`；值为字符串时长（如 `5m`、`6h30m`）
- `rateLimitRps`（number）、`rateLimitBurst`（int）— 对应 `SIM_UCG_RATE_LIMIT_*`

响应 MUST NOT 含 DSN 凭据、`SIM_ADMIN_PASSWORD`、`GLM_API_KEY` 或默认登录密码明文。

#### Scenario: Admin reads runtime after login

- **WHEN** 已鉴权管理员 GET `/sim/admin/api/runtime`
- **THEN** 响应 MUST 返回上述字段且 `intervals.register` 与容器内 `SIM_INTERVAL_REGISTER` 解析结果一致

#### Scenario: Runtime excludes secrets

- **WHEN** 管理员 GET `/sim/admin/api/runtime`
- **THEN** 响应 body MUST NOT 包含 `@tcp(` 连接串或 `password` 字段

### Requirement: sim admin UI SHALL display read-only runtime configuration

`sim-admin.html` MUST 在可编辑配置区下方展示「运行配置（只读）」区块，数据来自 `GET /sim/admin/api/runtime`。

页面 MUST：

- 区分展示进程总开关（`serviceEnabled`）与 DB 业务开关（`dbEnabled`，与勾选框一致）
- 展示 `databaseName`、`simUserCount` / `maxSimUsers`
- 只读列出各 `taskSwitches` 与 `intervals`（MUST NOT 提供输入框或保存按钮修改周期）
- 展示 `rateLimitRps`、`rateLimitBurst` 与 E1 相关 `ephemeralChatLoop`、`ephemeralChatWindow`
- 含简短说明：修改 env 后须 `force-recreate sim-user-service`；任务首次执行延迟约为错峰加时长的整周期

#### Scenario: Runtime panel visible

- **WHEN** 管理员打开 `/device/admin/sim-admin.html` 且 runtime API 可用
- **THEN** 页面 MUST 渲染运行配置区块且周期间隔为只读文本而非表单控件

#### Scenario: No interval editor on runtime panel

- **WHEN** 管理员查看运行配置区块
- **THEN** UI MUST NOT 提供修改 `intervals` 或 `taskSwitches` 的输入控件

### Requirement: sim admin UI SHALL show structured task status

`sim-admin.html` MUST 将 `GET /sim/admin/api/status` 结果以结构化形式展示（表格或等价列表），字段至少含：任务名、上次运行时间、成功次数、失败次数、最近错误；并单独展示 `pendingVideoJobs`。

原始 JSON dump  alone MUST NOT 作为唯一展示方式（可保留「查看原始 JSON」折叠为可选）。

#### Scenario: Task status table

- **WHEN** 管理员加载页面或点击刷新状态
- **THEN** 各 `status.tasks` 条目 MUST 以可读表格行展示而非仅 `JSON.stringify` 整块输出

## MODIFIED Requirements

### Requirement: Task intervals SHALL NOT be editable from admin UI in v1

任务周期（24h/6h/3.5h/6.5h/1h/7h/1min 及 `sim-gentle-polling` 引入的可配置 `SIM_INTERVAL_*`）MUST 由代码或环境变量固定；模拟管理页 MUST NOT 提供周期**编辑**控件以免误配导致 LLM 风暴。

管理页 MUST 只读展示当前生效周期（见 runtime API / 运行配置区块），展示不构成可编辑能力。

#### Scenario: No interval editor

- **WHEN** 管理员打开 sim-admin 配置页
- **THEN** UI MUST NOT 展示任务周期间隔输入框或保存周期按钮

#### Scenario: Read-only interval display allowed

- **WHEN** 管理员查看运行配置区块
- **THEN** UI MUST 以只读文本展示各任务当前周期间隔
