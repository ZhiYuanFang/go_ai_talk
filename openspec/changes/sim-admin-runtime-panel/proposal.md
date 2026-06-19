## Why

模拟用户管理页（`sim-admin.html`）目前仅展示 DB 可编辑项（`enabled`、`maxSimUsers`）与原始 JSON 任务状态，无法反映 **进程级环境变量**（`SIM_USER_SERVICE_ENABLED`、任务开关、`SIM_INTERVAL_*`、限速、数据库名、当前 sim 用户数等）。运维在排查「配置了周期但任务无动静」时，必须在容器内 `printenv` 与日志对照，易误判为服务故障。在管理页只读展示运行时配置，可显著降低排障成本，并与 `sim-gentle-polling` 引入的可配置 env 对齐。

## What Changes

- 新增 Admin API `GET /sim/admin/api/runtime`：返回当前进程生效的运行时配置（只读），含双层开关说明字段、任务开关、周期、错峰、限速、E1 窗口、数据库名（脱敏）、当前 sim 用户数。
- 扩展 `sim-admin.html`：新增「运行配置（只读）」面板；将任务状态由原始 JSON 改为结构化表格；附注「修改 env 后须 `force-recreate sim-user-service`」。
- **不**提供周期或 env 开关的页面编辑（延续 v1 防误配策略）；**不**暴露 DSN 密码、`SIM_ADMIN_PASSWORD`、`DefaultPassword`、API Key。
- 澄清规格：`sim-user-admin` 禁止的是周期间隔**输入框**，允许只读展示。

## Capabilities

### New Capabilities

（无独立新 capability；本变更扩展现有 sim-user-admin 管理能力。）

### Modified Capabilities

- `sim-user-admin`：新增 runtime 只读 API 与 UI 展示要求；明确周期间隔可只读展示、不可编辑；任务状态须结构化展示。

## Impact

- `api/v1/sim_admin_http.go` — 新增 runtime DTO 与路由定义
- `internal/controller/sim_admin_api.go` — `RuntimeGet` handler
- `internal/services/simuser/` — 组装 runtime 快照（复用 `LoadRuntimeFlags`、`countSimUsers`、`dbcfg.DatabaseNameFromLink`）
- `resource/public/sim-admin.html` — 运行配置面板与状态表格
- `gateway-app` 反代路径不变（`/sim/admin/api/*` 已存在）
- 无新 Redis、无新后台任务、无 usage 统计变更（Admin 只读接口）
