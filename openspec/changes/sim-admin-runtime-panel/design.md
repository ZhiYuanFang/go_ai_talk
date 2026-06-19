## Context

`sim-user-service` 的运行行为由两层配置驱动：

1. **环境变量**（Compose 注入）：`SIM_USER_SERVICE_ENABLED` 控制 scheduler 是否启动；`SIM_TASK_*`、`SIM_INTERVAL_*`、`SIM_STARTUP_STAGGER_MAX`、`SIM_UCG_RATE_LIMIT_*`、`SIM_EPHEMERAL_CHAT_*` 等控制各任务开关与周期（见 `LoadRuntimeFlags`）。
2. **DB `sim_config`**：页面上可编辑的 `enabled`、`maxSimUsers`，由 `taskEnabled()` 在每次 tick 内判定。

当前 `sim-admin.html` 仅绑定层 2，任务状态以 `JSON.stringify` 展示，运维无法从页面判断 env 是否生效（例如 compose 未透传 `SIM_INTERVAL_REGISTER` 时容器内仍为默认 24h）。

鉴权沿用现有 `X-Admin-Password`（Hub 经 `admin-common.js` 注入），gateway 反代 `/sim/admin/api/*` 至 `sim-user-service`，无需改网关路由。

## Goals / Non-Goals

**Goals:**

- 提供 `GET /sim/admin/api/runtime` 只读快照，字段与进程实际生效值一致（含 `LoadRuntimeFlags` 解析后的 duration 字符串）。
- 管理页展示：进程总开关、DB 业务开关、库名（脱敏）、sim 用户数/上限、任务开关、周期、错峰、出站限速、E1 参数。
- 任务状态改为表格（任务名、上次运行、成功/失败、最近错误）+ pending video 计数。
- 页内提示：修改 env 须 `docker compose ... up -d --force-recreate sim-user-service`；首次 tick 延迟 ≈ 错峰 + 一整轮周期。

**Non-Goals:**

- 从管理页编辑 env 周期或任务开关（避免 LLM 风暴与 compose 漂移）。
- 暴露完整 DSN、密码、`GLM_API_KEY`、`SIM_ADMIN_PASSWORD`、`simUser.defaultPassword`。
- 新增 Redis 缓存或后台轮询刷新 runtime。
- 计入 App usage 统计（Admin 只读，维护型接口）。

## Decisions

### 1. 独立 `GET /sim/admin/api/runtime` 而非扩展现有 `/config`

**选择**：新端点返回 `{ runtime: SimAdminRuntimeDTO }`。

**理由**：`config` 语义为 DB 可写配置；runtime 为进程只读快照，混在同一 PUT/GET 易误导运维以为可在线改周期。

**备选**：嵌入 `GET /status` — 拒绝，status 已承载任务执行记录，合并后 payload 臃肿且刷新语义不同。

### 2. Runtime 组装位置：`simuser.GetRuntimeSnapshot(ctx)`

在 `internal/services/simuser` 新增函数：

- 调用 `LoadRuntimeFlags(ctx)` 获取开关与周期。
- `databaseName`：`dbcfg.DatabaseNameFromLink(os.Getenv("SIM_DB_LINK"))`，空则回退 `APP_DB_LINK` 解析（与 `cmd/sim-user-service` 启动逻辑一致时仅 SIM）。
- `simUserCount`：复用现有 `countSimUsers`（device internal HTTP）。
- `rateLimitRps` / `rateLimitBurst`：读取与 `rate_limit.go` 相同 env 键（可抽小函数 `LoadRateLimitConfig()` 或内联 `envFloat`/`envInt`）。
- `dbEnabled`：同步 `GetConfig(ctx).Enabled`，便于 UI 对照双层开关。
- Duration 字段 JSON 使用 Go `String()`（如 `5m0s`）或统一格式化为人类可读（实现时 `d.String()` 即可）。

**不返回** `DefaultPassword`。

### 3. UI：只读面板 + 表格，沿用 `admin-pages.css`

在 `sim-admin.html` 配置区下方插入 `#runtimePanel`（`dl` 或 grid），加载时 `GET /sim/admin/api/runtime` 与 `config` 并行。

任务状态：遍历 `status.tasks` 渲染 `<table>`，保留「刷新状态」按钮同时刷新 runtime（或单按钮刷新全部）。

双层开关文案示例：

- 进程开关（env）：`SIM_USER_SERVICE_ENABLED`
- 业务开关（DB）：`sim_config.enabled`（与勾选框一致）

### 4. 规格澄清：周期间隔可只读展示

对 `sim-user-admin` 中「MUST NOT 展示任务周期间隔输入框」保持不变；新增「MUST 只读展示当前生效周期」要求，避免与 v1 防编辑策略冲突。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| `countSimUsers` 调用 device 失败导致 runtime 接口 500 | 失败时 `simUserCount` 置 `-1` 并附 `simUserCountError` 可选字段，其余 runtime 仍返回 |
| 展示库名仍可能泄露环境拓扑 | 仅库名，不含 host；与运维已有 DB 认知一致 |
| env 变更后页面仍显示旧值直至 recreate | 页内固定提示 recreate；runtime 反映**当前进程** env，不读 compose 文件 |
| `countSimUsers` 增加 device 负载 | 仅 Admin 手动/进入页时调用，非轮询 |

## Migration Plan

1. 部署新版 `sim-user-service` 镜像（含 runtime API）。
2. 同步 `sim-admin.html` 静态资源（gateway-app 挂载或镜像内 `resource/public`）。
3. 无需 DB 迁移、无需 env 变更。
4. 回滚：旧镜像 + 旧 HTML；新端点消失，页面 JS 应对 404 降级（实现时 try/catch 显示「运行配置不可用」）。

## Open Questions

（无阻塞项；实现按本设计直接进行。）
