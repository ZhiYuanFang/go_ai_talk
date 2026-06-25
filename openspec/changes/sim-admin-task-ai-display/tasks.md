## 1. 后端映射与 API

- [x] 1.1 新增 `internal/services/simuser/task_llm.go`：`taskLLMUsage` 与 `BuildTaskAIModelCatalog`
- [x] 1.2 扩展 `api/v1/sim_admin_http.go`：`SimAdminTaskAIModelDTO`、`taskAiModels` 字段
- [x] 1.3 `StatusGet` 填充 `taskAiModels`（`LoadAllLaneProfiles` + catalog）
- [x] 1.4 `RuntimeGet` / `GetRuntimeSnapshot` 同步返回 `taskAiModels`

## 2. sim-admin 前端

- [x] 2.1 任务状态表增加「AI 模型」列与 `fmtAiModels` 渲染
- [x] 2.2 补充说明文案与 ai-model-admin 链接

## 3. 验证

- [x] 3.1 本地编译 `go build ./...`
- [ ] 3.2 测试环境：sim-admin 各任务 AI 列与 ai-model-admin 配置一致
- [x] 3.3 运行 `openspec validate sim-admin-task-ai-display --strict`
