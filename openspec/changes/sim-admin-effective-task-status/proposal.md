## Why

sim-admin 保存配置后 `taskSchedule` 的「开关」列仅反映 `runtime_json` 中各任务 bool（如 `taskPostVideo`），未叠加 `sim_config.enabled` 与 `SIM_USER_SERVICE_ENABLED`。运维在 **DB 业务关 + T4 配置开** 或 **env 关 + 配置开** 时，保存结果仍显示「T4 开」并给出「约 6h 后」类下一跑提示，与 scheduler 实际不启动的行为不一致，易误判任务已在跑。

## What Changes

- `PUT /sim/admin/api/config` 响应 `taskSchedule[]`：**开关列改为「实际生效」**（`taskSwitch && dbEnabled && serviceEnabled`）；新增 **`configEnabled`** 保留配置层开关供对照。
- **`nextRunHint`**：配置开但总闸关时 MUST 说明原因（业务总闸 / 进程总闸），MUST NOT 给出虚假下一跑时间。
- 保存 **`effects`**：当 `sim_config.enabled=false` 且存在任务配置为开时，追加提示「任务开关已保存，自动调度未启动」。
- **`sim-admin.html`**：保存结果表展示「生效 / 配置」或等价文案；总闸关时不将任务显示为已启用。
- 不改动 scheduler 运行逻辑；不改动手动「执行」语义（仍 bypass 总闸）。
- 不新增 `*_test.go`。

## Capabilities

### New Capabilities

（无。）

### Modified Capabilities

- `sim-user-admin`：`taskSchedule` 生效语义、hint 与 Admin UI 保存结果展示。

## Impact

- **代码**：`internal/services/simuser/config_admin.go`、`api/v1/sim_admin_http.go`、`internal/controller/sim_admin_api.go`、`resource/public/sim-admin.html`。
- **进程**：仅 **sim-user-service**（及静态页经 gateway-app 分发）。
- **DB / env**：无迁移；只读 `serviceEnabled` 自 env。
- **App usage 统计**：无新增 App HTTP 接口。
