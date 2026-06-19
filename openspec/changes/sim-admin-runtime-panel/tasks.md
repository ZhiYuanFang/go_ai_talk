## 1. API 与 DTO

- [x] 1.1 在 `api/v1/sim_admin_http.go` 新增 `SimAdminRuntimeDTO`、`SimAdminRuntimeGetReq/Res` 及 `GET /sim/admin/api/runtime` 路由元数据
- [x] 1.2 在 `internal/services/simuser` 实现 `GetRuntimeSnapshot(ctx)`：组装 `LoadRuntimeFlags`、库名、`GetConfig`、`countSimUsers`、限速 env；不返回密码类字段
- [x] 1.3 在 `internal/controller/sim_admin_api.go` 实现 `RuntimeGet`，鉴权与现有 admin handler 一致
- [x] 1.4 在 `sim-user-service` 路由注册处绑定 `RuntimeGet`（与现有 `ConfigGet`/`StatusGet` 同组）

## 2. 管理页 UI

- [x] 2.1 在 `sim-admin.html` 新增「运行配置（只读）」面板：双层开关、库名、用户数/上限、任务开关、周期、错峰、限速、E1 参数
- [x] 2.2 页内增加 env 变更须 `force-recreate sim-user-service` 及首次 tick 延迟说明
- [x] 2.3 将任务状态由纯 `JSON.stringify` 改为表格展示（任务名、上次运行、成功/失败、最近错误、pending video）
- [x] 2.4 刷新按钮同时拉取 runtime + status；runtime API 失败时降级提示而非阻断整页

## 3. 验收

- [x] 3.1 本地或测试环境：修改 `.env` 中 `SIM_INTERVAL_REGISTER` 后 `force-recreate`，确认管理页展示值与容器 `printenv` 一致
- [x] 3.2 确认响应 body 不含 DSN 密码、`defaultPassword` 等敏感字段
- [x] 3.3 确认页面无周期/任务开关编辑控件，仅只读展示
