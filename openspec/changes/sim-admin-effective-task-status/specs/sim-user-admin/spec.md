## MODIFIED Requirements

### Requirement: sim config PUT response taskSchedule SHALL reflect effective scheduling state

`PUT /sim/admin/api/config` 响应中的 `taskSchedule[]` 每一项 MUST 含：

- `configEnabled`（bool）— `runtime_json` 中对应任务开关（配置层）
- `enabled`（bool）— **自动调度实际是否启用**：MUST 等于 `configEnabled && dbEnabled && serviceEnabled`，其中 `dbEnabled` 为保存后的 `sim_config.enabled`，`serviceEnabled` 为当前进程 `SIM_USER_SERVICE_ENABLED`
- `nextRunHint`（string）— 当 `enabled=false` 时 MUST 说明阻塞原因（任务配置关 / 业务总闸关 / 进程总闸关），MUST NOT 在总闸关闭时给出「约 X 后」类下一跑时间

`name`、`label`、`intervalSec`、`lastRunAt` 语义不变。

#### Scenario: DB off and T4 config on shows not effectively enabled

- **WHEN** PUT 保存 `sim_config.enabled=false` 且 `taskPostVideo=true`
- **THEN** `taskSchedule` 中 `post_video_submit` MUST 含 `configEnabled=true` 且 `enabled=false`，且 `nextRunHint` MUST 表明业务总闸关闭

#### Scenario: Env off and task config on

- **WHEN** 进程 `SIM_USER_SERVICE_ENABLED=false` 且 PUT 保存某任务 `configEnabled=true` 与 `sim_config.enabled=true`
- **THEN** 该任务 `enabled=false` 且 `nextRunHint` MUST 表明进程总闸关闭

#### Scenario: All gates on and task on

- **WHEN** `serviceEnabled=true`、`dbEnabled=true`、任务 `configEnabled=true`
- **THEN** `enabled=true` 且 `nextRunHint` MAY 基于 `lastRunAt` 与周期估算下一跑

### Requirement: sim config PUT effects SHALL note scheduler blocked when gates closed

当保存后存在 `configEnabled=true` 的任务且（`dbEnabled=false` 或 `serviceEnabled=false`）时，`effects[]` MUST 含可读提示：任务开关已保存但自动调度未启动，并 MAY 说明可手动执行任务。

#### Scenario: Effects on DB disabled save

- **WHEN** PUT 将 `enabled=false` 且至少一任务配置开关为 true
- **THEN** `effects` MUST 含业务总闸相关提示

### Requirement: sim admin save result UI SHALL distinguish config vs effective task state

`sim-admin.html` 保存结果中的 `taskSchedule` 表格 MUST 区分 **配置开关**（`configEnabled`）与 **自动调度是否生效**（`enabled`）。当 `enabled=false` 时 UI MUST NOT 仅展示「开」而暗示任务已在运行。

#### Scenario: Save result shows effective off for T4

- **WHEN** 管理员保存 DB 关、T4 配置开
- **THEN** 保存结果表 MUST 显示 T4 配置为开、自动调度为关（或等价文案）
