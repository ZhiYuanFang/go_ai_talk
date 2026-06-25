## ADDED Requirements

### Requirement: sim-user-service SHALL persist sim LLM lanes in database

`sim-user-service` MUST 在 `SIM_DB_LINK` 对应库维护表 **`sim_llm_lane_config`**（或语义等价名），主键 `lane`，取值 `simText`、`simVision`、`simImageGen`、`simVideoGen`。每行 MUST 含 `provider`、`model`、`max_in_flight`、`max_waiters`、`timeout_sec`（可选）、`updated_at`、`updated_by`。

`SimLLMLaneStore.Load(lane)` MUST 按优先级：**DB 行** > **环境变量**（`SIM_LLM_*`，仅 seed/迁移）> **代码默认种子**。EnsureSchema MUST 在无行时写入默认种子（与当前 env 默认语义一致）。

#### Scenario: 新环境 DB 种子

- **WHEN** sim 库尚无 `sim_llm_lane_config` 行且进程首次 EnsureSchema
- **THEN** 四 lane MUST 存在默认行且 Admin GET MUST 可读

#### Scenario: DB 覆盖 env

- **WHEN** DB 行与 `SIM_LLM_*` env 冲突
- **THEN** 运行时 MUST 使用 DB 值

### Requirement: sim-user-service SHALL provide sim LLM lane Admin API

sim-user-service MUST 提供 `GET /sim/admin/api/llm-lanes` 与 `PUT /sim/admin/api/llm-lanes`，鉴权 MUST 与现有 sim-admin 一致（Header `X-Admin-Password`）。响应与请求 MUST 含四 lane 子对象，各含 `provider`、`model`、`maxInFlight`、`maxWaiters`、`updatedAt`、`updatedBy`。GET MUST 返回 provider→model allowlist（含 zhipu 生图/生视频 model）。PUT MUST 校验 allowlist 与边界（`maxInFlight>=1`，`maxWaiters>=0`）。PUT 成功后 MUST 调用 `aimodel.InvalidateLaneCache()` 且 MUST NOT 触发 scheduler reload。

#### Scenario: 管理员读取 sim lanes

- **WHEN** 已鉴权 GET `/sim/admin/api/llm-lanes`
- **THEN** 响应 MUST 含四 lane 配置与 allowlist

#### Scenario: 管理员更新 simText 并发

- **WHEN** PUT `simText.maxInFlight=2` 且 model 在 allowlist
- **THEN** sim-service MUST 持久化且 MUST 失效 lane 缓存

#### Scenario: sim LLM PUT 不 reload scheduler

- **WHEN** 仅 PUT llm-lanes 成功
- **THEN** scheduler goroutine MUST NOT 因该 PUT 而 Stop/Start

### Requirement: aimodel allowlist SHALL include sim image and video generation models

`internal/services/aimodel` 的 `ProviderModels`（zhipu）MUST 包含 `cogview-3-flash`、`cogvideox-flash`（及 normalize 大小写变体），供 simImageGen/simVideoGen Admin PUT 校验通过。

#### Scenario: Admin 配置 sim 生图 model

- **WHEN** PUT simImageGen model=`cogview-3-flash`
- **THEN** 校验 MUST 通过且 MUST 持久化

### Requirement: sim LLM lane Admin API MUST NOT count toward App usage stats

`/sim/admin/api/llm-lanes` MUST NOT 计入 gateway-app App API 使用统计。

#### Scenario: sim llm-lanes PUT 不计 usage

- **WHEN** 管理员 PUT llm-lanes 成功
- **THEN** usage 统计 MUST NOT 递增
