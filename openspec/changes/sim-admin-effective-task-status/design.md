## Context

双层（实为三层）开关：

1. `SIM_USER_SERVICE_ENABLED`（env，进程总闸）
2. `sim_config.enabled`（DB，业务总闸）
3. `runtime_json.task*`（各任务配置开关）

`buildTaskSchedule` 当前仅读第 3 层作为 `Enabled`，导致保存结果与 `startSchedulerGoroutines` 行为脱节。手动「执行」仍 bypass 总闸，展示层须区分 **自动调度生效** vs **配置保留** vs **可手动测**。

## Goals / Non-Goals

**Goals:**

- 保存后 `taskSchedule` 的 `enabled` = **自动调度是否真的会跑该任务**。
- `configEnabled` = **runtime_json 中任务开关**（保存后的配置意图）。
- `nextRunHint` 在未生效时给出可读的阻塞原因。
- Admin 保存结果表格一眼可辨「配了但没跑」。

**Non-Goals:**

- 合并 env 与 DB 为单一 Admin 开关。
- 改变 `taskEnabled()` / scheduler 判定逻辑。
- 在 status 表或 runtime 只读区大改（本变更聚焦 **PUT config 的 taskSchedule + 保存结果 UI**）。

## Decisions

### 1. 生效开关计算（方案 A）

```text
configEnabled  = runtime_json.taskX（现有逻辑）
effectiveEnabled = configEnabled && dbEnabled && serviceEnabled
```

`TaskScheduleItem.Enabled` **语义变更**（对 Admin 消费者）：由「配置开关」改为 **`effectiveEnabled`**。新增字段 **`ConfigEnabled`** 承载原语义。

**备选**：仅改 hint、不改 Enabled — 否决（用户明确要求开关列不应显示「开」）。

### 2. nextRunHint（方案 C）

优先级：

| 条件 | nextRunHint |
|------|-------------|
| `!configEnabled` | `已关闭` |
| `configEnabled && !serviceEnabled` | `进程总闸关闭（SIM_USER_SERVICE_ENABLED=false），未调度` |
| `configEnabled && !dbEnabled` | `业务总闸关闭（sim_config.enabled=false），未调度` |
| 三者皆 true | 沿用现有 `nextRunHint`（周期 / 下一跑估算） |

### 3. buildTaskSchedule 入参

**选择**：`buildTaskSchedule(ctx, rt RuntimeConfigDB, dbEnabled, serviceEnabled bool)`。

- `dbEnabled`：来自本次 PUT 的 `req.Enabled`（保存后值）或 `GetFullConfig`。
- `serviceEnabled`：`LoadRuntimeFlags(ctx).Enabled`（当前进程 env，只读）。

Reload 与 no-reload 路径均传入 **保存后的** `dbEnabled`。

### 4. API DTO

`SimAdminTaskScheduleDTO` 增加：

- `configEnabled`（bool）— 配置层
- `enabled`（bool）— **生效层**（与上表 `effectiveEnabled` 一致，Breaking 仅影响 Admin 消费方对字段语义的解读）

可选 `blockedBy`（string enum: `""` | `service` | `db`）— 若 hint 已足够则可省略，首版 **不增** blockedBy，靠 hint 即可。

### 5. effects 补充

当 `!req.Enabled` 且任一 `task*` 为 true：

```text
kind: scheduler_not_running
message: 业务总闸已关闭，任务开关仅作配置保留，自动调度未启动；可手动执行任务
```

当 `serviceEnabled=false` 且任一 task config 开：

```text
message: 进程总闸关闭，自动调度未启动；可手动执行任务或修改 env 后 recreate
```

### 6. sim-admin.html

保存结果表头建议：

| 任务 | 配置 | 自动调度 | 下一跑 |
|------|------|----------|--------|

- **配置** ← `configEnabled`
- **自动调度** ← `enabled`（effective）
- 或使用单列「自动调度」+ 小字配置状态 — 实现时选双列更清晰。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 依赖方把 `enabled` 当配置开关 | 变更仅 Admin；文档 + `configEnabled` 显式 |
| env 在容器外不可改，hint 仍显示进程关 | 与 runtime 只读区一致，已预期 |
| 手动执行仍可用，用户以为「关=完全不能跑」 | effects / 页内已有手动测试说明 |

## Migration Plan

1. 部署 sim-user-service + sim-admin.html。
2. 无 DB 迁移；Admin 客户端若缓存旧语义，刷新页面即可。
3. 回滚：旧 DTO 无 `configEnabled`，前端忽略未知字段。

## Open Questions

（无。）
