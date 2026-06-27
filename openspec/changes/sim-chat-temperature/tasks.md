## 1. aimodel 平台

- [x] 1.1 `request.go`：`ChatRequest.Temperature *float64` 及中文注释
- [x] 1.2 `client.go` `buildRequestBody`：非 nil 时写入 `temperature`

## 2. sim 任务常量

- [x] 2.1 新增 `task_llm_temp.go`：5 个任务常量 + `simChatRequest` helper
- [x] 2.2 `tasks.go`：5 处 `aimodel.Invoke` 改用 helper 与对应常量

## 3. 校验

- [x] 3.1 `go build` simuser + aimodel 通过
- [x] 3.2 `openspec validate sim-chat-temperature` 通过
