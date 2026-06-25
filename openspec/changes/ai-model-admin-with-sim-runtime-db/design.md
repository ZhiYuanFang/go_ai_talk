## Context

- `llm-lane-gate-admin` 已在 voice/ucg 实现 `aimodel` + Redis 闸门 + 分域 Admin API；UI 分散在 voice-admin / ucg-admin。
- sim 四 lane 仅 env + `SimLLMLaneStore`；任务开关/周期由 `LoadRuntimeFlags` 读 env，scheduler 在进程启动时一次性 `StartScheduler`（ctx 不可 cancel）。
- `sim-admin-runtime-panel` 将 env 运行配置设计为**只读**展示；本变更 overturn 该 Non-Goal，改为 DB 可编辑 + 保存时 reload。
- 闸门按 **model** 全局共享（Redis）；统一 Admin 页需提示同 model 多 lane 共用池。
- 项目约束：sim 进程使用 `SIM_DB_LINK`；禁止新增无 OpenSpec 批准的 ticker 扫表；reload 为 Admin PUT 触发的同步操作，非后台 reconciler。

## Goals / Non-Goals

**Goals:**

- 单一 Admin 页管理 7 lane 的 provider/model/maxInFlight/maxWaiters。
- sim LLM 四 lane DB 持久化 + Admin 热更新（`InvalidateLaneCache`）。
- sim 任务开关、周期、限速、E1 等迁入 DB；保存后 SchedulerManager reload，避免 tick 内频繁读 DB。
- PUT 响应返回 `effects` / `taskSchedule` 等字段，前端展示「立即生效 vs 预计下一跑」。
- 热重启 reload **跳过** `startupStaggerMax` 长错峰（或仅用 0～30s 短延迟）。
- `.env` sim 段收敛为基础设施 + 进程总闸 + Admin 口令。

**Non-Goals:**

- 不实现 ai-service / gateway 承载 LLM 执行。
- 不保证 save 时**强杀**进行中的 LLM/HTTP/E1 聊天 goroutine（文档说明「进行中任务可能跑完」）。
- 不在 Admin 编辑 API Key。
- 不迁移百度 ASR/TTS 并发至本页。
- 不新增 `*_test.go`。

## Decisions

### D1：统一页 `ai-model-admin.html`（gateway-app 静态）

- 路径：`/device/admin/ai-model-admin.html`；`admin-modules.js` 增加 `id: ai-model-admin`，`showInNav: true`。
- 加载：并行 GET voice llm-lanes、ucg ai-config（polish 字段）、sim llm-lanes。
- 保存：`Promise.all` PUT 三块；部分失败分域提示。
- voice/ucg/sim 原页 **删除** LLM 编辑控件，保留一句链接：「模型与并发 → AI模型与并发」。

**备选**：保留 voice-admin LLM Tab 双写 — 否决（用户明确要求不保留）。

### D2：sim `sim_llm_lane_config` 表

- 结构对齐 voice `llm_lane_config`：`lane` PK（`simText`|`simVision`|`simImageGen`|`simVideoGen`）、`provider`、`model`、`max_in_flight`、`max_waiters`、`timeout_sec`、`updated_at`、`updated_by`。
- `SimLLMLaneStore.Load`：DB > env（`ProfileFromEnv`，仅 seed 行）> `DefaultSeedProfile`。
- `GET/PUT /sim/admin/api/llm-lanes`：鉴权同现有 sim-admin（`X-Admin-Password`）；PUT 后 `InvalidateLaneCache()`。
- EnsureDefaultRows：无行时写代码种子；若 env 有 `SIM_LLM_*` 且行为 seed，可 import 一次。

### D3：sim 运行时配置表

- 扩展 `sim_config` 单行（`id=1`）JSON 或列式字段，包含：
  - 已有：`enabled`、`max_sim_users`
  - 新增：`task_register`、`task_comment`、`task_post_image`、`task_post_video`、`task_chat`、`task_follow`、`video_poll`（bool）
  - 新增：`interval_*`（秒或 duration 字符串）、`startup_stagger_sec`、`ephemeral_chat_loop_sec`、`ephemeral_chat_window_sec`、`ucg_rate_limit_rps`、`ucg_rate_limit_burst`
