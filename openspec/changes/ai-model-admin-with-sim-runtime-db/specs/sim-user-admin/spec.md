## MODIFIED Requirements

### Requirement: sim admin API SHALL expose read-only runtime snapshot

`sim-user-service` MUST 提供 `GET /sim/admin/api/runtime`（鉴权与现有 sim-admin API 一致）。响应 MUST 包含字段，反映**当前 DB 生效值**（非 env 只读镜像）：

- `serviceEnabled`（bool）— 对应 `SIM_USER_SERVICE_ENABLED`（env 硬闸）
- `dbEnabled`（bool）— 对应 `sim_config.enabled`
- `databaseName`（string）— 自 `SIM_DB_LINK` 解析的库名，MUST NOT 含账号密码或 host
- `simUserCount`（int）— 当前 `is_simulated=1` 用户数；拉取失败时可为 `-1`
- `maxSimUsers`（int）— 来自 `sim_config`
- `taskSwitches`（object）— 键至少含：`register`、`comment`、`postImage`、`postVideo`、`chat`、`follow`、`videoPoll`
- `intervals`（object）— 键至少含：`register`、`comment`、`postImage`、`postVideo`、`chat`、`follow`、`videoPollIdle`、`videoPollActive`、`startupStaggerMax`、`ephemeralChatLoop`、`ephemeralChatWindow`；值为字符串时长
- `rateLimitRps`（number）、`rateLimitBurst`（int）

响应 MUST NOT 含 DSN 凭据、`SIM_ADMIN_PASSWORD`、`GLM_API_KEY` 或默认登录密码明文。

#### Scenario: Admin reads runtime from DB

- **WHEN** 已鉴权管理员 GET `/sim/admin/api/runtime` 且 DB 中 comment 周期为 3h
- **THEN** 响应 `intervals.comment` MUST 为 3h 且 MUST NOT 依赖 env 覆盖

#### Scenario: Runtime excludes secrets

- **WHEN** 管理员 GET `/sim/admin/api/runtime`
- **THEN** 响应 body MUST NOT 包含 `@tcp(` 连接串或 `password` 字段

## REMOVED Requirements

### Requirement: Task intervals SHALL NOT be editable from admin UI in v1

**Reason**: 运行时与 LLM lane 已迁入 DB，需在线可调周期与任务开关；原只读约束由 `sim-runtime-config` 取代。

**Migration**: 使用 sim-admin 可编辑表单或 `PUT /sim/admin/api/config`；LLM lane 使用 ai-model-admin。

### Requirement: sim admin UI SHALL display read-only runtime configuration

**Reason**: 运行配置改为 DB 可编辑 + 保存生效语义，只读区块不足。

**Migration**: sim-admin 提供运行时表单、调用 PUT config，并展示 `effects` / `taskSchedule` 保存结果面板。

## ADDED Requirements

### Requirement: sim admin UI SHALL edit runtime configuration with save effect feedback

`sim-admin.html` MUST 提供可编辑运行时配置表单，字段覆盖 `taskSwitches`、`intervals`、`rateLimitRps`、`rateLimitBurst`、`ephemeralChatLoop`、`ephemeralChatWindow`（及既有 `enabled`、`maxSimUsers`）。保存 MUST 调用扩展后的 **`PUT /sim/admin/api/config`**。保存成功后 MUST 展示 API 返回的 **`effects`** 与 **`taskSchedule`**（含各任务「立即生效 / 预计下一跑」提示）。页面 MUST 区分 `serviceEnabled`（env，说明须改 env 并 recreate）与 `dbEnabled`（可在线保存）。页面 MUST NOT 提供 sim LLM lane 编辑（链至 ai-model-admin）。`GET /sim/admin/api/status` 结构化任务状态展示 MUST 保留。

#### Scenario: 可编辑任务开关

- **WHEN** 管理员在 sim-admin 取消勾选 chat 并保存
- **THEN** 页面 MUST 调用 PUT config 且 MUST 展示保存结果面板含 schedule 提示

#### Scenario: 保存后展示下一跑提示

- **WHEN** PUT config 返回 `taskSchedule` 含 postImage 的 `nextRunHint`
- **THEN** sim-admin MUST 在保存结果区展示该提示

#### Scenario: Interval inputs are editable

- **WHEN** 管理员查看运行配置区块
- **THEN** UI MUST 提供周期间隔输入控件（非只读文本）
