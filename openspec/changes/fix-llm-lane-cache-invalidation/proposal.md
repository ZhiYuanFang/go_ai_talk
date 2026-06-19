## Why

`llm-lane-gate-admin` 交付后，Admin **PUT** `/ucg/admin/api/ai-config` 与 `/voice/admin/api/llm-lanes` 在保存成功后会调用 `InvalidateLaneCache()` 与 `ProfileStore.InvalidateCache()`，二者互相回调形成**无限递归**，触发 Go **stack overflow**，导致 **ucg-service / voice-service 进程退出**（测试环境已观测容器由运行中变为已停止）。同时 `voice-admin.html` 缺少 `api()` 与 `AdminCommon.requireAdmin()`，页面仅显示标题；UCG outbox worker 在队列为空时将 `sql.ErrNoRows` 误记为 WARN。需在不影响 lane 闸门语义的前提下修复稳定性与运维页可用性。

## What Changes

- 打破 `aimodel.InvalidateLaneCache()` 与 `UcgPolishProfileStore` / `VoiceLLMLaneStore` 的 `InvalidateCache()` 互相调用；Admin PUT 后缓存失效语义保持不变（进程内 profile 短 TTL + store 本地 cache 清空）。
- 修正 `voice-admin.html`：对齐 `ucg-admin.html` 的 `AdminCommon.adminFetch` 与 `requireAdmin()` 初始化模式。
- 修正 `StartChatPersistWorker` / `StartAuditPublishRelayWorker` 轮询：`Scan` 无行时（`sql.ErrNoRows`）视为空闲 tick，**MUST NOT** 打 WARN。
- runbook 补充：`ucg_ai_config` ALTER 迁移与 ai-config PUT 验收步骤（运维，非代码行为变更）。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `llm-lane-admin`：明确 Admin PUT 更新 lane profile 后失效缓存**不得**导致进程崩溃；`InvalidateLaneCache` 与 store 本地失效的调用边界。
- `voice-admin-ui`：`voice-admin.html` **MUST** 在 Hub 登录后展示 AI 额度与 LLM 车道表单（与 `ucg-admin.html` 同类 AdminCommon 模式）。

## Impact

- **代码**：`internal/services/aimodel/profile.go`；`internal/services/ucg/polish_lane_store.go`、`ai_config.go`；`internal/services/voice/llm_lane_store.go`；`internal/services/ucg/chat_persist_worker.go`、`audit_publish_relay_worker.go`；`resource/public/voice-admin.html`。
- **API**：无契约变更；PUT ai-config / llm-lanes 由「可能打挂进程」恢复为正常 200。
- **部署**：测试/生产须在 `ai_voice_ucg` 执行既有 runbook `ucg_ai_config` ALTER（与 llm-lane-gate-admin 相同），否则 PUT 仍可能 SQL 失败（但不因本变更引入）。
- **App API 使用统计**：无新增接口；不计入 usage 统计。
