## 1. aimodel 契约与 adapter

- [x] 1.1 在 `internal/services/aimodel/request.go` 为 `ThinkingEnabled`、`MaxTokens` 补充中文注释（默认 disabled、MaxTokens 与 reasoning 共预算）
- [x] 1.2 修改 `provider_zhipu.go`：`ApplyThinkingOptions` 始终写入 `thinking.type`（false→disabled，true→enabled）
- [x] 1.3 审阅 `provider_deepseek.go`：确认 `ThinkingEnabled=false` 时不写入 thinking 启用字段；必要时补充中文注释

## 2. 调用方与回归审阅

- [x] 2.1 确认 `internal/services/voice/clinic_llm.go` 仍设 `ThinkingEnabled: true`（clinic opt-in）
- [x] 2.2 grep 全仓 `aimodel.ChatRequest` / `Invoke(` 调用点，确认无遗漏需 opt-in thinking 的路径
- [x] 2.3 手动或 test 环境验证：sim 昵称任务（LaneSimText）上游请求含 `thinking.disabled` 且响应 `content` 非空

## 3. 评审与归档

- [x] 3.1 评审 checklist：对照 `specs/aimodel-thinking-mode`、`llm-lane-gate`、`pangbao-ai-clinic` delta 验收
- [x] 3.2 确认无新增 App HTTP 路由、无 DB/Redis/env 变更
