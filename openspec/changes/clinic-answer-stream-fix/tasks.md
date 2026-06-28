## 1. aimodel 流式路由

- [x] 1.1 `invokeStreamHTTP`：`sawReasoning` + `ThinkingEnabled` 时将 content 路由到 answer
- [x] 1.2 `provider.go` 注释说明分片语义

## 2. 验证

- [x] 2.1 `go build ./...` 通过
- [x] 2.2 `openspec validate clinic-answer-stream-fix --strict` 通过
