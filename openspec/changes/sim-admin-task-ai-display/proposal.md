## Why

`ai-model-admin-with-sim-runtime-db` 将 sim 四条 LLM lane 迁入 DB 与统一 Admin 页后，运维在 sim-admin 调整任务周期/开关时，仍无法一眼看出「该任务实际会调用哪条 lane、当前生效的 provider/model 是什么」。需要在任务状态表中展示 AI 使用情况，减少跨页对照与误配风险。

## What Changes

- 后端在 `simuser` 维护调度任务名 → LLM lane 映射（与 `tasks.go` 调用一致），解析当前生效 profile。
- `GET /sim/admin/api/status` 与 `GET /sim/admin/api/runtime` 响应增加 `taskAiModels`（按任务名索引）。
- `sim-admin.html` 任务状态表新增「AI 模型」列，展示 lane、用途说明、provider/model；T6 关注无 LLM 显示「—」。
- 页面文案链至 ai-model-admin 修改模型；本变更不提供 lane 编辑。

## Capabilities

### New Capabilities

（无独立新 capability；增量并入 sim-user-admin。）

### Modified Capabilities

- `sim-user-admin`：status/runtime API 与任务状态 UI 增加 per-task AI 展示要求。

## Impact

- **服务**：`sim-user-service`（`internal/services/simuser`、`internal/controller/sim_admin_api.go`）
- **API**：`api/v1/sim_admin_http.go` — `SimAdminStatusDTO`、`SimAdminRuntimeDTO` 扩展
- **前端**：`resource/public/sim-admin.html`
- **依赖**：`SimLLMLaneStore` / `LoadAllLaneProfiles()`；无新表、无 Redis、无新后台任务
