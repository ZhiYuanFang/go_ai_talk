## 1. Sim 数据库与 LLM Lane 后端

- [x] 1.1 在 `internal/services/simuser/schema.go` 新增 `sim_llm_lane_config` 表及 EnsureSchema 默认种子
- [x] 1.2 扩展 `sim_config`（或等价）持久化 taskSwitches、intervals、rateLimit、ephemeral 字段
- [x] 1.3 实现 `SimLLMLaneStore` DB 优先加载（DB > env > seed）及 EnsureDefaultRows
- [x] 1.4 新增 `GET/PUT /sim/admin/api/llm-lanes`（controller + api 类型 + 路由注册）
- [x] 1.5 PUT llm-lanes 后调用 `aimodel.InvalidateLaneCache()`，不触发 scheduler reload
- [x] 1.6 扩展 `aimodel.ProviderModels` 含 `cogview-3-flash`、`cogvideox-flash`

## 2. Sim 运行时与 SchedulerManager

- [x] 2.1 实现 `LoadRuntimeFromDB` 组装 `RuntimeFlags`（DB 优先，env 兜底）
- [x] 2.2 实现 `SchedulerManager`（cancel ctx、WaitGroup、Reload、热重启 skipStagger）
- [x] 2.3 改造 `StartScheduler` 使用 SchedulerManager 可 cancel context
- [x] 2.4 扩展 `PUT /sim/admin/api/config` body/持久化及 diff 判断是否 Reload
- [x] 2.5 PUT config 响应增加 `scheduleReloaded`、`effects[]`、`taskSchedule[]`
- [x] 2.6 调整 `GET /sim/admin/api/runtime` 读 DB 生效值

## 3. 统一 AI 模型 Admin 前端

- [x] 3.1 新增 `resource/public/ai-model-admin.html`（七 lane 加载/保存/分域错误提示）
- [x] 3.2 `admin-modules.js` 增加 ai-model-admin Hub 入口
- [x] 3.3 `admin_static_pages.go` 注册 ai-model-admin（及补齐 voice-admin 若缺失）
- [x] 3.4 `voice-admin.html` 移除 LLM Tab，增加链至统一页
- [x] 3.5 `ucg-admin.html` 移除 polish 并发/模型编辑，增加链至统一页
- [x] 3.6 `sim-admin.html` 移除 LLM 编辑，增加链至统一页

## 4. Sim Admin 运行时 UI

- [x] 4.1 sim-admin 运行配置区改为可编辑表单（task/interval/rateLimit/ephemeral）
- [x] 4.2 保存后渲染 `effects` 与 `taskSchedule`（立即生效 vs 下一跑提示）
- [x] 4.3 区分 `serviceEnabled`（env recreate）与 `dbEnabled`（在线保存）说明文案

## 5. 配置与文档

- [x] 5.1 精简 `manifest/docker/docker-compose.microservices.yml` sim env 块
- [x] 5.2 更新 `manifest/docker/.env.example` sim 段注释与保留项
- [x] 5.3 更新 `docs/runbooks/release-deploy-and-run.md`（sim DB 迁移、Admin 使用、env 变更）
- [x] 5.4 精简 `manifest/docker/env/.env.test` 与 `.env.prod` 中 SIM_TASK_* / SIM_INTERVAL_* / SIM_LLM_* 等
- [x] 5.5 精简 voice/ucg 的 `*_MAX_INFLIGHT` / `*_MAX_WAITERS` / `*_TIMEOUT_SEC`（compose、.env.example、.env.test/.prod）
- [x] 5.6 精简 `manifest/docker/env/.env.test` 与 `.env.prod` 冗余注释与空行

## 6. 验证

- [x] 6.1 本地/测试：ai-model-admin 七 lane GET/PUT 联通
- [x] 6.2 测试：sim config PUT 触发 reload 且 `nextRunHint` 合理
- [x] 6.3 测试：仅 maxSimUsers 变更可不 reload
- [x] 6.4 测试：llm-lanes PUT 不 reload scheduler
- [x] 6.5 运行 `openspec validate ai-model-admin-with-sim-runtime-db --strict`
