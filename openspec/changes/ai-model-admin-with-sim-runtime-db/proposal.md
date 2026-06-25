## Why

LLM lane 的模型与并发/等待队列配置分散在 voice-admin、ucg-admin 与 sim 的 env 变量中；sim 另有大量 `SIM_TASK_*` / `SIM_INTERVAL_*` / `SIM_LLM_*` env，运维需在多个页面与 `.env` 间切换，且 sim LLM 与任务周期无法在线热更新。需要**统一 AI 模型 Admin 入口**，并将 sim 运行时与 LLM lane 配置迁入 DB，显著精简 Compose env。

## What Changes

- **新增** `resource/public/ai-model-admin.html`：一屏管理 7 条 LLM lane（voice×2、ucg polish、sim×4）的 provider/model/maxInFlight/maxWaiters；Hub 导航新增「AI 模型与并发」。
- **voice-admin / ucg-admin / sim-admin**：**移除** LLM 并发编辑 UI，**仅保留**跳转链接至统一页；额度/Prompt/任务状态等职责不变。
- **sim-user-service**：
  - 新增 `sim_llm_lane_config` 表与 `GET/PUT /sim/admin/api/llm-lanes`（对齐 voice llm-lanes 契约）。
  - 扩展 `sim_config`（或等价表）持久化任务开关、周期、限速、E1 参数等原 env 运行时项。
  - 引入 **SchedulerManager**：调度类配置保存后 Stop→Start reload；**热重启跳过长错峰**；PUT 响应含 **生效语义**（立即 / 预计下一跑）。
  - `SimLLMLaneStore` 改为 DB > env 兜底 > 代码种子。
- **精简 env**：从 `docker-compose.microservices.yml` 与 `.env.example` 移除 sim 的 `SIM_LLM_*` 及可 DB 化的 `SIM_TASK_*` / `SIM_INTERVAL_*` 等；保留 `SIM_DB_LINK`、`SIM_USER_SERVICE_ENABLED`、`SIM_ADMIN_PASSWORD` 及 API Key。
- **aimodel allowlist**：扩展 zhipu 生图/生视频 model（`cogview-3-flash`、`cogvideox-flash` 等）供 sim lane Admin 校验。

## Capabilities

### New Capabilities

- `ai-model-admin-ui`：gateway-app 统一 AI 模型与并发 Admin 页及 Hub 导航。
- `sim-llm-lane-admin`：sim 四 lane DB 存储与 Admin GET/PUT API。
- `sim-runtime-config`：sim 运行时配置 DB 化、SchedulerManager reload 与保存生效响应。

### Modified Capabilities

- `voice-admin-ui`：移除 LLM 车道 Tab/表单，改为链接统一页；保留 AI 额度。
- `llm-lane-admin`：LLM lane 运维 UI 主入口迁至统一页（voice/ucg API 不变）。
- `sim-user-admin`：运行配置由 env 只读改为 DB 可编辑；runtime API 与保存结果语义更新。

## Impact

- **代码**：`internal/services/simuser/**`（schema、lane_store、runtime、scheduler）、`internal/controller/sim_admin_api.go`、`api/v1/sim_admin_http.go`、新增 `sim_llm_lane*.go`；`internal/services/aimodel/profile.go`（allowlist）；`resource/public/ai-model-admin.html`、`admin-modules.js`、`admin_static_pages.go`、`voice-admin.html`、`ucg-admin.html`、`sim-admin.html`。
- **配置**：`manifest/docker/docker-compose.microservices.yml`、`manifest/docker/.env.example`；`docs/runbooks/release-deploy-and-run.md`（sim 部署与 Admin 说明）。
- **数据库**：sim 库新增/扩展表；EnsureSchema 种子与 env 导入（`updated_by=seed`）。
- **服务**：sim-user-service（主变更）、gateway-app（静态页）；voice/ucg **无** LLM 运行时变更。
- **usage 统计**：新增 Admin 静态页与 sim/voice/ucg Admin PUT **不计入** App usage（运维型）。
- **非目标**：不新建 ai-service；不迁移 voice 流式 LLM 执行路径；不在 Admin 配置 API Key。
