## Context

`llm-lane-gate-admin` 引入 `aimodel.ProfileStore` 与 `InvalidateLaneCache()`：Admin PUT 后需清空进程内 profile 短 TTL 缓存与 store 持有的本地 cache。当前实现中：

- `aimodel.InvalidateLaneCache()` 在清空 `profileCache` 后调用 `defaultStore.InvalidateCache()`
- `UcgPolishProfileStore.InvalidateCache()` / `VoiceLLMLaneStore.InvalidateCache()` 清空本地 cache 后又调用 `aimodel.InvalidateLaneCache()`

形成互相递归，PUT 成功路径上触发 **stack overflow**，进程退出。

另：`voice-admin.html` 使用未定义的 `adminRequireLogin()` 与 `api()`，主面板永不展示；outbox worker 将空队列 `Scan` 的 `sql.ErrNoRows` 记为 WARN。

## Goals / Non-Goals

**Goals:**

- Admin PUT `/ucg/admin/api/ai-config` 与 `/voice/admin/api/llm-lanes` 成功后进程 **MUST** 保持运行；缓存失效后下一笔 LLM 调用 **MUST** 读到新 profile。
- `voice-admin.html` Hub 登录后 **MUST** 展示 AI 额度与 LLM 车道 UI。
- outbox 轮询无待处理行时 **MUST NOT** 产生 `[ucg-chat-persist]` / `[ucg-audit-outbox]` WARN。

**Non-Goals:**

- 不改 lane 闸门 Redis 语义、50301、allowlist 或 Admin API JSON 契约。
- 不新增测试文件（仓库当前阶段约定）。
- 不在此变更内实现 `ucg_ai_config` 自动 migration（仍依赖 runbook 手工 ALTER）。

## Decisions

### D1：单向 cache 失效（采用）

**决策**：`ProfileStore.InvalidateCache()` **仅**清空 store 本地 cache，**不再**调用 `aimodel.InvalidateLaneCache()`。Admin PUT 路径统一调用：

1. 域内 `InvalidateAIConfigCache()`（ucg）或等价
2. `store.InvalidateCache()`（本地）
3. `aimodel.InvalidateLaneCache()` **一次**（清空 `profileCache`；**不再**回调 store）

**理由**：与 `llm-lane-gate-admin` design「PUT 后 InvalidateLaneCache」一致，且调用链有明确顶层入口。

**备选（未采用）**：在 `InvalidateLaneCache` 内不调 store——则 store 本地 cache 可能 stale 至 TTL；需额外约定 store 无本地 cache 或 TTL=0。

### D2：voice-admin 对齐 ucg-admin 的 AdminCommon 模式（采用）

**决策**：在 `voice-admin.html` 内定义 `function api(...){ return AdminCommon.adminFetch(...); }`；初始化使用 `if (AdminCommon.requireAdmin()) { ... }`，与 `ucg-admin.html` 一致。

**理由**：`admin-common.js` 已提供完整能力；避免引入新的全局 helper 名。

### D3：outbox 空队列视为成功 tick（采用）

**决策**：`flushOneChatOutbox` / `relayOneAuditPublishOutbox` 在 `Scan` 返回 `sql.ErrNoRows` 时 `return nil`；与 `media_register.go` 等同族处理方式。

**理由**：无 pending 行是常态，不应 WARN。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 只调 `InvalidateLaneCache` 而不调 store 本地失效 | PUT 路径显式两步：store + aimodel |
| voice PUT 仍调两次顶层失效 | 删除 store 内递归后，保留 `store.InvalidateCache()` + `InvalidateLaneCache()` 各一次即可 |
| 测试库缺 `ucg_ai_config` 新列 PUT 仍 SQL 失败 | runbook 验收步骤写入 tasks；与崩溃无关 |

## Migration Plan

1. 合并并部署 ucg-service / voice-service 新镜像；**gateway-app** 仅当 `voice-admin.html` 变更时需 redeploy（静态资源挂载或重建 gateway-app 镜像，按现有 test 栈惯例）。
2. 测试环境：确认 `ai_voice_ucg` 已执行 `ucg_ai_config` ALTER（runbook 既有 SQL）。
3. 验收：PUT ai-config 后 `docker ps` ucg 仍为 running；voice-admin 页展示表单；ucg 日志无周期性 outbox WARN（空队列时）。

## Open Questions

（无）