- `LoadRuntimeFromDB(ctx)` 组装 `RuntimeFlags`；`LoadRuntimeFlags` 改为 **DB 优先**，env 仅当 DB 列缺失/迁移期兜底。
- 进程级 **`SIM_USER_SERVICE_ENABLED=false`** 仍阻止 scheduler 启动（硬闸，保留 env）。

### D4：SchedulerManager reload

```text
PUT /sim/admin/api/config (扩展 body)
  → 写 DB
  → diff 变更类型
  → 若含调度类字段 OR enabled 变化：SchedulerManager.Reload(ctx)
       Stop: cancel(schedulerCtx); WaitGroup 等待 goroutine 在 timer 处退出
       Start: LoadRuntimeFromDB; 按开关启动 goroutine（热重启 skipStagger=true）
  → 若仅 maxSimUsers 变更：可不 Reload（注册任务 tick 内已 GetConfig）
  → 构造 ConfigPutRes：scheduleReloaded, effects[], taskSchedule[]
```

- `runPeriodic` 继续使用可 cancel 的 parent ctx；**不在**本变更内改为 tick 内读 DB。
- `spawnEphemeralChat`  detached goroutine：reload 不强制终止；`effects` 含 `ephemeral_may_continue` 提示。

### D5：保存生效语义（API 响应）

- `effects[]`：`kind` + 可选 `task`/`interval`/`etaSec`/`message`。
- `taskSchedule[]`：每任务 `enabled`、`interval`、`lastRunAt`（来自 `sim_task_run`）、`nextRunHint`（估算：`lastRunAt + interval`，无历史则「重启后首轮约 interval ± jitter」）。
- 前端「保存结果」面板渲染；**不**二次弹窗「是否立即生效」（保存即生效意图）。

### D6：aimodel allowlist 扩展

- `ProviderModels[ProviderZhipu]` 增加 `cogview-3-flash`、`cogvideox-flash`（及 normalize 大小写）供 simImageGen/simVideoGen Admin PUT 校验。

### D7：env 精简

- 从 compose 删除：`SIM_LLM_*`（20 项）、`SIM_TASK_*`、`SIM_INTERVAL_*`、`SIM_EPHEMERAL_*`、`SIM_STARTUP_STAGGER_MAX`、`SIM_UCG_RATE_LIMIT_*`、`SIM_VIDEO_POLL_ENABLED`。
- 保留：`SIM_DB_LINK`、`SIM_USER_SERVICE_ADDR`、`SIM_USER_SERVICE_ENABLED`、`SIM_ADMIN_PASSWORD`；gateway 侧 `SIM_SERVICE_BASE_URL` 不变。
- `.env.example` 注释说明：运行时与 LLM 默认值在 DB seed，经 Admin 修改。

### D8：runtime GET 调整

- `GET /sim/admin/api/runtime` 改为读 **DB 生效值** + `serviceEnabled`（env）；不再强调「改 env 须 recreate」作为主路径（recreate 仅进程总闸或部署）。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| Reload 时 tick 进行中 | 文档 + effects 说明；下一周期用新配置 |
| E1 聊天 goroutine 泄漏旧 flags | 新 spawn 用新 flags；旧窗口自然结束 |
| 每次保存全量 Reload | 仅 `maxSimUsers` 时跳过 Reload |
| 迁移期 DB 空、env 已删 | EnsureSchema seed + 可选 env 导入 |
| cogview 不在 allowlist | D6 扩展 |
| InvalidateLaneCache 递归 | 沿用 fix-llm-lane-cache-invalidation 约定，store 不得回调 InvalidateLaneCache |

## Migration Plan

1. 部署 sim-user-service（EnsureSchema 建表/扩列 + 种子；可选从 env 导入 seed 行）。
2. 部署 gateway-app（新静态页 + 三页 LLM 链接化）。
3. 更新 test/prod `.env`：删除已 DB 化 sim 变量；**不** require recreate 即可用 Admin 改 LLM/周期。
4. 回滚：旧镜像 + 恢复 env 块；DB 新列可保留。

## Open Questions

（无阻塞项；实现按本设计进行。）
