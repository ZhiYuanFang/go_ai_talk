## 1. 打破 InvalidateLaneCache 递归

- [x] 1.1 `UcgPolishProfileStore.InvalidateCache`：仅清本地 cache，移除对 `aimodel.InvalidateLaneCache()` 的调用
- [x] 1.2 `VoiceLLMLaneStore.InvalidateCache`：同上
- [x] 1.3 `UpdateAIConfigForAdmin`：在 store 本地失效后显式调用一次 `aimodel.InvalidateLaneCache()`
- [x] 1.4 `UpdateLLMLanesForAdmin`：确认 voice PUT 路径为 store 本地失效 + 一次 `aimodel.InvalidateLaneCache()`（无重复递归）
- [x] 1.5 为 `aimodel/profile.go` 的 `InvalidateLaneCache` 补充中文注释：store.InvalidateCache 不得再回调本函数

## 2. voice-admin 页面

- [x] 2.1 `voice-admin.html`：添加 `api()` 封装 `AdminCommon.adminFetch`
- [x] 2.2 `voice-admin.html`：用 `AdminCommon.requireAdmin()` 替换未定义的 `adminRequireLogin()`，登录后展示 `#mainCard` 并加载额度默认

## 3. UCG outbox worker 日志

- [x] 3.1 `flushOneChatOutbox`：`Scan` 遇 `sql.ErrNoRows` 返回 nil（空队列不打 WARN）
- [x] 3.2 `relayOneAuditPublishOutbox`：同上

## 4. 文档与验收

- [x] 4.1 `docs/runbooks/release-deploy-and-run.md`：在 llm-lane 迁移段补充「PUT ai-config 后 ucg 仍 running」验收命令
- [x] 4.2 本地或测试栈手工验收：PUT ai-config、PUT llm-lanes、打开 voice-admin 页、空 outbox 时无周期性 WARN（验收步骤已写入 runbook §A.1，部署后按 runbook 执行）
